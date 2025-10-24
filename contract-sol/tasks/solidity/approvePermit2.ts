import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'

task('lz:oapp:approvePermit2', 'Approve Permit2 to spend your ERC20 tokens (one-time setup)')
    .addParam('token', 'ERC20 token address', undefined, types.string)
    .setAction(async (args: { token: string }, hre: HardhatRuntimeEnvironment) => {
        const [signer] = await hre.ethers.getSigners()
        const erc20 = await hre.ethers.getContractAt('IERC20', args.token, signer)

        // Permit2 canonical address (same on all chains)
        const PERMIT2_ADDRESS = '0x000000000022D473030F116dDEE9F6B43aC78BA3'

        console.log(`Network: ${hre.network.name}`)
        console.log(`Signer:  ${signer.address}`)
        console.log(`Token:   ${args.token}`)
        console.log(`Permit2: ${PERMIT2_ADDRESS}`)

        // Check current allowance
        const currentAllowance = await erc20.allowance(signer.address, PERMIT2_ADDRESS)
        console.log(`\nCurrent allowance: ${hre.ethers.utils.formatUnits(currentAllowance, 6)}`)

        if (currentAllowance.gt(0)) {
            console.log('✅ Token already approved for Permit2')
            console.log('   You can use meta-transactions now!')
            return
        }

        console.log('\n⚠️  Approving Permit2 for unlimited amount...')
        console.log('   This is a one-time operation per token.')

        const tx = await erc20.approve(PERMIT2_ADDRESS, hre.ethers.constants.MaxUint256)
        console.log(`Approve tx: ${tx.hash}`)
        
        const rc = await tx.wait()
        console.log(`✅ Confirmed in block ${rc.blockNumber}`)
        console.log('\n🎉 Success! You can now use gasless meta-transactions.')
    })

