#!/usr/bin/env node

import { AnchorProvider, Program, web3 } from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";
import { lzSend } from "../tasks/lzSend";
import { lzSendTokenPayout } from "../tasks/lzSendTokenPayout";
import { lzQuote } from "../tasks/lzQuote";
import { set_value } from "../tasks/set_value";
import { simulateLzReceive } from "../tasks/simulateLzReceive";
import { relaySend } from "../tasks/relaySend";
import { relayTokenPayout } from "../tasks/relayTokenPayout";

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
            case "send-token-payout":
                await handleSendTokenPayout(program, provider, args);
                break;
            case "quote":
                await handleQuote(program, provider, args);
                break;
            case "set-value":
                await handleSetValue(program, provider, args);
                break;
            case "simulate-receive":
                await handleSimulateReceive(program, provider, args);
                break;
            case "relay-send":
                await handleRelaySend(program, provider, args);
                break;
            case "relay-token-payout":
                await handleRelayTokenPayout(program, provider, args);
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

async function handleSendTokenPayout(program: Program<MyOapp>, provider: AnchorProvider, args: string[]) {
    const fromEid = parseInt(args[1]?.replace("--from-eid=", ""));
    const dstEid = parseInt(args[2]?.replace("--dst-eid=", ""));
    const dstToken = args[3]?.replace("--dst-token=", "");
    const merchant = args[4]?.replace("--merchant=", "");
    const netAmount = parseInt(args[5]?.replace("--amount=", ""));

    if (!fromEid || !dstEid || !dstToken || !merchant || !netAmount) {
        console.error("❌ Missing required arguments. Use: send-token-payout --from-eid=40168 --dst-eid=40231 --dst-token=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v --merchant=11111111111111111111111111111111 --amount=1000000");
        return;
    }

    await lzSendTokenPayout(program, provider, fromEid, dstEid, dstToken, merchant, netAmount);
}

async function handleQuote(program: Program<MyOapp>, provider: AnchorProvider, args: string[]) {
    const dstEid = parseInt(args[1]?.replace("--dst-eid=", ""));
    const message = args[2]?.replace("--message=", "");

    if (!dstEid || !message) {
        console.error("❌ Missing required arguments. Use: quote --dst-eid=40231 --message=\"Hello\"");
        return;
    }

    await lzQuote(program, provider, dstEid, message);
}

async function handleSetValue(program: Program<MyOapp>, provider: AnchorProvider, args: string[]) {
    const transferContract = args[1]?.replace("--transfer-contract=", "");

    if (!transferContract) {
        console.error("❌ Missing required arguments. Use: set-value --transfer-contract=GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1");
        return;
    }

    await set_value(program, provider, transferContract);
}

async function handleSimulateReceive(program: Program<MyOapp>, provider: AnchorProvider, args: string[]) {
    const transferContract = args[1]?.replace("--transfer-contract=", "");

    if (!transferContract) {
        console.error("❌ Missing required arguments. Use: simulate-receive --transfer-contract=GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1");
        return;
    }

    await simulateLzReceive(program, provider, transferContract);
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

async function handleRelayTokenPayout(program: Program<MyOapp>, provider: AnchorProvider, args: string[]) {
    const dstEid = parseInt(args[1]?.replace("--dst-eid=", ""));
    const dstToken = args[2]?.replace("--dst-token=", "");
    const merchant = args[3]?.replace("--merchant=", "");
    const netAmount = parseInt(args[4]?.replace("--amount=", ""));
    const nativeFee = parseInt(args[5]?.replace("--native-fee=", "") || "0");
    const lzTokenFee = parseInt(args[6]?.replace("--lz-token-fee=", "") || "0");

    if (!dstEid || !dstToken || !merchant || !netAmount) {
        console.error("❌ Missing required arguments. Use: relay-token-payout --dst-eid=40231 --dst-token=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v --merchant=11111111111111111111111111111111 --amount=1000000 [--native-fee=1000000] [--lz-token-fee=0]");
        return;
    }

    await relayTokenPayout(program, provider, dstEid, dstToken, merchant, netAmount, Buffer.from([]), nativeFee, lzTokenFee);
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

  send-token-payout       Send a token payout message
    --from-eid=<eid>      Source endpoint ID
    --dst-eid=<eid>       Destination endpoint ID
    --dst-token=<addr>    Token mint address
    --merchant=<addr>     Merchant address
    --amount=<amount>     Amount to transfer

  quote                   Get message quote
    --dst-eid=<eid>       Destination endpoint ID
    --message=<msg>       Message to quote

  set-value               Set transfer contract address
    --transfer-contract=<addr>  Transfer contract address

  simulate-receive        Simulate lz_receive call
    --transfer-contract=<addr>  Transfer contract address

  relay-send              Relay a string message (for other contracts)
    --dst-eid=<eid>       Destination endpoint ID
    --message=<msg>       Message to relay
    --native-fee=<fee>    Native fee (optional)
    --lz-token-fee=<fee>  LZ token fee (optional)

  relay-token-payout      Relay a token payout message (for other contracts)
    --dst-eid=<eid>       Destination endpoint ID
    --dst-token=<addr>    Token mint address
    --merchant=<addr>     Merchant address
    --amount=<amount>     Amount to transfer
    --native-fee=<fee>    Native fee (optional)
    --lz-token-fee=<fee>  LZ token fee (optional)

  get-addresses           Get all contract PDA addresses (after deployment)
  calculate-pdas          Calculate PDA addresses (before deployment)
    --program-id=<id>     Program ID to calculate PDAs for
  check-stability         Check program ID stability

Examples:
  npx ts-node scripts/lz-cli.ts send --from-eid=40168 --dst-eid=40231 --message="Hello from Solana Devnet"
  npx ts-node scripts/lz-cli.ts send-token-payout --from-eid=40168 --dst-eid=40231 --dst-token=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v --merchant=11111111111111111111111111111111 --amount=1000000
  npx ts-node scripts/lz-cli.ts quote --dst-eid=40231 --message="Hello"
  npx ts-node scripts/lz-cli.ts set-value --transfer-contract=GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1
  npx ts-node scripts/lz-cli.ts simulate-receive --transfer-contract=GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1
  npx ts-node scripts/lz-cli.ts relay-send --dst-eid=40231 --message="Hello from relayer" --native-fee=1000000
  npx ts-node scripts/lz-cli.ts relay-token-payout --dst-eid=40231 --dst-token=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v --merchant=11111111111111111111111111111111 --amount=1000000 --native-fee=1000000
  npx ts-node scripts/lz-cli.ts get-addresses
  npx ts-node scripts/lz-cli.ts calculate-pdas --program-id=41NCdrEvXhQ4mZgyJkmqYxL6A1uEmnraGj31UJ6PsXd3
  npx ts-node scripts/lz-cli.ts check-stability
`);
}

main().catch(console.error);
