import { ContractTransaction } from 'ethers'
// eslint-disable-next-line @typescript-eslint/no-var-requires
const { task } = require('hardhat/config')
import { PublicKey } from '@solana/web3.js'
import { endpointIdToNetwork } from '@layerzerolabs/lz-definitions'
import { Options } from '@layerzerolabs/lz-v2-utilities'

task(
    'lz:oapp:requestPayoutToken',
    'Request token payout cross-chain (pull srcToken, pay dstToken 97% to merchant)'
)
    .addParam('dstEid', 'Destination endpoint ID')
    .addParam('srcToken', 'Source chain token address to pull from caller')
    .addParam('dstToken', 'Destination token address (EVM 0x.. or Solana base58)')
    .addParam('merchant', 'Destination merchant address (EVM 0x.. or Solana base58)')
    .addParam('amount', 'Gross token amount in smallest units (e.g., 1 USDC = 1000000)')
    .setAction(
        async (
            args: {
                dstEid: string
                srcToken: string
                dstToken: string
                merchant: string
                amount: string
            },
            hre: any
        ) => {
            const [signer] = await hre.ethers.getSigners()
            const deployment = await hre.deployments.get('MyOApp')
            const myOApp = await hre.ethers.getContractAt('MyOApp', deployment.address, signer)

            const dstEidNum = Number(args.dstEid)
            console.log(`Initiating token payout from ${hre.network.name} to ${endpointIdToNetwork(dstEidNum)}`)
            console.log(`MyOApp:   ${deployment.address}`)
            console.log(`Signer:   ${signer.address}`)
            console.log(`SrcToken: ${args.srcToken}`)
            console.log(`DstToken: ${args.dstToken}`)
            console.log(`Merchant: ${args.merchant}`)
            console.log(`Amount:   ${args.amount}`)

            const optionsHex = Options.newOptions().addExecutorLzReceiveOption(150000, 0).toHex()
            console.log(`Options:  ${optionsHex} (gas=${150000})`)

            // Helper: encode address to bytes32 by dstEid
            const toBytes32 = (value: string): string => {
                if (dstEidNum === 40168) {
                    // Solana: base58 -> 32 bytes
                    const bytes = new PublicKey(value).toBytes()
                    if (bytes.length !== 32) throw new Error('Solana pubkey must be 32 bytes')
                    return '0x' + Buffer.from(bytes).toString('hex')
                } else {
                    // EVM: address -> left-padded bytes32
                    return hre.ethers.utils.hexZeroPad(value, 32)
                }
            }

            const dstToken32 = toBytes32(args.dstToken)
            const merchant32 = toBytes32(args.merchant)

            // 1) Quote message fee
            console.log('Quoting LayerZero message fee...')
            const fee = await myOApp.quotePayoutToken(
                dstEidNum,
                dstToken32,
                merchant32,
                args.amount,
                optionsHex,
                false
            )
            console.log(`  Native fee: ${hre.ethers.utils.formatEther(fee.nativeFee)} ETH`)

            // 2) Send request (caller must have approved srcToken to MyOApp for at least amount)
            console.log('Sending requestPayoutToken...')
            const tx: ContractTransaction = await myOApp.requestPayoutToken(
                dstEidNum,
                args.srcToken,
                dstToken32,
                merchant32,
                args.amount,
                optionsHex,
                {
                    value: fee.nativeFee,
                }
            )
            console.log(`  Tx hash: ${tx.hash}`)

            // 3) Wait for confirmation
            const rc = await tx.wait()
            console.log(`Confirmed in block ${rc.blockNumber}`)
        }
    )


