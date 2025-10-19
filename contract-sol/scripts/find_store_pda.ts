import { PublicKey } from "@solana/web3.js";

// 目标程序 ID 还有可能 CV1qjq8phMMpxv62TExA9PpvTyZx58TNCqkFB2QQgJXH
const programId = new PublicKey("CV1qjq8phMMpxv62TExA9PpvTyZx58TNCqkFB2QQgJXH");

// 用 seed 生成 PDA
const [storePda, bump] = PublicKey.findProgramAddressSync(
  [Buffer.from("Store")],  // STORE_SEED
  programId
);

console.log("Store PDA:", storePda.toBase58());
console.log("Bump:", bump);
