import { ContractTransaction } from 'ethers'
import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'
import { endpointIdToNetwork } from '@layerzerolabs/lz-definitions'
import { Options } from '@layerzerolabs/lz-v2-utilities'

task(
    'lz:oapp:requestPayoutToken',
    'Request token payout cross-chain (pull srcToken, pay dstToken 97% to merchant)'
)
    .addParam('dstEid', 'Destination endpoint ID', undefined, types.int)
    .addParam('srcToken', 'Source chain token address to pull from caller', undefined, types.string)
    .addParam('merchant', 'Destination chain merchant address to receive payout', undefined, types.string)
    .addParam(
        'amount',
        'Gross token amount in smallest units (e.g., 1 USDC = 1000000)',
        undefined,
        types.string
    )
    .addOptionalParam('options', 'Execution options (hex string)', '0x', types.string)
    .addOptionalParam('gas', 'LZ_RECEIVE gas for msgType=2 (used if options is 0x)', 150000, types.int)
    .setAction(
        async (
            args: {
                dstEid: number
                srcToken: string
                merchant: string
                amount: string
                options?: string
                gas?: number
            },
            hre: HardhatRuntimeEnvironment
        ) => {
            const [signer] = await hre.ethers.getSigners()
            const deployment = await hre.deployments.get('MyOApp')
            const myOApp = await hre.ethers.getContractAt('MyOApp', deployment.address, signer)

            console.log(`Initiating token payout from ${hre.network.name} to ${endpointIdToNetwork(args.dstEid)}`)
            console.log(`MyOApp:   ${deployment.address}`)
            console.log(`Signer:   ${signer.address}`)
            console.log(`SrcToken: ${args.srcToken}`)
            console.log(`Merchant: ${args.merchant}`)
            console.log(`Amount:   ${args.amount}`)

            // Build options if not provided
            const options = args.options || '0x'
            const optionsHex =
                options !== '0x'
                    ? options
                    : Options.newOptions().addExecutorLzReceiveOption(Number(args.gas), 0).toHex()
            console.log(`Options:  ${optionsHex} (gas=${args.gas})`)

            // 1) Quote message fee
            console.log('Quoting LayerZero message fee...')
            const fee = await myOApp.quotePayoutToken(
                args.dstEid,
                args.srcToken,
                args.merchant,
                args.amount,
                optionsHex,
                false
            )
            console.log(`  Native fee: ${hre.ethers.utils.formatEther(fee.nativeFee)} ETH`)

            // 2) Send request (caller must have approved srcToken to MyOApp for at least amount)
            console.log('Sending requestPayoutToken...')
            const tx: ContractTransaction = await myOApp.requestPayoutToken(
                args.dstEid,
                args.srcToken,
                args.merchant,
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


