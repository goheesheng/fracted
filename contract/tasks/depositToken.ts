import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'

task('lz:oapp:depositToken', 'Owner deposits ERC20 token into MyOApp (requires prior approve)')
    .addParam('token', 'ERC20 token address', undefined, types.string)
    .addParam('amount', 'Token amount (wei units)', undefined, types.string)
    .setAction(async (args: { token: string; amount: string }, hre: HardhatRuntimeEnvironment) => {
        const [signer] = await hre.ethers.getSigners()
        const deployment = await hre.deployments.get('MyOApp')
        const myOApp = await hre.ethers.getContractAt('MyOApp', deployment.address, signer)
        const erc20 = await hre.ethers.getContractAt('IERC20', args.token, signer)

        console.log(`Network: ${hre.network.name}`)
        console.log(`Signer:  ${signer.address}`)
        console.log(`MyOApp:  ${deployment.address}`)
        console.log(`Token:   ${args.token}`)
        console.log(`Amount:  ${args.amount}`)

        const allowance = await erc20.allowance(signer.address, deployment.address)
        if (allowance.lt(args.amount)) {
            console.log('Approving token allowance...')
            const approveTx = await erc20.approve(deployment.address, args.amount)
            console.log(`Approve tx: ${approveTx.hash}`)
            await approveTx.wait()
        }

        console.log('Depositing token using ownerDepositToken...')
        const tx = await myOApp.ownerDepositToken(args.token, args.amount)
        console.log(`Tx sent: ${tx.hash}`)
        const rc = await tx.wait()
        console.log(`Confirmed in block ${rc.blockNumber}`)
    })


