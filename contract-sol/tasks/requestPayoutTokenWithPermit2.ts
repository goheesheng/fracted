import { ContractTransaction } from 'ethers'
import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'
import { endpointIdToNetwork } from '@layerzerolabs/lz-definitions'
import { Options } from '@layerzerolabs/lz-v2-utilities'

// EIP-712 types for Uniswap Permit2 (PermitSingle)
const PermitDetails = [
  { name: 'token', type: 'address' },
  { name: 'amount', type: 'uint160' },
  { name: 'expiration', type: 'uint48' },
  { name: 'nonce', type: 'uint48' },
]

const PermitSingle = [
  { name: 'details', type: 'PermitDetails' },
  { name: 'spender', type: 'address' },
  { name: 'sigDeadline', type: 'uint256' },
]

task(
  'lz:oapp:requestPayoutTokenWithPermit2',
  'Request token payout via Permit2 (no prior ERC20 approve needed)'
)
  .addParam('dstEid', 'Destination endpoint ID', undefined, types.int)
  .addParam('srcToken', 'Source chain token address to pull from caller', undefined, types.string)
  .addParam('dstToken', 'Destination chain token address to pay to merchant', undefined, types.string)
  .addParam('merchant', 'Destination chain merchant address to receive payout', undefined, types.string)
  .addParam(
    'amount',
    'Gross token amount in smallest units (e.g., 1 USDC = 1000000)',
    undefined,
    types.string
  )
  .addOptionalParam('deadline', 'Signature deadline (unix seconds)', undefined, types.int)
  .addOptionalParam('expiration', 'Permit2 allowance expiration (unix seconds)', undefined, types.int)
  .addOptionalParam('nonce', 'Permit2 nonce (uint48); default uses current timestamp', undefined, types.int)
  .setAction(
    async (
      args: {
        dstEid: number
        srcToken: string
        dstToken: string
        merchant: string
        amount: string
        deadline?: number
        expiration?: number
        nonce?: number
      },
      hre: HardhatRuntimeEnvironment
    ) => {
      const [signer] = await hre.ethers.getSigners()
      const deployment = await hre.deployments.get('MyOApp')
      const myOApp = await hre.ethers.getContractAt('MyOApp', deployment.address, signer)

      const chainId = (await hre.ethers.provider.getNetwork()).chainId
      const permit2Address: string = await myOApp.permit2()
      if (permit2Address === hre.ethers.constants.AddressZero) {
        throw new Error('Permit2 address not set on MyOApp. Run task lz:oapp:setPermit2 first.')
      }

      console.log(`Initiating token payout (Permit2) from ${hre.network.name} -> ${endpointIdToNetwork(args.dstEid)}`)
      console.log(`MyOApp:   ${deployment.address}`)
      console.log(`Signer:   ${signer.address}`)
      console.log(`Permit2:  ${permit2Address}`)
      console.log(`SrcToken: ${args.srcToken}`)
      console.log(`DstToken: ${args.dstToken}`)
      console.log(`Merchant: ${args.merchant}`)
      console.log(`Amount:   ${args.amount}`)

      const optionsHex = Options.newOptions().addExecutorLzReceiveOption(150000, 0).toHex()
      console.log(`Options:  ${optionsHex} (gas=${150000})`)

      // 1) Quote message fee
      console.log('Quoting LayerZero message fee...')
      const fee = await myOApp.quotePayoutToken(
        args.dstEid,
        args.dstToken,
        args.merchant,
        args.amount,
        optionsHex,
        false
      )
      console.log(`  Native fee: ${hre.ethers.utils.formatEther(fee.nativeFee)} ETH`)

      // 2) Get current nonce from Permit2 allowance
      const permit2Contract = new hre.ethers.Contract(permit2Address, [
        {
          "inputs": [
            {"internalType": "address", "name": "owner", "type": "address"},
            {"internalType": "address", "name": "token", "type": "address"},
            {"internalType": "address", "name": "spender", "type": "address"}
          ],
          "name": "allowance",
          "outputs": [
            {
              "components": [
                {"internalType": "uint160", "name": "amount", "type": "uint160"},
                {"internalType": "uint32", "name": "expiration", "type": "uint32"},
                {"internalType": "uint32", "name": "nonce", "type": "uint32"}
              ],
              "internalType": "struct AllowanceDetails",
              "name": "",
              "type": "tuple"
            }
          ],
          "stateMutability": "view",
          "type": "function"
        }
      ], signer)

      const allowance = await permit2Contract.allowance(signer.address, args.srcToken, deployment.address)
      const currentNonce = allowance.nonce

      // 3) Build Permit2 PermitSingle and sign EIP-712
      const now = Math.floor(Date.now() / 1000)
      const sigDeadline = args.deadline ?? now + 60 * 10 // 10 minutes default
      const expiration = args.expiration ?? now + 60 * 60 // 1 hour default
      const nonce = args.nonce ?? currentNonce // use current nonce from Permit2

      // amount must fit uint160
      const amountBN = hre.ethers.BigNumber.from(args.amount)
      const maxUint160 = hre.ethers.BigNumber.from('0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF') // 160 bits
      if (amountBN.gt(maxUint160)) {
        throw new Error('amount does not fit into uint160 for Permit2')
      }

      const permit = {
        details: {
          token: args.srcToken,
          amount: amountBN.toHexString(),
          expiration,
          nonce,
        },
        spender: deployment.address,
        sigDeadline,
      }

      const domain = {
        name: 'Permit2',
        chainId: Number(chainId),
        verifyingContract: permit2Address,
      }

      const types = {
        PermitDetails,
        PermitSingle,
      }

      console.log('Signing Permit2 EIP-712...')
      const signature = await (signer as any)._signTypedData(domain, types as any, permit)
      console.log(`  Signature: ${signature}`)

      // 3) Call requestPayoutTokenWithPermit2
      console.log('Sending requestPayoutTokenWithPermit2...')
      const tx: ContractTransaction = await myOApp.requestPayoutTokenWithPermit2(
        args.dstEid,
        args.srcToken,
        args.dstToken,
        args.merchant,
        args.amount,
        optionsHex,
        permit,
        signature,
        { value: fee.nativeFee }
      )
      console.log(`  Tx hash: ${tx.hash}`)

      const rc = await tx.wait()
      console.log(`Confirmed in block ${rc.blockNumber}`)
    }
  )
