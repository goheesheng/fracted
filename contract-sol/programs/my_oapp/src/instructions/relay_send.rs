use crate::*;
use crate::errors::MyOAppError;
use anchor_lang::prelude::*;
use oapp::endpoint::{
    instructions::SendParams, state::EndpointSettings, ENDPOINT_SEED, ID as ENDPOINT_ID,
};
use std::str::FromStr;

#[derive(Accounts)]
#[instruction(params: RelaySendParams)]
pub struct RelaySend<'info> {
    #[account(
        seeds = [
            PEER_SEED,
            &store.key().to_bytes(),
            &params.dst_eid.to_be_bytes()
        ],
        bump = peer.bump
    )]
    /// Configuration for the destination chain. Holds the peer address and any
    /// enforced messaging options.
    pub peer: Account<'info, PeerConfig>,
    #[account(seeds = [STORE_SEED], bump = store.bump)]
    /// OApp Store PDA that signs the send instruction
    pub store: Account<'info, Store>,
    #[account(seeds = [ENDPOINT_SEED], bump = endpoint.bump, seeds::program = ENDPOINT_ID)]
    pub endpoint: Account<'info, EndpointSettings>,
    /// The caller who wants to send a message (must be authorized program's PDA)
    pub caller: Signer<'info>,
}

impl<'info> RelaySend<'info> {
    pub fn apply(ctx: &mut Context<RelaySend>, params: &RelaySendParams) -> Result<()> {
        // Enforce caller identity
        // Allowed program id and PDA as requested
        let allowed_program_id = Pubkey::from_str("GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1")
            .map_err(|_| error!(MyOAppError::UnauthorizedCallerProgram))?;
        let allowed_pda = Pubkey::from_str("X6ci3v3wgpFrRvmsFsjeemr1EDeaHaok23UsehuQcvn")
            .map_err(|_| error!(MyOAppError::UnauthorizedCallerPda))?;

        // Check PDA key matches exactly
        require_keys_eq!(ctx.accounts.caller.key(), allowed_pda, MyOAppError::UnauthorizedCallerPda);

        // Check owner program id of the PDA account
        let caller_info = ctx.accounts.caller.to_account_info();
        require_keys_eq!(*caller_info.owner, allowed_program_id, MyOAppError::UnauthorizedCallerProgram);

        // Serialize the message according to our codec
        let message = msg_codec::encode(&params.message);
        
        // Prepare the seeds for the OApp Store PDA, which is used to sign the CPI call to the Endpoint program.
        let seeds: &[&[u8]] = &[STORE_SEED, &[ctx.accounts.store.bump]];

        // Prepare the SendParams for the Endpoint::send CPI call.
        let send_params = SendParams {
            dst_eid: params.dst_eid,
            receiver: ctx.accounts.peer.peer_address,
            message,
            options: ctx
                .accounts
                .peer
                .enforced_options
                .combine_options(&None::<Vec<u8>>, &params.options)?,
            native_fee: params.native_fee,
            lz_token_fee: params.lz_token_fee,
        };
        
        // Call the Endpoint::send CPI to send the message.
        oapp::endpoint_cpi::send(
            ENDPOINT_ID,
            ctx.accounts.store.key(),
            ctx.remaining_accounts,
            seeds,
            send_params,
        )?;
        
        // Emit event for tracking
        emit!(RelaySendEvent {
            caller: ctx.accounts.caller.key(),
            dst_eid: params.dst_eid,
            message: params.message.clone(),
            native_fee: params.native_fee,
            lz_token_fee: params.lz_token_fee,
        });
        
        Ok(())
    }
}

#[derive(Clone, AnchorSerialize, AnchorDeserialize)]
pub struct RelaySendParams {
    pub dst_eid: u32,
    pub message: String,
    pub options: Vec<u8>,
    pub native_fee: u64,
    pub lz_token_fee: u64,
}

#[event]
pub struct RelaySendEvent {
    pub caller: Pubkey,
    pub dst_eid: u32,
    pub message: String,
    pub native_fee: u64,
    pub lz_token_fee: u64,
}
