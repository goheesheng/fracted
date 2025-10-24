// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.22;

import { OApp, Origin, MessagingFee } from "@layerzerolabs/oapp-evm/contracts/oapp/OApp.sol";
import { OAppOptionsType3 } from "@layerzerolabs/oapp-evm/contracts/oapp/libs/OAppOptionsType3.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/// @notice Minimal Permit2 allowance-transfer interface (Uniswap Permit2)
interface IAllowanceTransfer {
    struct PermitDetails {
        address token;
        uint160 amount;
        uint48 expiration;
        uint48 nonce;
    }

    struct PermitSingle {
        PermitDetails details;
        address spender;
        uint256 sigDeadline;
    }

    function permit(address owner, PermitSingle calldata permitSingle, bytes calldata signature) external;
    function transferFrom(address from, address to, uint160 amount, address token) external;
}

contract MyOApp is OApp, OAppOptionsType3 {
    using SafeERC20 for IERC20;
    /// @notice Last string received from any remote chain (legacy example)
    string public lastMessage;

    /// @notice Msg type for sending a string (legacy example)
    uint16 public constant SEND = 1;

    /// @notice Msg type for payout requests
    uint16 public constant PAYOUT = 2;

    /// @notice Platform fee in basis points (3%)
    uint16 public constant FEE_BPS = 300; // 300 / 10000 = 3%

    /// @notice Message tag for token payout messages
    uint8 public constant TAG_TOKEN_PAYOUT = 101;

    /// @dev Emitted when a token payout is requested cross-chain
    /// merchant and dstToken are bytes32 to support both EVM (20-byte) and Solana (32-byte) addresses
    event TokenPayoutRequested(
        uint32 indexed dstEid,
        address indexed payer,
        bytes32 indexed merchant,
        address srcToken,
        bytes32 dstToken,
        uint256 grossAmount,
        uint256 netAmount,
        uint256 feeAmount
    );

    /// @dev Emitted when a token payout is executed on the destination chain
    event TokenPayoutExecuted(address indexed merchant, address token, uint256 amount);

    /// @dev Reverts when insufficient msg.value is provided to cover messaging fee
    error InsufficientMsgValue(uint256 provided, uint256 required);

    /// @dev Reverts when the contract lacks liquidity to fulfill a payout
    error InsufficientLiquidity(uint256 requested, uint256 available);

    /// @notice Permit2 contract used for gasless token pulls (Uniswap Permit2)
    address public permit2;

    /// @notice Initialize with Endpoint V2 and owner address
    /// @param _endpoint The local chain's LayerZero Endpoint V2 address
    /// @param _owner    The address permitted to configure this OApp
    constructor(address _endpoint, address _owner) OApp(_endpoint, _owner) Ownable(_owner) {}

    // Removed ETH receive fallback: contract does not accept direct ETH deposits

    // ──────────────────────────────────────────────────────────────────────────────
    // 0. (Optional) Quote business logic
    //
    // Example: Get a quote from the Endpoint for a cost estimate of sending a message.
    // Replace this to mirror your own send business logic.
    // ──────────────────────────────────────────────────────────────────────────────

    /**
     * @notice Quotes the gas needed to pay for the full omnichain transaction in native gas or ZRO token.
     * @param _dstEid Destination chain's endpoint ID.
     * @param _string The string to send.
     * @param _options Message execution options (e.g., for sending gas to destination).
     * @param _payInLzToken Whether to return fee in ZRO token.
     * @return fee A `MessagingFee` struct containing the calculated gas fee in either the native token or ZRO token.
     */
    function quoteSendString(
        uint32 _dstEid,
        string calldata _string,
        bytes calldata _options,
        bool _payInLzToken
    ) public view returns (MessagingFee memory fee) {
        bytes memory _message = abi.encode(_string);
        // combineOptions (from OAppOptionsType3) merges enforced options set by the contract owner
        // with any additional execution options provided by the caller
        fee = _quote(_dstEid, _message, combineOptions(_dstEid, SEND, _options), _payInLzToken);
    }

    // ──────────────────────────────────────────────────────────────────────────────
    // Admin token routes and liquidity for tokens
    // ──────────────────────────────────────────────────────────────────────────────

    /// @notice Owner sets/updates the Permit2 contract address
    function setPermit2(address _permit2) external onlyOwner {
        require(_permit2 != address(0), "permit2=0");
        permit2 = _permit2;
    }

    /// @notice Owner deposits tokens into the contract as liquidity
    function ownerDepositToken(address _token, uint256 _amount) external onlyOwner {
        IERC20(_token).safeTransferFrom(msg.sender, address(this), _amount);
    }

    /// @notice Owner withdraws tokens from the contract
    function ownerWithdrawToken(address _token, address _to, uint256 _amount) external onlyOwner {
        require(_to != address(0), "to=0");
        IERC20(_token).safeTransfer(_to, _amount);
    }

    // ──────────────────────────────────────────────────────────────────────────────
    // 1. Send business logic
    //
    // Example: send a simple string to a remote chain. Replace this with your
    // own state-update logic, then encode whatever data your application needs.
    // ──────────────────────────────────────────────────────────────────────────────

    /// @notice Send a string to a remote OApp on another chain
    /// @param _dstEid   Destination Endpoint ID (uint32)
    /// @param _string  The string to send
    /// @param _options  Execution options for gas on the destination (bytes)
    function sendString(uint32 _dstEid, string calldata _string, bytes calldata _options) external payable {
        // 1. (Optional) Update any local state here.
        //    e.g., record that a message was "sent":
        //    sentCount += 1;

        // 2. Encode any data structures you wish to send into bytes
        //    You can use abi.encode, abi.encodePacked, or directly splice bytes
        //    if you know the format of your data structures
        bytes memory _message = abi.encode(_string);

        // 3. Call OAppSender._lzSend to package and dispatch the cross-chain message
        //    - _dstEid:   remote chain's Endpoint ID
        //    - _message:  ABI-encoded string
        //    - _options:  combined execution options (enforced + caller-provided)
        //    - MessagingFee(msg.value, 0): pay all gas as native token; no ZRO
        //    - payable(msg.sender): refund excess gas to caller
        //
        //    combineOptions (from OAppOptionsType3) merges enforced options set by the contract owner
        //    with any additional execution options provided by the caller
        _lzSend(
            _dstEid,
            _message,
            combineOptions(_dstEid, SEND, _options),
            MessagingFee(msg.value, 0),
            payable(msg.sender)
        );
    }

    // ──────────────────────────────────────────────────────────────────────────────
    /// @notice Computes net amount after fee and the fee amount
    function _netOfFee(uint256 _amount) internal pure returns (uint256 netAmount, uint256 feeAmount) {
        feeAmount = (_amount * FEE_BPS) / 10_000;
        netAmount = _amount - feeAmount;
    }
    

    // ──────────────────────────────────────────────────────────────────────────────
    // 1c. Token-based cross-chain payout (Base USDT -> Arbitrum USDC)
    // ──────────────────────────────────────────────────────────────────────────────

    /**
     * @notice Quotes the fee for token payout message.
     */
    function quotePayoutToken(
        uint32 _dstEid,
        bytes32 _dstToken,
        bytes32 _merchant,
        uint256 _amount,
        bytes calldata _options,
        bool _payInLzToken
    ) public view returns (MessagingFee memory fee) {
        require(_dstToken != bytes32(0), "dst token=0");
        (uint256 netAmount, ) = _netOfFee(_amount);
        // Encode destination token for the receiver to use
        bytes memory _message = abi.encode(TAG_TOKEN_PAYOUT, _dstToken, _merchant, netAmount);
        fee = _quote(_dstEid, _message, combineOptions(_dstEid, PAYOUT, _options), _payInLzToken);
    }

    /**
     * @notice User requests a token-based payout: user provides USDT on Base; Arbitrum pays 97% USDC to merchant.
     * @dev    Caller must approve `sourceStablecoin` to this contract for `_amount` prior to calling.
     */
    function requestPayoutToken(
        uint32 _dstEid,
        address _srcToken,
        bytes32 _dstToken,
        bytes32 _merchant,
        uint256 _amount,
        bytes calldata _options
    ) external payable {
        require(_srcToken != address(0), "src token=0");
        require(_dstToken != bytes32(0), "dst token=0");
        require(_merchant != bytes32(0), "merchant=0");
        require(_amount > 0, "amount=0");

        // Pull source token from user on this chain
        IERC20(_srcToken).safeTransferFrom(msg.sender, address(this), _amount);

        (uint256 netAmount, uint256 feeAmount) = _netOfFee(_amount);
        // Encode destination token explicitly for the receiver
        bytes memory _message = abi.encode(TAG_TOKEN_PAYOUT, _dstToken, _merchant, netAmount);

        // Quote and enforce native messaging fee
        MessagingFee memory fee = _quote(_dstEid, _message, combineOptions(_dstEid, PAYOUT, _options), false);
        uint256 required = fee.nativeFee;
        if (msg.value < required) revert InsufficientMsgValue(msg.value, required);

        _lzSend(
            _dstEid,
            _message,
            combineOptions(_dstEid, PAYOUT, _options),
            MessagingFee(fee.nativeFee, 0),
            payable(msg.sender)
        );

        uint256 surplus = msg.value - required;
        if (surplus > 0) {
            (bool ok, ) = payable(msg.sender).call{ value: surplus }("");
            require(ok, "refund failed");
        }

        emit TokenPayoutRequested(
            _dstEid,
            msg.sender,
            _merchant,
            _srcToken,
            _dstToken,
            _amount,
            netAmount,
            feeAmount
        );
    }

    /**
     * @notice User requests a token-based payout using Permit2 for allowance + transfer in a single tx.
     * @dev    Caller provides a Permit2 signature so no prior ERC20 approve to this contract is needed.
     * @param _dstEid   Destination Endpoint ID
     * @param _dstToken Destination token paid out to merchant on destination chain
     * @param _merchant Merchant address on destination chain
     * @param _amount   Gross amount of source token to pull from caller
     * @param _options  Execution options for the cross-chain message
     * @param _permit   Permit2 PermitSingle struct authorizing this contract (spender) for the amount
     * @param _signature EIP-712 signature for Permit2
     */
    function requestPayoutTokenWithPermit2(
        uint32 _dstEid,
        address _srcToken,
        bytes32 _dstToken,
        bytes32 _merchant,
        uint256 _amount,
        bytes calldata _options,
        IAllowanceTransfer.PermitSingle calldata _permit,
        bytes calldata _signature
    ) external payable {
        require(permit2 != address(0), "permit2 not set");
        require(_srcToken != address(0), "src token=0");
        require(_dstToken != bytes32(0), "dst token=0");
        require(_merchant != bytes32(0), "merchant=0");
        require(_amount > 0, "amount=0");

        // Sanity checks against provided permit
        require(_permit.details.token == _srcToken, "permit token mismatch");
        require(_permit.spender == address(this), "permit spender!=this");
        require(uint256(_permit.details.amount) >= _amount, "permit amount<amount");

        // Consume Permit2 signature to grant allowance to this contract
        IAllowanceTransfer(permit2).permit(msg.sender, _permit, _signature);

        // Pull tokens via Permit2 transferFrom using the permitted amount
        IAllowanceTransfer(permit2).transferFrom(msg.sender, address(this), uint160(_amount), _srcToken);

        // Process the payout using internal function to reduce stack depth
        _processTokenPayout(_dstEid, _srcToken, _dstToken, _merchant, _amount, _options);
    }

    /**
     * @notice Internal function to process token payout to reduce stack depth
     */
    function _processTokenPayout(
        uint32 _dstEid,
        address _srcToken,
        bytes32 _dstToken,
        bytes32 _merchant,
        uint256 _amount,
        bytes calldata _options
    ) internal {
        (uint256 netAmount, uint256 feeAmount) = _netOfFee(_amount);

        bytes memory _message = abi.encode(TAG_TOKEN_PAYOUT, _dstToken, _merchant, netAmount);

        // Quote and enforce native messaging fee
        MessagingFee memory f = _quote(_dstEid, _message, combineOptions(_dstEid, PAYOUT, _options), false);
        uint256 required = f.nativeFee;
        if (msg.value < required) revert InsufficientMsgValue(msg.value, required);

        _lzSend(
            _dstEid,
            _message,
            combineOptions(_dstEid, PAYOUT, _options),
            MessagingFee(f.nativeFee, 0),
            payable(msg.sender)
        );

        uint256 surplus = msg.value - required;
        if (surplus > 0) {
            (bool ok, ) = payable(msg.sender).call{ value: surplus }("");
            require(ok, "refund failed");
        }

        emit TokenPayoutRequested(
            _dstEid,
            msg.sender,
            _merchant,
            _srcToken,
            _dstToken,
            _amount,
            netAmount,
            feeAmount
        );
    }

    // ──────────────────────────────────────────────────────────────────────────────
    // 2. Receive business logic
    //
    // Override _lzReceive to decode the incoming bytes and apply your logic.
    // The base OAppReceiver.lzReceive ensures:
    //   • Only the LayerZero Endpoint can call this method
    //   • The sender is a registered peer (peers[srcEid] == origin.sender)
    // ──────────────────────────────────────────────────────────────────────────────

    /// @notice Invoked by OAppReceiver when EndpointV2.lzReceive is called
    /// @dev   _origin    Metadata (source chain, sender address, nonce)
    /// @dev   _guid      Global unique ID for tracking this message
    /// @param _message   ABI-encoded bytes (the string we sent earlier)
    /// @dev   _executor  Executor address that delivered the message
    /// @dev   _extraData Additional data from the Executor (unused here)
    function _lzReceive(
        Origin calldata /*_origin*/,
        bytes32 /*_guid*/,
        bytes calldata _message,
        address /*_executor*/,
        bytes calldata /*_extraData*/
    ) internal override {
        // Token payout path: (uint8 tag, address dstToken, address merchant, uint256 netAmount)
        if (_message.length == 128) {
            (uint8 tag, bytes32 dstToken32, bytes32 merchant32, uint256 netAmountToken) = abi.decode(_message, (uint8, bytes32, bytes32, uint256));
            if (tag == TAG_TOKEN_PAYOUT) {
                // Convert bytes32 (either EVM or Solana pubkey form) to EVM address for ERC20
                address dstToken = address(uint160(uint256(dstToken32)));
                address merchantToken = address(uint160(uint256(merchant32)));
                IERC20 token = IERC20(dstToken);
                uint256 tokenBal = token.balanceOf(address(this));
                if (tokenBal < netAmountToken) revert InsufficientLiquidity(netAmountToken, tokenBal);
                token.safeTransfer(merchantToken, netAmountToken);
                emit TokenPayoutExecuted(merchantToken, dstToken, netAmountToken);
                return;
            }
        }

        // Legacy example: treat as string message
        string memory _string = abi.decode(_message, (string));
        lastMessage = _string;
    }
}
