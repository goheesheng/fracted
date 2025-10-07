# Stargate Cross-Chain Bridge Contracts

This project contains three Solidity smart contracts for bridging USDC from Sepolia to Optimism Sepolia using Stargate Finance (LayerZero V2):

1. **StargateBridgeWithFee** - Bridge with 3% fee + swap USDC to ETH on Optimism
2. **SimpleBridge** - Simple bridge with no fees
3. **SimpleBridgeWithFee** - Bridge with 0.03% fee (no swap)

---

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Contract Overview](#contract-overview)
- [Deployment](#deployment)
- [Testing](#testing)
- [Contract Addresses](#contract-addresses)
- [How It Works](#how-it-works)

---

## Prerequisites

- Node.js v18+ and npm
- Sepolia testnet ETH (for gas)
- USDC on Sepolia testnet ([Faucet](https://faucet.circle.com/))
- Private key with funds on both Sepolia and Optimism Sepolia

---

## Installation

```bash
npm install
```

---

## Configuration

Create a `.env` file in the project root:

```env
# RPC URLs
SEPOLIA_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY
OPTIMISM_SEPOLIA_RPC_URL=https://optimism-sepolia.infura.io/v3/YOUR_INFURA_KEY

# Private Key
DEPLOYER_PRIVATE_KEY=your_private_key_here

# Sepolia addresses
USDC_SEPOLIA=0x2F6F07CDcf3588944Bf4C42aC74ff24bF56e7590
STARGATE_POOL_USDC_SEPOLIA=0x4985b8fcEA3659FD801a5b857dA1D00e985863F0

# Optimism Sepolia addresses
OPTIMISM_UNISWAP_ROUTER=0x851116d9223fabed8e56c0e6b8ad0c31d98b3507
STARGATE_ENDPOINT_OPTIMISM=0x6EDCE65403992e310A62460808c4b910D972f10f
COMPOSER_OPTIMISM_ADDRESS=0x6D4B828F526b9f4BF60C2230AD915dc4d6e196e7
OPTIMISTIC_RECIPIENT=0xCFB86C607B09150042C584Ee23308413aB4Dff39

# Deployed contracts (will be updated after deployment)
SIMPLE_BRIDGE_SEPOLIA=0x6e78a726050299Ca99f9031d4d887463eC0fAF61
SIMPLE_BRIDGE_WITH_FEE_SEPOLIA=0x45bdfF18693d79c592A36d7273Ce4F06f140DC61
```

---

## Contract Overview

### 1. StargateBridgeWithFee

**Purpose:** Bridge USDC with 3% fee, swap to ETH on Optimism, send to recipient

**Features:**
- Takes 3% fee on bridged amount
- Bridges remaining 97% via Stargate
- Automatically swaps USDC → WETH → ETH on Optimism using Uniswap V3
- Sends native ETH to recipient
- Owner can withdraw accumulated fees

**Deployed:**
- Sepolia: `0xa3204D24ff21DB7cBDCf932D1592e8D27E9E838A`
- Optimism Sepolia (Receiver): `0x6D4B828F526b9f4BF60C2230AD915dc4d6e196e7`

---

### 2. SimpleBridge

**Purpose:** Simple USDC bridge with no fees or swaps

**Features:**
- No fees
- Direct bridge from Sepolia → Optimism Sepolia
- Recipient receives exact amount (minus LayerZero gas)
- No swaps or conversions

**Deployed:**
- Sepolia: `0x6e78a726050299Ca99f9031d4d887463eC0fAF61`

---

### 3. SimpleBridgeWithFee

**Purpose:** Bridge USDC with minimal 0.03% protocol fee

**Features:**
- Takes 0.03% fee (3 basis points)
- Bridges remaining 99.97% via Stargate
- Fees accumulate in contract
- Anyone can call `withdrawFees()` to send fees to treasury
- No swaps - recipient receives USDC on Optimism

**Deployed:**
- Sepolia: `0x45bdfF18693d79c592A36d7273Ce4F06f140DC61`

---

## Deployment

### Deploy All Contracts

```bash
# 1. Compile contracts
npm run build

# 2. Deploy receiver on Optimism Sepolia (for StargateBridgeWithFee)
node ./scripts/deploy-optimism-raw.mjs

# 3. Update COMPOSER_OPTIMISM_ADDRESS in .env with deployed address

# 4. Deploy StargateBridgeWithFee on Sepolia
node ./scripts/deploy-sepolia-raw.mjs

# 5. Deploy SimpleBridge on Sepolia
node ./scripts/deploy-simple-bridge.mjs

# 6. Deploy SimpleBridgeWithFee on Sepolia
node ./scripts/deploy-simple-bridge-with-fee.mjs
```

---

## Testing

### Test StargateBridgeWithFee (3% fee + swap to ETH)

```bash
# Bridge 10 USDC, swap to ETH on Optimism
node ./scripts/test-bridge.mjs
```

**What happens:**
1. Takes 3% fee (0.3 USDC) → stays in contract
2. Bridges 9.7 USDC to Optimism
3. Swaps 9.7 USDC → WETH on Uniswap V3
4. Unwraps WETH → native ETH
5. Sends ETH to recipient address

---

### Test SimpleBridge (no fees)

```bash
# Bridge 10 USDC with no fees
node ./scripts/test-simple-bridge.mjs
```

**What happens:**
1. No fees taken
2. Bridges full 10 USDC to Optimism
3. Recipient receives 10 USDC (minus LayerZero gas)

---

### Test SimpleBridgeWithFee (0.03% fee)

```bash
# Bridge 10 USDC with 0.03% fee
node ./scripts/test-simple-bridge-with-fee.mjs
```

**What happens:**
1. Takes 0.03% fee (0.003 USDC) → accumulated in contract
2. Bridges 9.997 USDC to Optimism
3. Recipient receives 9.997 USDC

**Withdraw accumulated fees to treasury:**
```javascript
// Anyone can call this to send fees to treasury
await simpleBridgeWithFee.withdrawFees();
```

---

## Contract Addresses

### Sepolia Testnet

| Contract | Address | Purpose |
|----------|---------|---------|
| USDC | `0x2F6F07CDcf3588944Bf4C42aC74ff24bF56e7590` | USDC token |
| Stargate Pool | `0x4985b8fcEA3659FD801a5b857dA1D00e985863F0` | Stargate USDC pool |
| StargateBridgeWithFee | `0xa3204D24ff21DB7cBDCf932D1592e8D27E9E838A` | 3% fee + swap |
| SimpleBridge | `0x6e78a726050299Ca99f9031d4d887463eC0fAF61` | No fees |
| SimpleBridgeWithFee | `0x45bdfF18693d79c592A36d7273Ce4F06f140DC61` | 0.03% fee |

### Optimism Sepolia Testnet

| Contract | Address | Purpose |
|----------|---------|---------|
| SwapReceiverOptimism | `0x6D4B828F526b9f4BF60C2230AD915dc4d6e196e7` | Receives & swaps to ETH |
| WETH | `0x4200000000000000000000000000000000000006` | Canonical WETH |
| USDC/WETH Pool | `0x86e63F9f307891438AdcFcd6FEa865338080848F` | Uniswap V3 pool |

---

## How It Works

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         SEPOLIA TESTNET                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  User → [StargateBridgeWithFee] → Stargate → LayerZero        │
│           ↓ 3% fee                                              │
│           ↓ 97% bridged                                         │
│                                                                 │
│  User → [SimpleBridge] → Stargate → LayerZero                  │
│           ↓ No fees                                             │
│           ↓ 100% bridged                                        │
│                                                                 │
│  User → [SimpleBridgeWithFee] → Stargate → LayerZero          │
│           ↓ 0.03% fee                                           │
│           ↓ 99.97% bridged                                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              ↓
                      LayerZero Bridge
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    OPTIMISM SEPOLIA TESTNET                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  [SwapReceiverOptimism] → Uniswap V3 → WETH → ETH             │
│           ↓ receives USDC                                       │
│           ↓ swaps to WETH                                       │
│           ↓ unwraps to ETH                                      │
│           ↓ sends to recipient                                  │
│                                                                 │
│  Recipient receives USDC (SimpleBridge)                         │
│  Recipient receives USDC (SimpleBridgeWithFee)                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Gas & Fee Breakdown

**StargateBridgeWithFee:**
- Protocol fee: 3% (kept in contract on Sepolia)
- LayerZero gas: ~0.00001 ETH (paid by user)
- Recipient receives: Native ETH on Optimism

**SimpleBridge:**
- Protocol fee: 0%
- LayerZero gas: ~0.000007 ETH (paid by user)
- Recipient receives: USDC on Optimism

**SimpleBridgeWithFee:**
- Protocol fee: 0.03% (accumulated in contract, sent to treasury)
- LayerZero gas: ~0.000007 ETH (paid by user)
- Recipient receives: USDC on Optimism

---

## Bridge Timing

Cross-chain bridging via LayerZero typically takes **1-5 minutes**.

Monitor transactions:
- Sepolia: https://sepolia.etherscan.io
- Optimism Sepolia: https://sepolia-optimism.etherscan.io
- LayerZero Scan: https://testnet.layerzeroscan.com

---

## Withdraw Fees

### StargateBridgeWithFee (3% fee)

```javascript
// Only owner can withdraw
await stargateBridgeWithFee.withdrawFees();
// Sends accumulated USDC fees to owner
```

### SimpleBridgeWithFee (0.03% fee)

```javascript
// Anyone can call to send fees to treasury
await simpleBridgeWithFee.withdrawFees();
// Sends accumulated USDC fees to treasury address
```

---

## Security Notes

⚠️ **Important:**
- Never commit your `.env` file or private keys
- Test with small amounts first
- Ensure sufficient ETH for gas on both chains
- Verify contract addresses before use
- LayerZero bridging is non-reversible

---

## Troubleshooting

**"USDC transfer failed"**
- Ensure you have enough USDC balance
- Approve the bridge contract first

**"Insufficient ETH"**
- Get Sepolia ETH from faucets
- Ensure you have ~0.01 ETH for gas

**"Wrong network"**
- Scripts check chainId automatically
- Verify RPC URLs in `.env`

**Bridge taking too long?**
- Wait 5-10 minutes
- Check LayerZero Scan: https://testnet.layerzeroscan.com

---

## Technology Stack

- **Solidity 0.8.22** - Smart contract language
- **Hardhat** - Development framework
- **LayerZero V2** - Cross-chain messaging
- **Stargate Finance** - Cross-chain bridge
- **Uniswap V3** - DEX for swaps
- **OpenZeppelin** - Security libraries
- **Ethers.js v6** - Ethereum library

---

## License

MIT

---

## Support

For issues or questions:
1. Check LayerZero documentation: https://docs.layerzero.network
2. Check Stargate documentation: https://stargatefi.gitbook.io
3. Review transaction on block explorers
4. Ensure all environment variables are set correctly

---

## Quick Start Commands

```bash
# 1. Install dependencies
npm install

# 2. Compile contracts
npm run build

# 3. Deploy contracts (in order)
node ./scripts/deploy-optimism-raw.mjs          # Deploy receiver
node ./scripts/deploy-sepolia-raw.mjs           # Deploy main bridge
node ./scripts/deploy-simple-bridge.mjs         # Deploy simple bridge
node ./scripts/deploy-simple-bridge-with-fee.mjs # Deploy bridge with 0.03% fee

# 4. Test bridges
node ./scripts/test-bridge.mjs                  # Test 3% fee + swap
node ./scripts/test-simple-bridge.mjs           # Test no fees
node ./scripts/test-simple-bridge-with-fee.mjs  # Test 0.03% fee
```

---

**Happy Bridging! 🌉**
