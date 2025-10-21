import { AnchorProvider, Program, web3 } from "@coral-xyz/anchor";
import { PublicKey, SystemProgram } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";

/**
 * Test function to simulate lz_receive call
 * Note: In real scenario, lz_receive can only be called by LayerZero Executor
 * This is just for testing purposes
 */
export async function testLzReceive(
    program: Program<MyOapp>,
    provider: AnchorProvider,
    transferContractAddress: string
) {
    console.log("🧪 Testing lz_receive functionality...");

    // Get the store PDA
    const [storePda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Store")],
        program.programId
    );

    // Get peer config PDA (for testing, using src_eid = 0)
    const [peerPda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Peer"), storePda.toBytes(), Buffer.from([0, 0, 0, 0])],
        program.programId
    );

    // Example token payout message data
    const tag = 101; // TAG_TOKEN_PAYOUT
    const dstToken = new PublicKey("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"); // USDC mint
    const merchant = new PublicKey("11111111111111111111111111111111"); // Example merchant address
    const netAmount = 1000000; // 1 USDC (6 decimals)

    // Encode message: (uint8 tag, address dstToken, address merchant, uint256 netAmount)
    const message = Buffer.alloc(128);
    message.writeUInt8(tag, 0);
    message.set(dstToken.toBytes(), 32);
    message.set(merchant.toBytes(), 64);
    message.writeBigUInt64LE(BigInt(netAmount), 96);

    console.log("📝 Message data:");
    console.log("Tag:", tag);
    console.log("dstToken:", dstToken.toString());
    console.log("merchant:", merchant.toString());
    console.log("netAmount:", netAmount);

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

        // Note: In real scenario, lz_receive would be called by LayerZero Executor
        // with proper LzReceiveParams and remaining accounts
        console.log("⚠️  lz_receive can only be called by LayerZero Executor");
        console.log("This test shows the message format that would be processed");
        
        return {
            store: storePda,
            peer: peerPda,
            message: message,
            transferContract: transferContractAddress
        };

    } catch (error) {
        console.error("❌ Error in test:", error);
        throw error;
    }
}

/**
 * Create a mock LzReceiveParams for testing
 */
export function createMockLzReceiveParams(
    srcEid: number,
    sender: string,
    nonce: number,
    guid: string,
    message: Buffer
) {
    return {
        srcEid,
        sender: new PublicKey(sender),
        nonce,
        guid: new PublicKey(guid),
        message: Array.from(message)
    };
}

/**
 * Example of how to call lz_receive with proper accounts
 * This is what LayerZero Executor would do
 */
export function getLzReceiveAccounts(
    program: Program<MyOapp>,
    transferContractAddress: string,
    vaultTokenAccount: string,
    recipientTokenAccount: string,
    mintAddress: string,
    configAccount: string
) {
    const [storePda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Store")],
        program.programId
    );

    const [peerPda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Peer"), storePda.toBytes(), Buffer.from([0, 0, 0, 0])],
        program.programId
    );

    return {
        store: storePda,
        peer: peerPda,
        transferProgram: new PublicKey(transferContractAddress),
        transferConfig: new PublicKey(configAccount),
        vaultAuthority: new PublicKey("VaultAuthorityAddress"), // Replace with actual
        vaultTokenAccount: new PublicKey(vaultTokenAccount),
        recipientTokenAccount: new PublicKey(recipientTokenAccount),
        mint: new PublicKey(mintAddress),
        tokenProgram: new PublicKey("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"),
    };
}

// Example usage:
// const transferContractAddress = "GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1";
// await testLzReceive(program, provider, transferContractAddress);
