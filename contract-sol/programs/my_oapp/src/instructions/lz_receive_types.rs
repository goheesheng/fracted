use crate::*;
use oapp::endpoint_cpi::{get_accounts_for_clear, LzAccount};
use oapp::{endpoint::ID as ENDPOINT_ID, LzReceiveParams};
use std::str::FromStr;

// Must match lz_receive.rs
const TRANSFER_PROGRAM_ID_STR: &str = "GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1";
const SPL_TOKEN_PROGRAM_ID_STR: &str = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA";
// Config PDA seed for the transfer contract (must match the callee contract)
const TRANSFER_CONFIG_SEED: &[u8] = b"config";
const TRANSFER_VAULT_SEED: &[u8] = b"vault";

/// `lz_receive_types` is queried off-chain by the Executor before calling
/// `lz_receive`. It must return **every** account that will be touched by the
/// actual `lz_receive` instruction as well as the accounts required by
/// `Endpoint::clear`.
///
/// The return order must match exactly what `lz_receive` expects or the
/// cross-program invocation will fail.
#[derive(Accounts)]
pub struct LzReceiveTypes<'info> {
    #[account(seeds = [STORE_SEED], bump = store.bump)]
    pub store: Account<'info, Store>,
}

impl LzReceiveTypes<'_> {
    pub fn apply(
        ctx: &Context<LzReceiveTypes>,
        params: &LzReceiveParams,
    ) -> Result<Vec<LzAccount>> {
        // 1. The store PDA is always the first account and is mutable.  If your
        // program derives the store PDA with additional seeds, ensure the same
        // seeds are used when providing the store account.
        let store = ctx.accounts.store.key();

        // 2. The peer PDA for the remote chain needs to be retrieved, for later verification of the `params.sender`.
        let peer_seeds = [PEER_SEED, &store.to_bytes(), &params.src_eid.to_be_bytes()];
        let (peer, _) = Pubkey::find_program_address(&peer_seeds, ctx.program_id);

        // Accounts used directly by `lz_receive`
        let mut accounts = vec![
            // store (mutable)
            LzAccount { pubkey: store, is_signer: false, is_writable: true },
            // peer (read-only)
            LzAccount { pubkey: peer, is_signer: false, is_writable: false }
        ];

        // If message is 128 bytes, add transfer accounts FIRST (these go into Option fields)
        if params.message.len() == 128 {
            let tag = params.message[31];
            if tag == 101 {
                // Decode merchant from message (bytes 64..96)
                let merchant_bytes: [u8; 32] = params.message[64..96].try_into().unwrap();
                let merchant = Pubkey::new_from_array(merchant_bytes);
                
                // Decode dst_token from message (bytes 32..64)
                let dst_token_bytes: [u8; 32] = params.message[32..64].try_into().unwrap();
                let dst_token = Pubkey::new_from_array(dst_token_bytes);

                // Transfer program
                let transfer_program = Pubkey::from_str(TRANSFER_PROGRAM_ID_STR).unwrap();
                
                // Transfer config PDA
                let (transfer_config, _) = Pubkey::find_program_address(
                    &[TRANSFER_CONFIG_SEED],
                    &transfer_program
                );
                
                // Vault authority PDA
                let (vault_authority, _) = Pubkey::find_program_address(
                    &[TRANSFER_VAULT_SEED, transfer_config.as_ref()],
                    &transfer_program
                );
                
                // Derive associated token accounts manually
                let associated_token_program = Pubkey::from_str("ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL").unwrap();
                let spl_token_program = Pubkey::from_str(SPL_TOKEN_PROGRAM_ID_STR).unwrap();
                
                let vault_token_account = Self::derive_ata(&vault_authority, &dst_token, &spl_token_program, &associated_token_program);
                let recipient_token_account = Self::derive_ata(&merchant, &dst_token, &spl_token_program, &associated_token_program);

                // Add transfer-related accounts (must match order in lz_receive.rs)
                accounts.push(LzAccount { pubkey: transfer_program, is_signer: false, is_writable: false });
                accounts.push(LzAccount { pubkey: transfer_config, is_signer: false, is_writable: false });
                accounts.push(LzAccount { pubkey: vault_authority, is_signer: false, is_writable: false });
                accounts.push(LzAccount { pubkey: vault_token_account, is_signer: false, is_writable: true });
                accounts.push(LzAccount { pubkey: recipient_token_account, is_signer: false, is_writable: true });
                accounts.push(LzAccount { pubkey: dst_token, is_signer: false, is_writable: false }); // mint
                accounts.push(LzAccount { pubkey: spl_token_program, is_signer: false, is_writable: false });
            }
        }

        // Append the additional accounts required for `Endpoint::clear` AFTER transfer accounts
        let accounts_for_clear = get_accounts_for_clear(
            ENDPOINT_ID,
            &store,
            params.src_eid,
            &params.sender,
            params.nonce,
        );
        accounts.extend(accounts_for_clear);

        Ok(accounts)
    }

    /// Manually derive associated token account address
    /// ATA = PDA([wallet, token_program, mint], associated_token_program)
    fn derive_ata(
        wallet: &Pubkey,
        mint: &Pubkey,
        token_program: &Pubkey,
        associated_token_program: &Pubkey,
    ) -> Pubkey {
        let seeds = &[
            wallet.as_ref(),
            token_program.as_ref(),
            mint.as_ref(),
        ];
        Pubkey::find_program_address(seeds, associated_token_program).0
    }
}
