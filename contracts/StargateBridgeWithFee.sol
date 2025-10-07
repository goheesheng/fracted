// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IOFT, SendParam, MessagingFee} from "@layerzerolabs/lz-evm-oapp-v2/contracts/oft/interfaces/IOFT.sol";
import {OptionsBuilder} from "@layerzerolabs/lz-evm-oapp-v2/contracts/oapp/libs/OptionsBuilder.sol";

contract StargateBridgeWithFee {
    using OptionsBuilder for bytes;

    // Addresses on Sepolia side
    IERC20 public immutable usdcSepolia;
    IOFT public immutable stargateSepolia; // Stargate OFT (USDC) on Sepolia

    // LayerZero endpoint IDs
    uint32 public constant SEP_CHAIN_ID = 40161; // Sepolia endpoint ID (for reference)
    uint32 public constant OPT_CHAIN_ID = 40232; // Optimism Sepolia endpoint ID

    // Destination config
    address public immutable composerOptimism; // Receiver contract on Optimism that executes the swap
    address public immutable optimisticRecipient; // Final recipient of USDT on Optimism

    address public owner;
    uint256 public accumulatedFee; // USDC fees accumulated on Sepolia

    event SwapAndBridge(address indexed sender, uint256 amount, uint256 feeAmount, uint256 amountAfterFee);
    event FeesWithdrawn(address indexed to, uint256 amount);

    constructor(
        address _usdcSepolia,
        address _stargateSepolia,
        address _composerOptimism,
        address _optimisticRecipient
    ) {
        require(_usdcSepolia != address(0), "usdc addr");
        require(_stargateSepolia != address(0), "stargate addr");
        require(_composerOptimism != address(0), "composer addr");
        require(_optimisticRecipient != address(0), "recipient addr");

        usdcSepolia = IERC20(_usdcSepolia);
        stargateSepolia = IOFT(_stargateSepolia);
        composerOptimism = _composerOptimism;
        optimisticRecipient = _optimisticRecipient;
        owner = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        _;
    }

    // ========== USER FUNCTION on Sepolia ==========
    function swapAndBridge(
        uint256 amount,
        address oftTokenOnDst,
        address usdtOnDst,
        uint256 minAmountOut,
        uint256 deadline
    ) external payable {
        require(amount > 0, "Amount must be > 0");

        // Pull USDC from user
        require(usdcSepolia.transferFrom(msg.sender, address(this), amount), "USDC transfer failed");

        // Take 3% fee
        uint256 feeAmount = (amount * 3) / 100;
        accumulatedFee += feeAmount;
        uint256 amountAfterFee = amount - feeAmount;

        // Approve Stargate to pull USDC for bridging
        require(usdcSepolia.approve(address(stargateSepolia), amountAfterFee), "Approve failed");

        // Compose message for Optimism receiver contract to swap USDC -> USDT via Uniswap
        bytes memory composeMsg = abi.encode(
            optimisticRecipient, // recipient of USDT on optimism
            oftTokenOnDst, // OFT (USDC) token address on destination
            usdtOnDst, // USDT token on optimism
            minAmountOut, // min amount out for swap
            deadline // swap deadline
        );

        // Build extra options: enable compose with gas limit
        bytes memory extraOptions = OptionsBuilder.newOptions().addExecutorLzComposeOption(0, 200000, 0);

        // Fill SendParam struct
        SendParam memory sendParam = SendParam({
            dstEid: OPT_CHAIN_ID,
            to: bytes32(uint256(uint160(composerOptimism))),
            amountLD: amountAfterFee,
            minAmountLD: amountAfterFee,
            extraOptions: extraOptions,
            composeMsg: composeMsg,
            oftCmd: ""
        });

        // Quote fees
        MessagingFee memory messagingFee = stargateSepolia.quoteSend(sendParam, false);
        uint256 nativeFee = messagingFee.nativeFee;
        require(msg.value >= nativeFee, "Insufficient native fee");

        // Send with value for messaging fees; refund any excess to sender
        stargateSepolia.send{value: nativeFee}(sendParam, messagingFee, msg.sender);
        if (msg.value > nativeFee) {
            (bool refundOk,) = msg.sender.call{value: msg.value - nativeFee}("");
            require(refundOk, "Refund failed");
        }

        emit SwapAndBridge(msg.sender, amount, feeAmount, amountAfterFee);
    }

    // Owner can withdraw accumulated fees (USDC)
    function withdrawFees(address to) external onlyOwner {
        require(to != address(0), "bad to");
        uint256 amount = accumulatedFee;
        require(amount > 0, "No fees");
        accumulatedFee = 0;
        require(usdcSepolia.transfer(to, amount), "Fee transfer failed");
        emit FeesWithdrawn(to, amount);
    }
}


