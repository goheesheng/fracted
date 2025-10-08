import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'

task('lz:oapp:getRoute', 'Get dstToken for given (dstEid, srcToken)')
    .addParam('dstEid', 'Destination endpoint ID', undefined, types.int)
    .addParam('srcToken', 'Source chain token address', undefined, types.string)
    .setAction(async (args: { dstEid: number; srcToken: string }, hre: HardhatRuntimeEnvironment) => {
        const [signer] = await hre.ethers.getSigners()
        const deployment = await hre.deployments.get('MyOApp')
        const myOApp = await hre.ethers.getContractAt('MyOApp', deployment.address, signer)

        console.log(`Network: ${hre.network.name}`)
        console.log(`MyOApp:  ${deployment.address}`)
        const dstToken = await myOApp.dstTokenByDstEidAndSrcToken(args.dstEid, args.srcToken)
        console.log(`Route: (dstEid=${args.dstEid}, srcToken=${args.srcToken}) -> dstToken=${dstToken}`)
    })


