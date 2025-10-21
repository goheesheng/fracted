import { AnchorProvider, Program, web3 } from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";

export async function lzQuote(
    program: Program<MyOapp>,
    provider: AnchorProvider,
    dstEid: number,
    message: string | number[]
) {
    console.log(`💸 Getting quote for message to EID ${dstEid}`);

    // Get the store PDA
    const [storePda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Store")],
        program.programId
    );

    try {
        const quote = await program.methods
            .quoteSend({
                dstEid,
                message: typeof message === 'string' ? Array.from(Buffer.from(message)) : message,
                options: Buffer.from([]),
            })
            .accounts({
                store: storePda,
            })
            .view();

        console.log("✅ Quote received:");
        console.log("Native fee:", quote.nativeFee.toString());
        console.log("LZ token fee:", quote.lzTokenFee.toString());
        
        return quote;
    } catch (error) {
        console.error("❌ Error getting quote:", error);
        throw error;
    }
}

// Example usage:
// await lzQuote(program, provider, 40231, "Hello from Solana Devnet");
