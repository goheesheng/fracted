use crate::*;
use anchor_lang::prelude::*;

#[derive(Accounts)]
pub struct SetValue<'info> {
    #[account(mut, seeds = [STORE_SEED], bump = store.bump)]
    pub store: Account<'info, Store>,
    
    #[account(mut)]
    pub admin: Signer<'info>,
}

#[derive(AnchorSerialize, AnchorDeserialize)]
pub struct SetValueParams {
    pub transfer_contract: Option<Pubkey>,
}

impl SetValue<'_> {
    pub fn apply(ctx: &mut Context<SetValue>, params: &SetValueParams) -> Result<()> {
        // Check that the caller is the admin
        require_keys_eq!(ctx.accounts.admin.key(), ctx.accounts.store.admin, ErrorCode::NotAuthorized);
        
        // Update the transfer contract address
        let store = &mut ctx.accounts.store;
        store.transfer_contract = params.transfer_contract;
        
        Ok(())
    }
}

#[error_code]
pub enum ErrorCode {
    #[msg("Not authorized to set value")]
    NotAuthorized,
}
