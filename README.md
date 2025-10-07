## Stargate Taxi Composable Swap (Sepolia -> Optimism Sepolia)

This Hardhat project deploys:
- `StargateBridgeWithFee` on Sepolia to bridge USDC via Stargate with a 3% fee and compose a swap message.
- `SwapReceiverOptimism` on Optimism Sepolia to receive bridged USDC and swap to USDT via Uniswap, sending to your wallet.

### Prerequisites
- Node.js 18+ (Hardhat warns on Node 22.2; prefer latest LTS)
- Testnet funds on Sepolia and Optimism Sepolia for gas

### Install
```bash
npm install
```

If peer conflicts arise, you can install with legacy peer deps:
```bash
npm install --legacy-peer-deps
```

### Configure
Copy `.env.example` to `.env` and fill values:
```
SEPOLIA_RPC_URL=
OPTIMISM_SEPOLIA_RPC_URL=
DEPLOYER_PRIVATE_KEY=

USDC_SEPOLIA=0x2F6F07CDcf3588944Bf4C42aC74ff24bF56e7590
STARGATE_POOL_USDC_SEPOLIA=<Stargate Pool USDC on Sepolia>

OPTIMISM_UNISWAP_ROUTER=<Uniswap V2 Router on Optimism Sepolia>
STARGATE_ENDPOINT_OPTIMISM=0x1a44076050125825900e736c501f859c50fe728c
COMPOSER_OPTIMISM_ADDRESS=<filled after deploying receiver>
OPTIMISTIC_RECIPIENT=0xCFB86C607B09150042C584Ee23308413aB4Dff39

This is the StargateBridgeWithFee.sol
https://sepolia.etherscan.io/address/0xbb36a68ba27a84d1f91d0ffe5c9c61dfcdbbb42b#code
```

Note: Endpoint IDs used in contracts: Sepolia=40161, Optimism Sepolia=40232.

### Compile
```bash
npx hardhat compile
```

### Deploy
1) Deploy receiver to Optimism Sepolia:
```bash
npx hardhat run scripts/deploy-optimism.ts --network optimismSepolia
```
Then set the deployed address into `.env` as `COMPOSER_OPTIMISM_ADDRESS=<deployed SwapReceiverOptimism address>`.

2) Deploy bridge to Sepolia:
```bash
npx hardhat run scripts/deploy-sepolia.ts --network sepolia
```

### Use: swap and bridge
On Sepolia, call `swapAndBridge(amount, oftTokenOnDst, usdtOnDst, minAmountOut, deadline)` and send the LayerZero native fee in `msg.value`:
- `oftTokenOnDst`: USDC OFT address on Optimism Sepolia
- `usdtOnDst`: USDT token on Optimism Sepolia
- `minAmountOut`: slippage control
- `deadline`: unix timestamp

Approve the bridge contract to spend your USDC first.

### Security notes
- Contracts assume Stargate v2 interfaces and compose format; verify addresses on testnet.
- Owner can withdraw accumulated USDC fees on Sepolia.


