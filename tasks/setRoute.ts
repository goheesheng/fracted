import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'

task('lz:oapp:setRoute', 'Set token route: srcToken (local) -> dstToken (remote) for a destination EID')
    .addParam('dstEid', 'Destination endpoint ID', undefined, types.int)
    .addParam('srcToken', 'Source chain token address', undefined, types.string)
    .addParam('dstToken', 'Destination chain token address', undefined, types.string)
    .setAction(async (args: { dstEid: number; srcToken: string; dstToken: string }, hre: HardhatRuntimeEnvironment) => {
        const [signer] = await hre.ethers.getSigners()
        const deployment = await hre.deployments.get('MyOApp')
        const myOApp = await hre.ethers.getContractAt('MyOApp', deployment.address, signer)

        console.log(`Network: ${hre.network.name}`)
        console.log(`Signer:  ${signer.address}`)
        console.log(`MyOApp:  ${deployment.address}`)
        console.log(`Setting route → dstEid=${args.dstEid}, srcToken=${args.srcToken}, dstToken=${args.dstToken}`)

        const tx = await myOApp.setTokenRoute(args.dstEid, args.srcToken, args.dstToken)
        console.log(`Tx sent: ${tx.hash}`)
        const rc = await tx.wait()
        console.log(`Confirmed in block ${rc.blockNumber}`)
    })


