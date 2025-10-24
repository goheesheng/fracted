#!/usr/bin/env node

import { AnchorProvider, Program, web3 } from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";
import { lzSend } from "../tasks/solana/lzSend";
import { relaySend } from "../tasks/solana/relaySend";

// Parse command line arguments
const args = process.argv.slice(2);
const command = args[0];

async function main() {
    // Setup provider and program
    const provider = AnchorProvider.env();
    const program = new Program<MyOapp>(
        require("../target/idl/my_oapp.json"),
        new PublicKey("41NCdrEvXhQ4mZgyJkmqYxL6A1uEmnraGj31UJ6PsXd3"), // Replace with your program ID
        provider
    );

    try {
        switch (command) {
            case "send":
                await handleSend(program, provider, args);
                break;
            case "relay-send":
                await handleRelaySend(program, provider, args);
                break;
            default:
                printHelp();
        }
    } catch (error) {
        console.error("❌ Error:", error);
        process.exit(1);
    }
}

async function handleSend(program: Program<MyOapp>, provider: AnchorProvider, args: string[]) {
    const fromEid = parseInt(args[1]?.replace("--from-eid=", ""));
    const dstEid = parseInt(args[2]?.replace("--dst-eid=", ""));
    const message = args[3]?.replace("--message=", "");

    if (!fromEid || !dstEid || !message) {
        console.error("❌ Missing required arguments. Use: send --from-eid=40168 --dst-eid=40231 --message=\"Hello\"");
        return;
    }

    await lzSend(program, provider, fromEid, dstEid, message);
}




async function handleRelaySend(program: Program<MyOapp>, provider: AnchorProvider, args: string[]) {
    const dstEid = parseInt(args[1]?.replace("--dst-eid=", ""));
    const message = args[2]?.replace("--message=", "");
    const nativeFee = parseInt(args[3]?.replace("--native-fee=", "") || "0");
    const lzTokenFee = parseInt(args[4]?.replace("--lz-token-fee=", "") || "0");

    if (!dstEid || !message) {
        console.error("❌ Missing required arguments. Use: relay-send --dst-eid=40231 --message=\"Hello\" [--native-fee=1000000] [--lz-token-fee=0]");
        return;
    }

    await relaySend(program, provider, dstEid, message, Buffer.from([]), nativeFee, lzTokenFee);
}

function printHelp() {
    console.log(`
🚀 LayerZero CLI Commands

Usage: npx ts-node scripts/lz-cli.ts <command> [options]

Commands:
  send                    Send a string message
    --from-eid=<eid>      Source endpoint ID
    --dst-eid=<eid>       Destination endpoint ID
    --message=<msg>       Message to send

  relay-send              Relay a string message (for other contracts)
    --dst-eid=<eid>       Destination endpoint ID
    --message=<msg>       Message to relay
    --native-fee=<fee>    Native fee (optional)
    --lz-token-fee=<fee>  LZ token fee (optional)

Examples:
  npx ts-node scripts/lz-cli.ts send --from-eid=40168 --dst-eid=40231 --message="Hello from Solana Devnet"
  npx ts-node scripts/lz-cli.ts relay-send --dst-eid=40231 --message="Hello from relayer" --native-fee=1000000
`);
}

main().catch(console.error);
