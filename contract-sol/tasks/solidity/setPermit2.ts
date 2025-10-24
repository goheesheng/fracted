import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'

// Set the Permit2 address on the deployed MyOApp contract (owner-only)
// Usage:
//   pnpm hardhat lz:oapp:setPermit2 \
//     --address 0x000000000022D473030F116dDEE9F6B43aC78BA3 \
//     --network base-sepolia

task('lz:oapp:setPermit2', 'Owner: set/update the Permit2 address on MyOApp')
  .addParam('address', 'Permit2 contract address', undefined, types.string)
  .setAction(async (args: { address: string }, hre: HardhatRuntimeEnvironment) => {
    const [signer] = await hre.ethers.getSigners()
    const deployment = await hre.deployments.get('MyOApp')
    const myOApp = await hre.ethers.getContractAt('MyOApp', deployment.address, signer)

    console.log(`Network:  ${hre.network.name}`)
    console.log(`Owner:    ${signer.address}`)
    console.log(`MyOApp:   ${deployment.address}`)
    console.log(`New P2:   ${args.address}`)

    const before = await myOApp.permit2()
    console.log(`Before:   ${before}`)

    const tx = await myOApp.setPermit2(args.address)
    console.log(`Tx hash:  ${tx.hash}`)
    const rc = await tx.wait()
    console.log(`Confirmed in block ${rc.blockNumber}`)

    const after = await myOApp.permit2()
    console.log(`After:    ${after}`)
  })
