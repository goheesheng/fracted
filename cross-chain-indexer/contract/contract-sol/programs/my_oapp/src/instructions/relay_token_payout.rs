use crate::*;
use anchor_lang::prelude::*;
use oapp::endpoint::{
    instructions::SendParams, state::EndpointSettings, ENDPOINT_SEED, ID as ENDPOINT_ID,
};

#[derive(Accounts)]
#[instruction(params: RelayTokenPayoutParams)]
pub struct RelayTokenPayout<'info> {
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
    /// The caller who wants to send a token payout message (can be any program)
    pub caller: Signer<'info>,
}

impl<'info> RelayTokenPayout<'info> {
    pub fn apply(ctx: &mut Context<RelayTokenPayout>, params: &RelayTokenPayoutParams) -> Result<()> {
        // Create token payout message: (uint8 tag, address dstToken, address merchant, uint256 netAmount)
        let mut message = Vec::new();
        message.push(101u8); // TAG_TOKEN_PAYOUT
        message.extend_from_slice(&params.dst_token.to_bytes());
        message.extend_from_slice(&params.merchant.to_bytes());
        message.extend_from_slice(&params.net_amount.to_le_bytes());
        // Pad to 128 bytes total
        while message.len() < 128 {
            message.push(0);
        }
        
        // Prepare the seeds for the OApp Store PDA
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
        emit!(RelayTokenPayoutEvent {
            caller: ctx.accounts.caller.key(),
            dst_eid: params.dst_eid,
            dst_token: params.dst_token,
            merchant: params.merchant,
            net_amount: params.net_amount,
            native_fee: params.native_fee,
            lz_token_fee: params.lz_token_fee,
        });
        
        Ok(())
    }
}

#[derive(Clone, AnchorSerialize, AnchorDeserialize)]
pub struct RelayTokenPayoutParams {
    pub dst_eid: u32,
    pub dst_token: Pubkey,
    pub merchant: Pubkey,
    pub net_amount: u64,
    pub options: Vec<u8>,
    pub native_fee: u64,
    pub lz_token_fee: u64,
}

#[event]
pub struct RelayTokenPayoutEvent {
    pub caller: Pubkey,
    pub dst_eid: u32,
    pub dst_token: Pubkey,
    pub merchant: Pubkey,
    pub net_amount: u64,
    pub native_fee: u64,
    pub lz_token_fee: u64,
}
