// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { IOFT, SendParam, MessagingFee } from "@layerzerolabs/lz-evm-oapp-v2/contracts/oft/interfaces/IOFT.sol";
import { OptionsBuilder } from "@layerzerolabs/lz-evm-oapp-v2/contracts/oapp/libs/OptionsBuilder.sol";

/**
 * @title SimpleBridge
 * @notice Simple contract to bridge USDC from Sepolia to Optimism Sepolia
 * No fees, no swaps - just bridging
 */
contract SimpleBridge {
    using OptionsBuilder for bytes;

    IERC20 public immutable usdc;
    IOFT public immutable stargatePool;
    
    uint32 public constant OPTIMISM_SEPOLIA_EID = 40232;
    
    event Bridged(address indexed sender, address indexed recipient, uint256 amount, uint32 dstEid);

    constructor(address _usdc, address _stargatePool) {
        require(_usdc != address(0) && _stargatePool != address(0), "Invalid address");
        usdc = IERC20(_usdc);
        stargatePool = IOFT(_stargatePool);
    }

    /**
     * @notice Bridge USDC to Optimism Sepolia
     * @param amount Amount of USDC to bridge (6 decimals)
     * @param recipient Address to receive USDC on Optimism Sepolia
     */
    function bridge(uint256 amount, address recipient) external payable {
        require(amount > 0, "Amount must be > 0");
        require(recipient != address(0), "Invalid recipient");

        // Transfer USDC from user to this contract
        require(usdc.transferFrom(msg.sender, address(this), amount), "USDC transfer failed");

        // Approve Stargate Pool to spend USDC
        require(usdc.approve(address(stargatePool), amount), "Approve failed");

        // Build send parameters
        bytes memory extraOptions = OptionsBuilder.newOptions().addExecutorLzReceiveOption(50000, 0);

        SendParam memory sendParam = SendParam({
            dstEid: OPTIMISM_SEPOLIA_EID,
            to: bytes32(uint256(uint160(recipient))),
            amountLD: amount,
            minAmountLD: amount,
            extraOptions: extraOptions,
            composeMsg: "",
            oftCmd: ""
        });

        // Get messaging fees
        MessagingFee memory messagingFee = stargatePool.quoteSend(sendParam, false);

        // Require user sent enough native token for fees
        require(msg.value >= messagingFee.nativeFee, "Insufficient fee");

        // Send tokens
        stargatePool.send{value: messagingFee.nativeFee}(sendParam, messagingFee, msg.sender);

        // Refund excess fee if any
        if (msg.value > messagingFee.nativeFee) {
            (bool success, ) = payable(msg.sender).call{value: msg.value - messagingFee.nativeFee}("");
            require(success, "Refund failed");
        }

        emit Bridged(msg.sender, recipient, amount, OPTIMISM_SEPOLIA_EID);
    }

    /**
     * @notice Quote the messaging fee for bridging
     * @param amount Amount of USDC to bridge
     * @param recipient Recipient address on destination chain
     * @return nativeFee Amount of native token needed for LayerZero fees
     */
    function quoteBridge(uint256 amount, address recipient) external view returns (uint256 nativeFee) {
        bytes memory extraOptions = OptionsBuilder.newOptions().addExecutorLzReceiveOption(50000, 0);

        SendParam memory sendParam = SendParam({
            dstEid: OPTIMISM_SEPOLIA_EID,
            to: bytes32(uint256(uint160(recipient))),
            amountLD: amount,
            minAmountLD: amount,
            extraOptions: extraOptions,
            composeMsg: "",
            oftCmd: ""
        });

        MessagingFee memory messagingFee = stargatePool.quoteSend(sendParam, false);
        return messagingFee.nativeFee;
    }
}

