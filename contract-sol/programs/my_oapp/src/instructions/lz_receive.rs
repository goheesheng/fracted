use crate::*;
use anchor_lang::prelude::*;
use std::str::FromStr;
use oapp::{
    endpoint::{
        cpi::accounts::Clear,
        instructions::ClearParams,
        ConstructCPIContext, ID as ENDPOINT_ID,
    },
    LzReceiveParams,
};

// Hardcoded callee program id for transfer_out CPI (parsed at runtime)
const TRANSFER_PROGRAM_ID_STR: &str = "GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1";
// Development flag: allow skipping Endpoint::clear when missing accounts (NOT SAFE FOR PRODUCTION)
const ALLOW_SKIP_CLEAR_ON_MISSING_ACCOUNTS: bool = true;

#[derive(Accounts)]
#[instruction(params: LzReceiveParams)]
pub struct LzReceive<'info> {
    /// OApp Store PDA.  This account represents the "address" of your OApp on
    /// Solana and can contain any state relevant to your application.
    /// Customize the fields in `Store` as needed.
    #[account(mut, seeds = [STORE_SEED], bump = store.bump)]
    pub store: Account<'info, Store>,
    /// Peer config PDA for the sending chain. 
    /// NOTE: constraint disabled for POC - MUST enable for production!
    #[account(
        seeds = [PEER_SEED, &store.key().to_bytes(), &params.src_eid.to_be_bytes()],
        bump = peer.bump
    )]
    pub peer: Account<'info, PeerConfig>,
    
    // Transfer contract related accounts (optional, only used if transfer_contract is set)
    /// CHECK: This is the transfer contract program
    pub transfer_program: Option<UncheckedAccount<'info>>,
    /// CHECK: This is the config account for the transfer contract
    pub transfer_config: Option<UncheckedAccount<'info>>,
    /// CHECK: This is the vault authority for the transfer contract
    pub vault_authority: Option<UncheckedAccount<'info>>,
    /// CHECK: Vault token account
    pub vault_token_account: Option<UncheckedAccount<'info>>,
    /// CHECK: Recipient token account
    pub recipient_token_account: Option<UncheckedAccount<'info>>,
    /// CHECK: Mint account
    pub mint: Option<UncheckedAccount<'info>>,
    /// CHECK: Token program
    pub token_program: Option<UncheckedAccount<'info>>,
}

