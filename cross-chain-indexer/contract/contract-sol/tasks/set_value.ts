import * as anchor from "@coral-xyz/anchor";
import { PublicKey } from "@solana/web3.js";
import { MyOapp } from "../target/types/my_oapp";

// snake_case helper to read CLI args in both snake and kebab styles
function get_arg(name: string): string | undefined {
    const args = process.argv.slice(2);
    const kebab = `--${name.replace(/_/g, "-")}=`;
    const snake = `--${name}=`;
    for (const a of args) {
        if (a.startsWith(kebab)) return a.slice(kebab.length);
        if (a.startsWith(snake)) return a.slice(snake.length);
    }
}

export async function set_value(
    program: any,
    provider: anchor.AnchorProvider,
    transfer_contract_address?: string
) {
    console.log("Setting transfer_contract...");

    const [store_pda] = PublicKey.findProgramAddressSync(
        [Buffer.from("Store")],
        program.programId
    );

    try {
        const tx = await program.methods
            .setValue({
                transferContract: transfer_contract_address ? new PublicKey(transfer_contract_address) : null,
            })
            .accounts({
                store: store_pda,
                admin: provider.wallet.publicKey,
            })
            .rpc();

        console.log("✅ set_value tx:", tx);
        console.log("transfer_contract:", transfer_contract_address ?? "null");
        return tx;
    } catch (error) {
        console.error("❌ Failed to set transfer_contract:", error);
        throw error;
    }
}

async function main() {
    const provider = anchor.AnchorProvider.env();

    // Resolve program id: prefer env, fallback to the default used in lib.rs
    const program_id_str = process.env.MYOAPP_ID || process.env.PROGRAM_ID || "41NCdrEvXhQ4mZgyJkmqYxL6A1uEmnraGj31UJ6PsXd3";
    const program_id = new PublicKey(program_id_str);

    const idl = require("../target/idl/my_oapp.json");
    const program = new (anchor as any).Program(idl, program_id, provider) as any;

    const transfer_contract_address = get_arg("transfer_contract");
    if (!transfer_contract_address) {
        console.error("Usage: npx ts-node tasks/setValue.ts --transfer_contract=<PROGRAM_ID> [--program_id=<MY_OAPP_PROGRAM_ID>]");
        process.exit(1);
    }

    await set_value(program, provider, transfer_contract_address);
}

// Execute when run directly via ts-node
if (require.main === module) {
    main().catch((e) => {
        console.error(e);
        process.exit(1);
    });
}
