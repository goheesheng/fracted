#!/usr/bin/env ts-node

import { AnchorProvider } from "@coral-xyz/anchor";
import { Connection, ComputeBudgetProgram, PublicKey, Transaction } from "@solana/web3.js";
import { lzReceive } from "@layerzerolabs/lz-solana-sdk-v2";
import { makeBytes32 } from "@layerzerolabs/devtools";

function getArg(name: string): string | undefined {
  const found = process.argv.find((a) => a.startsWith(`--${name}=`));
  return found ? found.split("=")[1] : undefined;
}

function printHelp() {
  console.log(`
调用说明: 调用 Solana 侧 LayerZero Executor 的 lz_receive 指令（测试/开发用途）

必填参数:
  --src-eid=<eid>           源链 EndpointId（十进制），如 40168
  --nonce=<nonce>           消息 nonce（字符串/整数）。建议与消息对应
  --sender=<hex32>          源 OApp 地址（32字节16进制，0x前缀）
  --guid=<hex32>            消息 GUID（32字节16进制，0x前缀）
  --message=<hex>           消息内容（16进制，0x前缀）。若要触发 transfer_out 路径，需128字节且首字节=101

可选参数:
  --program-id=<pubkey>     目标程序ID，默认 41NC...PsXd3，如本地链请显式传
  --receiver=<pubkey>       目标接收者（Store PDA）。缺省时按 programId 与 "Store" 计算
  --cu-limit=<n>            计算单元上限，默认 350000
  --lamports=<n>            随交易传入的 lamports，默认 0
  --cu-price=<microLamports>优先费(微lamports)

示例:
  npx ts-node scripts/call_lz_receive.ts \\
    --src-eid=40168 \\
    --nonce=12345 \\
    --sender=0x0000000000000000000000000000000000000000000000000000000000000000 \\
    --guid=0x0000000000000000000000000000000000000000000000000000000000000000 \\
    --message=0x0102abcd \\
    --program-id=41NCdrEvXhQ4mZgyJkmqYxL6A1uEmnraGj31UJ6PsXd3
`);
}

async function main() {
  const provider = AnchorProvider.env();
  const connection: Connection = provider.connection;
  const payerPubkey = provider.wallet.publicKey;

  const programIdArg = getArg("program-id");
  const srcEidArg = getArg("src-eid");
  const nonceArg = getArg("nonce");
  const senderHexArg = getArg("sender");
  const guidHexArg = getArg("guid");
  const messageHexArg = getArg("message");
  const receiverArg = getArg("receiver");
  const cuLimitArg = getArg("cu-limit");
  const lamportsArg = getArg("lamports");
  const cuPriceArg = getArg("cu-price");

  if (process.argv.includes("--help") || process.argv.includes("-h")) {
    printHelp();
    process.exit(0);
  }

  if (!srcEidArg || !nonceArg || !senderHexArg || !guidHexArg || !messageHexArg) {
    printHelp();
    process.exit(1);
  }

  const programId = new PublicKey(programIdArg ?? "41NCdrEvXhQ4mZgyJkmqYxL6A1uEmnraGj31UJ6PsXd3");

  const receiver = receiverArg
    ? receiverArg
    : (() => {
        const [storePda] = PublicKey.findProgramAddressSync([Buffer.from("Store")], programId);
        return storePda.toBase58();
      })();

  const tx = new Transaction();

  if (cuPriceArg) {
    tx.add(
      ComputeBudgetProgram.setComputeUnitPrice({
        microLamports: parseInt(cuPriceArg, 10),
      })
    );
  }

  const withComputeUnitLimit = parseInt(cuLimitArg ?? "350000", 10);
  const lamports = parseInt(lamportsArg ?? "0", 10);

  const ix = await lzReceive(
    connection,
    payerPubkey,
    {
      srcEid: parseInt(srcEidArg, 10),
      nonce: nonceArg,
      sender: makeBytes32(senderHexArg),
      receiver,
      guid: guidHexArg,
      message: messageHexArg,
    },
    Uint8Array.from([withComputeUnitLimit, lamports]),
    "confirmed"
  );

  tx.add(ix);

  const { blockhash, lastValidBlockHeight } = await connection.getLatestBlockhash("confirmed");
  tx.recentBlockhash = blockhash;
  tx.feePayer = payerPubkey;

  const sig = await provider.sendAndConfirm(tx, [], { commitment: "confirmed" });
  console.log("✅ lz_receive transaction sent:", sig);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});


