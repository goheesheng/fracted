// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {ILayerZeroComposer} from "@layerzerolabs/lz-evm-protocol-v2/contracts/interfaces/ILayerZeroComposer.sol";
import {OFTComposeMsgCodec} from "@layerzerolabs/lz-evm-oapp-v2/contracts/oft/libs/OFTComposeMsgCodec.sol";
import {IUniswapV3Pool} from "./interfaces/IUniswapV3Pool.sol";
import {IWETH} from "./interfaces/IWETH.sol";

contract SwapReceiverOptimism is ILayerZeroComposer {
    address public constant USDC_WETH_POOL = 0x86e63F9f307891438AdcFcd6FEa865338080848F;
    address public immutable stargateEndpoint;

    event ReceivedAndSwapped(address recipient, uint256 amountIn, uint256 minOut);
    
    // Uniswap V3 callback - called during swap to transfer tokens
    function uniswapV3SwapCallback(
        int256 amount0Delta,
        int256 amount1Delta,
        bytes calldata /* data */
    ) external {
        require(msg.sender == USDC_WETH_POOL, "Invalid caller");
        
        // Determine which token to pay (positive delta means we owe the pool)
        if (amount0Delta > 0) {
            IUniswapV3Pool pool = IUniswapV3Pool(USDC_WETH_POOL);
            address token0 = pool.token0();
            IERC20(token0).transfer(USDC_WETH_POOL, uint256(amount0Delta));
        } else if (amount1Delta > 0) {
            IUniswapV3Pool pool = IUniswapV3Pool(USDC_WETH_POOL);
            address token1 = pool.token1();
            IERC20(token1).transfer(USDC_WETH_POOL, uint256(amount1Delta));
        }
    }

    constructor(address /* _router */, address _endpoint) {
        require(_endpoint != address(0), "bad args");
        stargateEndpoint = _endpoint;
    }

    function lzCompose(
        address /* _from */,
        bytes32 /* _guid */,
        bytes calldata _message,
        address /* _executor */,
        bytes calldata /* _extraData */
    ) external payable override {
        require(msg.sender == stargateEndpoint, "Unauthorized");

        uint256 amountLD = OFTComposeMsgCodec.amountLD(_message);
        bytes memory composeMsg = OFTComposeMsgCodec.composeMsg(_message);

        (
            address recipient,
            address tokenIn,
            address tokenOut,
            uint256 amountOutMin,
            uint256 deadline
        ) = abi.decode(composeMsg, (address, address, address, uint256, uint256));

        if (tokenOut == address(0)) revert("tokenOut must be set");

        // Swap USDC -> WETH via direct pool call (no approval needed; callback transfers)
        IUniswapV3Pool pool = IUniswapV3Pool(USDC_WETH_POOL);
        address token0 = pool.token0();
        bool zeroForOne = (tokenIn == token0); // true if USDC is token0
        
        // Swap exact input; negative amount means exact input
        pool.swap(
            address(this),
            zeroForOne,
            int256(amountLD),
            zeroForOne ? 4295128740 : 1461446703485210103287273052203988822378723970341, // sqrtPriceLimit
            ""
        );
        // After swap, forward tokenOut to recipient; if WETH, unwrap to ETH first
        uint256 bal = IERC20(tokenOut).balanceOf(address(this));
        if (bal > 0) {
            // WETH on Optimism is 0x4200... canonical
            // If the tokenOut is WETH, unwrap and send native ETH
            if (tokenOut == address(0x4200000000000000000000000000000000000006)) {
                IWETH(tokenOut).withdraw(bal);
                (bool s, ) = payable(recipient).call{value: bal}("");
                require(s, "eth send fail");
            } else {
                IERC20(tokenOut).transfer(recipient, bal);
            }
        }

        emit ReceivedAndSwapped(recipient, amountLD, amountOutMin);
    }
}


