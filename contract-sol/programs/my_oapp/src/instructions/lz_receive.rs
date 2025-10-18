use crate::*;
use anchor_lang::prelude::*;
use oapp::{
    endpoint::{
        cpi::accounts::Clear,
        instructions::ClearParams,
        ConstructCPIContext, ID as ENDPOINT_ID,
    },
    LzReceiveParams,
};

#[derive(Accounts)]
#[instruction(params: LzReceiveParams)]
pub struct LzReceive<'info> {
    /// OApp Store PDA.  This account represents the "address" of your OApp on
    /// Solana and can contain any state relevant to your application.
    /// Customize the fields in `Store` as needed.
    #[account(mut, seeds = [STORE_SEED], bump = store.bump)]
    pub store: Account<'info, Store>,
    /// Peer config PDA for the sending chain. Ensures `params.sender` can only be the allowed peer from that remote chain.
    #[account(
        seeds = [PEER_SEED, &store.key().to_bytes(), &params.src_eid.to_be_bytes()],
        bump = peer.bump,
        constraint = params.sender == peer.peer_address
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

        // Process the message based on its format
        // Token payout path: (uint8 tag, address dstToken, address merchant, uint256 netAmount)
        if params.message.len() == 128 {
            // Try to decode as token payout message
            if let Ok((tag, dst_token, merchant, net_amount)) = Self::decode_token_payout_message(&params.message) {
                if tag == 101 { // TAG_TOKEN_PAYOUT
                    // Check if transfer contract is configured
                    if let Some(transfer_contract) = ctx.accounts.store.transfer_contract {
                        Self::call_transfer_out(
                            ctx,
                            transfer_contract,
                            dst_token,
                            merchant,
                            net_amount,
                        )?;
                        return Ok(());
                    }
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
        // Decode the message: (uint8 tag, address dstToken, address merchant, uint256 netAmount)
        // Each field is 32 bytes in Solana
        if message.len() != 128 {
            return Err(ErrorCode::InvalidMessageFormat.into());
        }

        let tag = message[0];
        let dst_token_bytes: [u8; 32] = message[32..64].try_into().map_err(|_| ErrorCode::InvalidMessageFormat)?;
        let merchant_bytes: [u8; 32] = message[64..96].try_into().map_err(|_| ErrorCode::InvalidMessageFormat)?;
        let net_amount_bytes: [u8; 32] = message[96..128].try_into().map_err(|_| ErrorCode::InvalidMessageFormat)?;

        let dst_token = Pubkey::new_from_array(dst_token_bytes);
        let merchant = Pubkey::new_from_array(merchant_bytes);
        let net_amount = u64::from_le_bytes(net_amount_bytes[0..8].try_into().map_err(|_| ErrorCode::InvalidMessageFormat)?);

        Ok((tag, dst_token, merchant, net_amount))
    }

    fn call_transfer_out(
        ctx: &mut Context<LzReceive>,
        transfer_contract: Pubkey,
        _dst_token: Pubkey,
        _merchant: Pubkey,
        amount: u64,
    ) -> Result<()> {
        // Check that all required accounts are provided
        let transfer_program = ctx.accounts.transfer_program.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let transfer_config = ctx.accounts.transfer_config.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let vault_authority = ctx.accounts.vault_authority.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let vault_token_account = ctx.accounts.vault_token_account.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let recipient_token_account = ctx.accounts.recipient_token_account.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let mint = ctx.accounts.mint.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;
        let token_program = ctx.accounts.token_program.as_ref().ok_or(ErrorCode::MissingTransferAccounts)?;

        // Create instruction data for transfer_out
        // transfer_out instruction: (discriminator: 8 bytes + amount: 8 bytes)
        let mut instruction_data = Vec::new();
        // Add discriminator for transfer_out instruction (you need to get this from the transfer_contract)
        // For now, using a placeholder - you'll need to get the actual discriminator
        instruction_data.extend_from_slice(&[0; 8]); // Placeholder discriminator
        instruction_data.extend_from_slice(&amount.to_le_bytes());

        // Create accounts for the CPI call
        let accounts = vec![
            AccountMeta::new_readonly(transfer_config.key(), false),
            AccountMeta::new_readonly(vault_authority.key(), true),
            AccountMeta::new_readonly(vault_authority.key(), false),
            AccountMeta::new(vault_token_account.key(), false),
            AccountMeta::new(recipient_token_account.key(), false),
            AccountMeta::new_readonly(mint.key(), false),
            AccountMeta::new_readonly(token_program.key(), false),
        ];

        // Create CPI instruction
        let cpi_instruction = anchor_lang::solana_program::instruction::Instruction {
            program_id: transfer_program.key(),
            accounts,
            data: instruction_data,
        };

        // Execute the CPI
        anchor_lang::solana_program::program::invoke(
            &cpi_instruction,
            &[
                transfer_config.to_account_info(),
                vault_authority.to_account_info(),
                vault_authority.to_account_info(),
                vault_token_account.to_account_info(),
                recipient_token_account.to_account_info(),
                mint.to_account_info(),
                token_program.to_account_info(),
            ],
        )?;

        Ok(())
    }
}

#[error_code]
pub enum ErrorCode {
    #[msg("Invalid message format")]
    InvalidMessageFormat,
    #[msg("Missing transfer contract accounts")]
    MissingTransferAccounts,
}

