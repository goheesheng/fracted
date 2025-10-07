// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { IOFT, SendParam, MessagingFee } from "@layerzerolabs/lz-evm-oapp-v2/contracts/oft/interfaces/IOFT.sol";
import { OptionsBuilder } from "@layerzerolabs/lz-evm-oapp-v2/contracts/oapp/libs/OptionsBuilder.sol";

/**
 * @title SimpleBridgeWithFee
 * @notice Bridge USDC from Sepolia to Optimism Sepolia with 0.03% fee
 */
contract SimpleBridgeWithFee {
    using OptionsBuilder for bytes;

    IERC20 public immutable usdc;
    IOFT public immutable stargatePool;
    
    uint32 public constant OPTIMISM_SEPOLIA_EID = 40232;
    uint256 public constant FEE_BASIS_POINTS = 3; // 0.03% = 3 basis points out of 10000
    uint256 public constant BASIS_POINTS_DIVISOR = 10000;
    
    address public treasury;
    address public owner;
    uint256 public accumulatedFees;
    
    event Bridged(address indexed sender, address indexed recipient, uint256 amount, uint256 fee, uint32 dstEid);
    event FeesWithdrawn(address indexed treasury, uint256 amount);
    event TreasuryUpdated(address indexed oldTreasury, address indexed newTreasury);

    constructor(address _usdc, address _stargatePool, address _treasury) {
        require(_usdc != address(0) && _stargatePool != address(0) && _treasury != address(0), "Invalid address");
        usdc = IERC20(_usdc);
        stargatePool = IOFT(_stargatePool);
        treasury = _treasury;
        owner = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        _;
    }

    /**
     * @notice Bridge USDC to Optimism Sepolia with 0.03% fee
     * @param amount Amount of USDC to bridge (6 decimals) - fee will be deducted
     * @param recipient Address to receive USDC on Optimism Sepolia
     */
    function bridge(uint256 amount, address recipient) external payable {
        require(amount > 0, "Amount must be > 0");
        require(recipient != address(0), "Invalid recipient");

        // Transfer USDC from user to this contract
        require(usdc.transferFrom(msg.sender, address(this), amount), "USDC transfer failed");

        // Calculate 0.03% fee
        uint256 feeAmount = (amount * FEE_BASIS_POINTS) / BASIS_POINTS_DIVISOR;
        accumulatedFees += feeAmount;
        uint256 amountAfterFee = amount - feeAmount;

        // Approve Stargate Pool to spend USDC (only the amount after fee)
        require(usdc.approve(address(stargatePool), amountAfterFee), "Approve failed");

        // Build send parameters
        bytes memory extraOptions = OptionsBuilder.newOptions().addExecutorLzReceiveOption(50000, 0);

        SendParam memory sendParam = SendParam({
            dstEid: OPTIMISM_SEPOLIA_EID,
            to: bytes32(uint256(uint160(recipient))),
            amountLD: amountAfterFee,
            minAmountLD: amountAfterFee,
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

        emit Bridged(msg.sender, recipient, amountAfterFee, feeAmount, OPTIMISM_SEPOLIA_EID);
    }

    /**
     * @notice Quote the messaging fee for bridging
     * @param amount Amount of USDC to bridge (before fee deduction)
     * @param recipient Recipient address on destination chain
     * @return nativeFee Amount of native token needed for LayerZero fees
     * @return protocolFee Amount of USDC fee (0.03%)
     * @return amountAfterFee Amount that will be bridged after fee
     */
    function quoteBridge(uint256 amount, address recipient) external view returns (
        uint256 nativeFee,
        uint256 protocolFee,
        uint256 amountAfterFee
    ) {
        protocolFee = (amount * FEE_BASIS_POINTS) / BASIS_POINTS_DIVISOR;
        amountAfterFee = amount - protocolFee;
        
        bytes memory extraOptions = OptionsBuilder.newOptions().addExecutorLzReceiveOption(50000, 0);

        SendParam memory sendParam = SendParam({
            dstEid: OPTIMISM_SEPOLIA_EID,
            to: bytes32(uint256(uint160(recipient))),
            amountLD: amountAfterFee,
            minAmountLD: amountAfterFee,
            extraOptions: extraOptions,
            composeMsg: "",
            oftCmd: ""
        });

        MessagingFee memory messagingFee = stargatePool.quoteSend(sendParam, false);
        nativeFee = messagingFee.nativeFee;
    }

    /**
     * @notice Withdraw accumulated fees to treasury
     */
    function withdrawFees() external {
        require(accumulatedFees > 0, "No fees to withdraw");
        uint256 amount = accumulatedFees;
        accumulatedFees = 0;
        require(usdc.transfer(treasury, amount), "Transfer failed");
        emit FeesWithdrawn(treasury, amount);
    }

    /**
     * @notice Update treasury address (owner only)
     */
    function updateTreasury(address newTreasury) external onlyOwner {
        require(newTreasury != address(0), "Invalid treasury");
        address oldTreasury = treasury;
        treasury = newTreasury;
        emit TreasuryUpdated(oldTreasury, newTreasury);
    }

    /**
     * @notice Transfer ownership (owner only)
     */
    function transferOwnership(address newOwner) external onlyOwner {
        require(newOwner != address(0), "Invalid owner");
        owner = newOwner;
    }
}

