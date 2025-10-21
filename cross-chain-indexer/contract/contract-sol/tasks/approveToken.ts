import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'

task('lz:oapp:approveToken', 'Approve MyOApp to spend your ERC20 tokens')
    .addParam('token', 'ERC20 token address', undefined, types.string)
    .addParam('amount', 'Approve amount (wei units)', undefined, types.string)
    .setAction(async (args: { token: string; amount: string }, hre: HardhatRuntimeEnvironment) => {
        const [signer] = await hre.ethers.getSigners()
        const deployment = await hre.deployments.get('MyOApp')
        const erc20 = await hre.ethers.getContractAt('IERC20', args.token, signer)

        console.log(`Network: ${hre.network.name}`)
        console.log(`Signer:  ${signer.address}`)
        console.log(`Token:   ${args.token}`)
        console.log(`Spender: ${deployment.address}`)
        console.log(`Amount:  ${args.amount}`)

        const tx = await erc20.approve(deployment.address, args.amount)
        console.log(`Approve tx: ${tx.hash}`)
        const rc = await tx.wait()
        console.log(`Confirmed in block ${rc.blockNumber}`)
    })


