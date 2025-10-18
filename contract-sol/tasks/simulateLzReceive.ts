import { AnchorProvider, Program, web3 } from "@coral-xyz/anchor";
import { PublicKey, SystemProgram } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";

/**
 * Simulate LayerZero Executor calling lz_receive
 * This is for testing purposes only - in production, only LayerZero Executor can call this
 */
export async function simulateLzReceive(
    program: Program<MyOapp>,
    provider: AnchorProvider,
    transferContractAddress: string
) {
    console.log("🎭 Simulating LayerZero Executor calling lz_receive...");

    // Get the store PDA
    const [storePda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Store")],
        program.programId
    );

    // Get peer config PDA
    const [peerPda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Peer"), storePda.toBytes(), Buffer.from([0, 0, 0, 0])],
        program.programId
    );

    // Create mock LzReceiveParams
    const mockParams = {
        srcEid: 101, // Arbitrum Sepolia
        sender: new PublicKey("11111111111111111111111111111111"), // Mock sender
        nonce: 12345,
        guid: new PublicKey("22222222222222222222222222222222"), // Mock GUID
        message: createTokenPayoutMessage()
    };

    console.log("📝 Mock LzReceiveParams:");
    console.log("srcEid:", mockParams.srcEid);
    console.log("sender:", mockParams.sender.toString());
    console.log("nonce:", mockParams.nonce);
    console.log("guid:", mockParams.guid.toString());

    try {
        // First, set the transfer contract address
        console.log("Setting transfer contract address...");
        await program.methods
            .setValue({
                transferContract: new PublicKey(transferContractAddress)
            })
            .accounts({
                store: storePda,
                admin: provider.wallet.publicKey,
            })
            .rpc();

        console.log("✅ Transfer contract address set");

        // Note: The actual lz_receive call would look like this:
        // await program.methods
        //     .lzReceive(mockParams)
        //     .accounts({
        //         store: storePda,
        //         peer: peerPda,
        //         // ... other accounts
        //     })
        //     .rpc();

        console.log("⚠️  lz_receive can only be called by LayerZero Executor");
        console.log("The message would be processed as follows:");
        console.log("1. Decode token payout message");
        console.log("2. Check if transfer contract is configured");
        console.log("3. Call transfer_out contract with decoded parameters");

        return {
            success: true,
            message: "Simulation completed - lz_receive would process the token payout message"
        };

    } catch (error) {
        console.error("❌ Error in simulation:", error);
        throw error;
    }
}

/**
 * Create a token payout message in the expected format
 */
function createTokenPayoutMessage(): number[] {
    const message = Buffer.alloc(128);
    
    // Tag: 101 (TAG_TOKEN_PAYOUT)
    message.writeUInt8(101, 0);
    
    // dstToken: USDC mint address
    const dstToken = new PublicKey("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v");
    message.set(dstToken.toBytes(), 32);
    
    // merchant: Example merchant address
    const merchant = new PublicKey("11111111111111111111111111111111");
    message.set(merchant.toBytes(), 64);
    
    // netAmount: 1 USDC (6 decimals)
    const netAmount = 1000000;
    message.writeBigUInt64LE(BigInt(netAmount), 96);
    
    return Array.from(message);
}

/**
 * Test the message decoding logic
 */
export function testMessageDecoding() {
    console.log("🔍 Testing message decoding logic...");
    
    const message = createTokenPayoutMessage();
    const messageBuffer = Buffer.from(message);
    
    if (messageBuffer.length === 128) {
        const tag = messageBuffer.readUInt8(0);
        const dstTokenBytes = messageBuffer.subarray(32, 64);
        const merchantBytes = messageBuffer.subarray(64, 96);
        const netAmountBytes = messageBuffer.subarray(96, 128);
        
        const dstToken = new PublicKey(dstTokenBytes);
        const merchant = new PublicKey(merchantBytes);
        const netAmount = Number(netAmountBytes.readBigUInt64LE(0));
        
        console.log("✅ Message decoded successfully:");
        console.log("Tag:", tag);
        console.log("dstToken:", dstToken.toString());
        console.log("merchant:", merchant.toString());
        console.log("netAmount:", netAmount);
        
        return { tag, dstToken, merchant, netAmount };
    } else {
        console.error("❌ Invalid message length:", messageBuffer.length);
        return null;
    }
}

// Example usage:
// const transferContractAddress = "GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1";
// await simulateLzReceive(program, provider, transferContractAddress);
// testMessageDecoding();
