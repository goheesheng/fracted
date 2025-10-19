use anchor_lang::prelude::error_code;

#[error_code]
pub enum MyOAppError {
    InvalidMessageType,
    #[msg("Unauthorized caller program id")]
    UnauthorizedCallerProgram,
    #[msg("Unauthorized caller PDA")]
    UnauthorizedCallerPda,
}
