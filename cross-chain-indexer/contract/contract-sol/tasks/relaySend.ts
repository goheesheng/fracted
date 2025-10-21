import { AnchorProvider, Program, web3 } from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";

/**
 * Relay a string message to another chain
 * This function can be called by other contracts to send messages through this OApp
 */
export async function relaySend(
    program: Program<MyOapp>,
    provider: AnchorProvider,
    dstEid: number,
    message: string,
    options: Buffer = Buffer.from([]),
    nativeFee: number = 0,
    lzTokenFee: number = 0
) {
    console.log(`📤 Relaying message to EID ${dstEid}`);
    console.log(`Message: "${message}"`);

    // Get the store PDA
    const [storePda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Store")],
        program.programId
    );

    // Get peer config PDA
    const [peerPda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Peer"), storePda.toBytes(), Buffer.from(dstEid.toString(16).padStart(8, '0'), 'hex')],
        program.programId
    );

    // Get endpoint PDA
    const [endpointPda] = PublicKey.findProgramAddressSync(
        [Buffer.from("endpoint")],
        new PublicKey("0x4D9434eBd2A8c0B97B4f47f702c3EDb65b4F9B0c") // LayerZero Endpoint V2
    );

    try {
        const tx = await program.methods
            .relaySend({
                dstEid,
                message,
                options,
                nativeFee,
                lzTokenFee,
            })
            .accounts({
                peer: peerPda,
                store: storePda,
                endpoint: endpointPda,
                caller: provider.wallet.publicKey,
            })
            .rpc();

        console.log("✅ Message relayed successfully!");
        console.log("Transaction signature:", tx);
        
        return tx;
    } catch (error) {
        console.error("❌ Error relaying message:", error);
        throw error;
    }
}

// Example usage:
// await relaySend(program, provider, 40231, "Hello from relayer", [], 1000000, 0);
