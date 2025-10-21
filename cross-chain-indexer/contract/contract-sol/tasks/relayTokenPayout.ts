import { AnchorProvider, Program, web3 } from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";

/**
 * Relay a token payout message to another chain
 * This function can be called by other contracts to send token payout messages through this OApp
 */
export async function relayTokenPayout(
    program: Program<MyOapp>,
    provider: AnchorProvider,
    dstEid: number,
    dstToken: string,
    merchant: string,
    netAmount: number,
    options: Buffer = Buffer.from([]),
    nativeFee: number = 0,
    lzTokenFee: number = 0
) {
    console.log(`💰 Relaying token payout to EID ${dstEid}`);
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
        new PublicKey("0x4D9434eBd2A8c0B97B4f47f702c3EDb65b4F9B0c") // LayerZero Endpoint V2
    );

    try {
        const tx = await program.methods
            .relayTokenPayout({
                dstEid,
                dstToken: new PublicKey(dstToken),
                merchant: new PublicKey(merchant),
                netAmount,
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

        console.log("✅ Token payout message relayed successfully!");
        console.log("Transaction signature:", tx);
        
        return tx;
    } catch (error) {
        console.error("❌ Error relaying token payout:", error);
        throw error;
    }
}

// Example usage:
// await relayTokenPayout(program, provider, 40231, "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", "11111111111111111111111111111111", 1000000, [], 1000000, 0);
