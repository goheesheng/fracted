import { AnchorProvider, Program, web3 } from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";

export async function lzSendTokenPayout(
    program: Program<MyOapp>,
    provider: AnchorProvider,
    fromEid: number,
    dstEid: number,
    dstToken: string,
    merchant: string,
    netAmount: number
) {
    console.log(`💰 Sending token payout from EID ${fromEid} to EID ${dstEid}`);
    console.log(`Token: ${dstToken}`);
    console.log(`Merchant: ${merchant}`);
    console.log(`Amount: ${netAmount}`);

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

    // Create token payout message: (uint8 tag, address dstToken, address merchant, uint256 netAmount)
    const message = Buffer.alloc(128);
    message.writeUInt8(101, 0); // TAG_TOKEN_PAYOUT
    message.set(new PublicKey(dstToken).toBytes(), 32);
    message.set(new PublicKey(merchant).toBytes(), 64);
    message.writeBigUInt64LE(BigInt(netAmount), 96);

    try {
        const tx = await program.methods
            .send({
                dstEid,
                message: Array.from(message),
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

        console.log("✅ Token payout message sent successfully!");
        console.log("Transaction signature:", tx);
        
        return tx;
    } catch (error) {
        console.error("❌ Error sending token payout:", error);
        throw error;
    }
}

// Example usage:
// await lzSendTokenPayout(program, provider, 40168, 40231, "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", "11111111111111111111111111111111", 1000000);
