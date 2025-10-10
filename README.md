# Fracted - Cross-Chain Payment System

A decentralized cross-chain payment platform built on LayerZero V2 that enables seamless token transfers between Base Sepolia and Arbitrum Sepolia testnets with a 3% platform fee.

## 🌟 Overview

Fracted is a cross-chain payment system that allows users to:
- Send tokens from one chain (e.g., USDT on Base Sepolia)
- Receive different tokens on another chain (e.g., USDC on Arbitrum Sepolia)
- Pay merchants across different blockchains with minimal friction
- Generate payment links for easy integration

## 📋 Table of Contents

- [Architecture](#architecture)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Smart Contracts](#smart-contracts)
- [Payment API](#payment-api)
- [Deployment](#deployment)
- [Usage](#usage)
- [Demo Website](#demo-website)
- [Contract Addresses](#contract-addresses)
- [Development](#development)
- [Security](#security)
- [License](#license)

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        BASE SEPOLIA                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  User → [MyOApp] → LayerZero V2 → Cross-Chain Message         │
│           ↓ 3% fee                                              │
│           ↓ 97% to merchant                                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              ↓
                      LayerZero Bridge
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      ARBITRUM SEPOLIA                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  [MyOApp] → Receives Message → Pays Merchant in USDC          │
│           ↓ processes payout                                    │
│           ↓ transfers tokens                                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## ✨ Features

### Core Functionality
- **Cross-Chain Token Payouts**: Send tokens on one chain, receive different tokens on another
- **Platform Fee**: 3% fee on all transactions (configurable)
- **Multi-Token Support**: USDT, USDC, and custom tokens
- **Merchant Payments**: Direct payments to merchant addresses
- **Payment Links**: Generate shareable payment links

### Technical Features
- **LayerZero V2**: Latest cross-chain messaging protocol
- **Gas Optimization**: Efficient cross-chain message handling
- **Liquidity Management**: Contract-based liquidity pools
- **Admin Controls**: Owner-managed token deposits and withdrawals
- **Event Logging**: Comprehensive transaction tracking

## 🔧 Prerequisites

- Node.js v18+ and npm/pnpm
- Base Sepolia testnet ETH (for gas)
- Arbitrum Sepolia testnet ETH (for gas)
- Test tokens (USDT/USDC) on both networks
- Private key with funds on both chains

## 📦 Installation

```bash
# Clone the repository
git clone <repository-url>
cd fracted

# Install contract dependencies
cd contract
pnpm install

# Install payment API dependencies
cd ../paymentapi
npm install
```

## ⚙️ Configuration

### Contract Configuration

Create a `.env` file in the `contract/` directory:

```env
# RPC URLs
BASE_SEPOLIA_RPC_URL=https://sepolia.base.org
ARBITRUM_SEPOLIA_RPC_URL=https://sepolia-rollup.arbitrum.io/rpc

# Private Key
DEPLOYER_PRIVATE_KEY=your_private_key_here

# LayerZero Endpoints
LAYERZERO_ENDPOINT_BASE_SEPOLIA=0x6EDCE65403992e310A62460808c4b910D972f10f
LAYERZERO_ENDPOINT_ARBITRUM_SEPOLIA=0x6EDCE65403992e310A62460808c4b910D972f10f

# Token Addresses
USDT_BASE_SEPOLIA=0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673
USDC_BASE_SEPOLIA=0x036CbD53842c5426634e7929541eC2318f3dCF7e
USDT_ARBITRUM_SEPOLIA=0x30fA2FbE15c1EaDfbEF28C188b7B8dbd3c1Ff2eB
USDC_ARBITRUM_SEPOLIA=0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d
```

### Payment API Configuration

Create a `.env` file in the `paymentapi/` directory:

```env
# Contract Addresses
OAPP_base_sepolia=0x5C5254f25C24eC1dFb612067AB6CbD15E6430071
OAPP_arbitrum_sepolia=0x0cfE9BdF5C027623C44991fE5Ca493A93B62bD27

# Token Addresses
TOKEN_base_sepolia_USDT=0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673
TOKEN_base_sepolia_USDC=0x036CbD53842c5426634e7929541eC2318f3dCF7e
TOKEN_arbitrum_sepolia_USDT=0x30fA2FbE15c1EaDfbEF28C188b7B8dbd3c1Ff2eB
TOKEN_arbitrum_sepolia_USDC=0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d

# Chain Mappings
EID_TO_CHAIN_40245=base-sepolia
EID_TO_CHAIN_40231=arbitrum-sepolia
```

## 📄 Smart Contracts

### MyOApp Contract

The main contract implementing cross-chain token payouts:

**Key Functions:**
- `requestPayoutToken()`: Initiate cross-chain token payout
- `quotePayoutToken()`: Get fee quote for payout
- `ownerDepositToken()`: Admin function to add liquidity
- `ownerWithdrawToken()`: Admin function to withdraw tokens

**Features:**
- 3% platform fee on all transactions
- Cross-chain message handling via LayerZero V2
- Automatic token transfers to merchants
- Liquidity management for payouts

## 🌐 Payment API

### Server Endpoints

- `GET /config`: Returns contract and token configuration
- `GET /options?gas=150000`: Generates LayerZero options for gas
- `GET /generate-link`: Creates payment links with parameters
- `GET /`: Serves the payment interface

### Usage Examples

```bash
# Generate payment link
curl "http://localhost:8080/generate-link?merchant=0x...&dstEid=40231&dstToken=0x...&amount=1000000"

# Get LayerZero options
curl "http://localhost:8080/options?gas=150000"
```

## 🚀 Deployment

### 1. Deploy Smart Contracts

```bash
cd contract

# Compile contracts
pnpm run compile:hardhat

# Deploy to Base Sepolia
npx hardhat deploy --network base-sepolia

# Deploy to Arbitrum Sepolia
npx hardhat deploy --network arbitrum-sepolia

# Configure LayerZero connections
pnpm hardhat lz:oapp:wire --oapp-config layerzero.config.ts
```

### 2. Add Liquidity

```bash
# Approve tokens for the contract
npx hardhat lz:oapp:approveToken --network base-sepolia --token 0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673 --amount 1000000000000000000

# Deposit tokens as liquidity
npx hardhat lz:oapp:depositToken --network arbitrum-sepolia --token 0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d --amount 10000
```

### 3. Start Payment API

```bash
cd paymentapi
npm start
```

## 💡 Usage

### Making a Payment

1. **Generate Payment Link**:
   ```bash
   curl "http://localhost:8080/generate-link?merchant=0xB7aa464b19037CF3dB7F723504dFafE7b63aAb84&dstEid=40231&dstToken=0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d&amount=1000000"
   ```

2. **User Pays**: User visits the payment link and connects their wallet

3. **Cross-Chain Transfer**: System automatically:
   - Takes 3% fee on source chain
   - Sends 97% to destination chain
   - Pays merchant in requested token

### Direct Contract Interaction

```javascript
// Request payout from Base to Arbitrum
await myOApp.requestPayoutToken(
  40231, // Arbitrum Sepolia EID
  "0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673", // USDT on Base
  "0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d", // USDC on Arbitrum
  "0xB7aa464b19037CF3dB7F723504dFafE7b63aAb84", // Merchant address
  1000000, // Amount (6 decimals)
  "0x" // Options
);
```

## 🎨 Demo Website

The project includes a demo website (`DemoWebsite/`) showcasing the payment interface:

- **index.html**: Main payment interface
- **generator.html**: Payment link generator
- **style.css**: Styling and responsive design

Access the demo at: `http://localhost:8080`

## 📍 Contract Addresses

### Base Sepolia Testnet
- **MyOApp**: `0x5C5254f25C24eC1dFb612067AB6CbD15E6430071`
- **USDT**: `0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673`
- **USDC**: `0x036CbD53842c5426634e7929541eC2318f3dCF7e`

### Arbitrum Sepolia Testnet
- **MyOApp**: `0x0cfE9BdF5C027623C44991fE5Ca493A93B62bD27`
- **USDT**: `0x30fA2FbE15c1EaDfbEF28C188b7B8dbd3c1Ff2eB`
- **USDC**: `0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d`

## 🛠️ Development

### Testing

```bash
cd contract

# Run Hardhat tests
pnpm run test:hardhat

# Run Foundry tests
pnpm run test:forge
```

### Linting

```bash
# Lint JavaScript/TypeScript
pnpm run lint:js

# Lint Solidity
pnpm run lint:sol

# Fix linting issues
pnpm run lint:fix
```

### Available Tasks

```bash
# Approve tokens
npx hardhat lz:oapp:approveToken --network base-sepolia --token <address> --amount <amount>

# Deposit liquidity
npx hardhat lz:oapp:depositToken --network arbitrum-sepolia --token <address> --amount <amount>

# Request payout
npx hardhat lz:oapp:requestPayoutToken --network base-sepolia --dst-eid 40231 --src-token <address> --merchant <address> --amount <amount>

# Send test string
npx hardhat lz:oapp:sendString --network base-sepolia --dst-eid 40231 --string "Hello Cross-Chain!"
```

## 🔒 Security

### Important Security Notes

⚠️ **Critical Security Considerations:**
- Never commit private keys or `.env` files
- Test with small amounts on testnets first
- Ensure sufficient liquidity before processing payouts
- Verify contract addresses before deployment
- Monitor for insufficient liquidity errors
- Cross-chain transactions are non-reversible

### Best Practices

- Use multi-signature wallets for admin functions
- Implement proper access controls
- Regular security audits
- Monitor contract events for anomalies
- Keep LayerZero configurations updated

## 📊 Fee Structure

- **Platform Fee**: 3% (300 basis points)
- **LayerZero Gas**: Variable (paid by user)
- **Net Amount**: 97% of original amount reaches merchant

## 🔗 Technology Stack

- **Solidity 0.8.22**: Smart contract language
- **LayerZero V2**: Cross-chain messaging protocol
- **Hardhat**: Development framework
- **Foundry**: Testing framework
- **Express.js**: Payment API server
- **OpenZeppelin**: Security libraries
- **Ethers.js v5**: Ethereum interaction library

## 📄 License

MIT License - see LICENSE file for details

## 🆘 Support

For issues or questions:

1. Check LayerZero documentation: https://docs.layerzero.network
2. Review transaction logs on block explorers
3. Verify environment configuration
4. Ensure sufficient liquidity in contracts

## 🚀 Quick Start

```bash
# 1. Install dependencies
cd contract && pnpm install
cd ../paymentapi && npm install

# 2. Configure environment variables
cp .env.example .env  # Edit with your values

# 3. Deploy contracts
cd contract
pnpm run compile:hardhat
npx hardhat deploy --network base-sepolia
npx hardhat deploy --network arbitrum-sepolia
pnpm hardhat lz:oapp:wire --oapp-config layerzero.config.ts

# 4. Add liquidity
npx hardhat lz:oapp:depositToken --network arbitrum-sepolia --token <USDC_ADDRESS> --amount 10000

# 5. Start payment API
cd ../paymentapi
npm start

# 6. Access demo at http://localhost:8080
```

---

**Built with ❤️ using LayerZero V2**

*Fracted - Bridging payments across chains*