impl LzReceive<'_> {
    pub fn apply(ctx: &mut Context<LzReceive>, params: &LzReceiveParams) -> Result<()> {
        // The OApp Store PDA is used to sign the CPI to the Endpoint program.
        let seeds: &[&[u8]] = &[STORE_SEED, &[ctx.accounts.store.bump]];

        // The first Clear::MIN_ACCOUNTS_LEN accounts were returned by
        // `lz_receive_types` and are required for Endpoint::clear
        msg!(
            "LzReceive: src_eid={}, nonce={}, msg_len={}, remaining_accounts_len={}",
            params.src_eid,
            params.nonce,
            params.message.len(),
            ctx.remaining_accounts.len()
        );
        if ctx.remaining_accounts.len() < Clear::MIN_ACCOUNTS_LEN {
            if ALLOW_SKIP_CLEAR_ON_MISSING_ACCOUNTS {
                msg!(
                    "LzReceive warning: missing Endpoint::clear accounts (have={}, need={}). Skipping clear (dev mode)",
                    ctx.remaining_accounts.len(),
                    Clear::MIN_ACCOUNTS_LEN
                );
            } else {
                msg!(
                    "LzReceive error: missing Endpoint::clear accounts. have={}, need={}",
                    ctx.remaining_accounts.len(),
                    Clear::MIN_ACCOUNTS_LEN
                );
                return Err(ErrorCode::MissingClearAccounts.into());
            }
        } else {
            let accounts_for_clear = &ctx.remaining_accounts[0..Clear::MIN_ACCOUNTS_LEN];
            // Call the Endpoint::clear CPI to clear the message from the Endpoint program.
            // This is necessary to ensure the message is processed only once and to
            // prevent replays.
            let _ = oapp::endpoint_cpi::clear(
                ENDPOINT_ID,
                ctx.accounts.store.key(),
                accounts_for_clear,
                seeds,
                ClearParams {
                    receiver: ctx.accounts.store.key(),
                    src_eid: params.src_eid,
                    sender: params.sender,
                    nonce: params.nonce,
                    guid: params.guid,
                    message: params.message.clone(),
                },
            )?;
        }

        // Process the message based on its format
        // Token payout path: ABI-encoded (uint8 tag, bytes32 dstToken, bytes32 merchant, uint256 netAmount)
        if params.message.len() == 128 {
            // Try to decode as token payout message
            if let Ok((tag, dst_token, merchant, net_amount)) = Self::decode_token_payout_message(&params.message) {
                if tag == 101 { // TAG_TOKEN_PAYOUT
                    if net_amount == 0 {
                        msg!("LzReceive error: net_amount is zero");
                        return Err(ErrorCode::ZeroNetAmount.into());
                    }
                    msg!(
                        "LzReceive payout: tag={}, net_amount={}, dst_token={}, merchant={}",
                        tag,
                        net_amount,
                        dst_token,
                        merchant
                    );
                    Self::call_transfer_out(
                        ctx,
                        dst_token,
                        merchant,
                        net_amount,
                    )?;
                    return Ok(());
                }
            }
        }

        // Legacy example: treat as string message
        let string_value = msg_codec::decode(&params.message)?;
        let store = &mut ctx.accounts.store;
        store.string = string_value;

        Ok(())
    }

    fn decode_token_payout_message(message: &[u8]) -> Result<(u8, Pubkey, Pubkey, u64)> {
        // ABI decode: (uint8 tag, bytes32 dstToken, bytes32 merchant, uint256 netAmount)
        // Layout: 4 x 32-byte words
        if message.len() != 128 {
            return Err(ErrorCode::InvalidMessageFormat.into());
        }

        // tag lives in the last byte of the first 32-byte word (right-aligned in abi.encode)
        let tag = message[31];

        // bytes32 values are raw 32 bytes
        let dst_token_bytes: [u8; 32] = message[32..64]
            .try_into()
            .map_err(|_| ErrorCode::InvalidAbiEncoding)?;
        let merchant_bytes: [u8; 32] = message[64..96]
            .try_into()
            .map_err(|_| ErrorCode::InvalidAbiEncoding)?;

        // uint256 is big-endian 32 bytes; downcast to u64 using the last 8 bytes
        let net_amount_be: [u8; 32] = message[96..128]
            .try_into()
            .map_err(|_| ErrorCode::InvalidAbiEncoding)?;
        // ensure high 24 bytes are zero to fit into u64
        if net_amount_be[..24].iter().any(|b| *b != 0) {
            msg!("LzReceive error: net_amount overflows u64");
            return Err(ErrorCode::AmountOverflow.into());
        }
        let net_amount_u64 = u64::from_be_bytes(net_amount_be[24..32].try_into().map_err(|_| ErrorCode::InvalidAbiEncoding)?);

        let dst_token = Pubkey::new_from_array(dst_token_bytes);
        let merchant = Pubkey::new_from_array(merchant_bytes);

        Ok((tag, dst_token, merchant, net_amount_u64))
    }

    fn call_transfer_out(
        ctx: &mut Context<LzReceive>,
        _dst_token: Pubkey,
        _merchant: Pubkey,
        amount: u64,
    ) -> Result<()> {
        // Check that all required accounts are provided
        let transfer_program = ctx
            .accounts
            .transfer_program
            .as_ref()
            .ok_or(ErrorCode::MissingTransferAccounts)?;
        let transfer_config = ctx.accounts.transfer_config.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let vault_authority = ctx.accounts.vault_authority.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let vault_token_account = ctx.accounts.vault_token_account.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let recipient_token_account = ctx.accounts.recipient_token_account.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let mint = ctx.accounts.mint.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let token_program = ctx.accounts.token_program.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;

        // Log accounts for debugging
        msg!(
            "transfer_out CPI: cfg={}, auth={}, vault_ta={}, recip_ta={}, mint={}, token_prog={}",
            transfer_config.key(),
            ctx.accounts.store.key(),
            vault_token_account.key(),
            recipient_token_account.key(),
            mint.key(),
            token_program.key()
        );

        // Create instruction data for transfer_out
        // transfer_out instruction: (discriminator: 8 bytes + amount: 8 bytes)
        let mut instruction_data = Vec::with_capacity(16);
        // Anchor discriminator: first 8 bytes of sha256("global:transfer_out")
        let disc = anchor_lang::solana_program::hash::hash(b"global:transfer_out").to_bytes();
        instruction_data.extend_from_slice(&disc[..8]);
        instruction_data.extend_from_slice(&amount.to_le_bytes());

        // Create accounts for the CPI call
        let accounts = vec![
            // config
            AccountMeta::new_readonly(transfer_config.key(), false),
            // authority = Store PDA (signer via invoke_signed, must be mutable per callee contract)
            AccountMeta::new(ctx.accounts.store.key(), true),
            // vault_authority (PDA of the transfer program)
            AccountMeta::new_readonly(vault_authority.key(), false),
            AccountMeta::new(vault_token_account.key(), false),
            AccountMeta::new(recipient_token_account.key(), false),
            AccountMeta::new_readonly(mint.key(), false),
            AccountMeta::new_readonly(token_program.key(), false),
        ];

        // Create CPI instruction
        let transfer_program_id = Pubkey::from_str(TRANSFER_PROGRAM_ID_STR).expect("Invalid hardcoded program id");
        let cpi_instruction = anchor_lang::solana_program::instruction::Instruction {
            program_id: transfer_program_id,
            accounts,
            data: instruction_data,
        };

        // Execute the CPI with Store PDA as signer
        let signer_seeds: &[&[u8]] = &[STORE_SEED, &[ctx.accounts.store.bump]];
        anchor_lang::solana_program::program::invoke_signed(
            &cpi_instruction,
            &[
                transfer_program.to_account_info(),
                transfer_config.to_account_info(),
                ctx.accounts.store.to_account_info(), // authority (Store PDA)
                vault_authority.to_account_info(),
                vault_token_account.to_account_info(),
                recipient_token_account.to_account_info(),
                mint.to_account_info(),
                token_program.to_account_info(),
            ],
            &[signer_seeds],
        )?;

        Ok(())
    }
}

#[error_code]
pub enum ErrorCode {
    #[msg("Invalid message format")]
    InvalidMessageFormat,
    #[msg("Invalid ABI encoding for payout message")]
    InvalidAbiEncoding,
    #[msg("Missing transfer contract accounts")]
    MissingTransferAccounts,
    #[msg("Missing Endpoint::clear accounts in remaining_accounts")]
    MissingClearAccounts,
    #[msg("Net amount is zero")]
    ZeroNetAmount,
    #[msg("Net amount exceeds u64 range")]
    AmountOverflow,
    #[msg("Invalid peer address: sender does not match configured peer")]
    InvalidPeerAddress,
}

