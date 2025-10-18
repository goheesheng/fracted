import { AnchorProvider, Program, web3 } from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";

export async function lzSend(
    program: Program<MyOapp>,
    provider: AnchorProvider,
    fromEid: number,
    dstEid: number,
    message: string
) {
    console.log(`📤 Sending message from EID ${fromEid} to EID ${dstEid}`);
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
        new PublicKey("0x4D9434eBd2A8c0B97B4f47f702c3EDb65b4F9B0c")
    );

    try {
        const tx = await program.methods
            .send({
                dstEid,
                message,
                options: Buffer.from([]),
                nativeFee: 1000000,
                lzTokenFee: 0,
            })
            .accounts({
                store: storePda,
                peer: peerPda,
                endpoint: endpointPda,
            })
            .rpc();

        console.log("✅ Message sent successfully!");
        console.log("Transaction signature:", tx);
        
        return tx;
    } catch (error) {
        console.error("❌ Error sending message:", error);
        throw error;
    }
}

// Example usage:
// await lzSend(program, provider, 40168, 40231, "Hello from Solana Devnet");
