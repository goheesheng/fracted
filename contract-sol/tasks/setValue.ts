import { AnchorProvider, Program, web3 } from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";

export async function setValue(
    program: Program<MyOapp>,
    provider: AnchorProvider,
    transferContractAddress?: string
) {
    console.log("Setting transfer contract value...");

    // Get the store PDA
    const [storePda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Store")],
        program.programId
    );

    try {
        // Set the transfer contract address
        const tx = await program.methods
            .setValue({
                transferContract: transferContractAddress ? new PublicKey(transferContractAddress) : null
            })
            .accounts({
                store: storePda,
                admin: provider.wallet.publicKey,
            })
            .rpc();

        console.log("✅ setValue transaction signature:", tx);
        console.log("Transfer contract address set to:", transferContractAddress || "null");
        
        return tx;
    } catch (error) {
        console.error("❌ Error setting value:", error);
        throw error;
    }
}

// Example usage:
// const transferContractAddress = "GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1"; // Replace with actual transfer contract address
// await setValue(program, provider, transferContractAddress);
