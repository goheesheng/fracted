/* Fracted Payment Frontend */

// ====== Configuration ======
// Chain IDs
const CHAINS = {
  'arbitrum-sepolia': {
    chainId: '0x66EED', // 421614
    chainName: 'Arbitrum Sepolia',
    nativeCurrency: { name: 'ETH', symbol: 'ETH', decimals: 18 },
    rpcUrls: ['https://sepolia-rollup.arbitrum.io/rpc'],
    blockExplorerUrls: ['https://sepolia.arbiscan.io/'],
  },
  'base-sepolia': {
    chainId: '0x14A34', // 84532
    chainName: 'Base Sepolia',
    nativeCurrency: { name: 'ETH', symbol: 'ETH', decimals: 18 },
    rpcUrls: ['https://sepolia.base.org'],
    blockExplorerUrls: ['https://sepolia.basescan.org/'],
  },
}

// Default token addresses per testnet (can be overridden by server /config)
let TOKEN_ADDRESSES = {
  'arbitrum-sepolia': {
    USDT: '',
    USDC: '0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d',
    XUSD: '',
  },
  'base-sepolia': {
    USDT: '0x323e78f944A9a1FcF3a10efcC5319DBb0bB6e673',
    USDC: '',
    XUSD: '',
  },
}

// MyOApp contract addresses per network, loaded from server /config
let OAPP_ADDRESSES = {
  'arbitrum-sepolia': '',
  'base-sepolia': '',
}

// Minimal ABIs
const ERC20_ABI = [
  'function approve(address spender, uint256 amount) external returns (bool)',
  'function allowance(address owner, address spender) external view returns (uint256)',
]

const MYOAPP_ABI = [
  'function quotePayoutToken(uint32 _dstEid, address _dstToken, address _merchant, uint256 _amount, bytes _options, bool _payInLzToken) view returns (tuple(uint256 nativeFee, uint256 lzTokenFee))',
  'function requestPayoutToken(uint32 _dstEid, address _srcToken, address _dstToken, address _merchant, uint256 _amount, bytes _options) payable',
]

// ====== Helpers ======
const $ = (id) => document.getElementById(id)
function log(msg) {
  const el = $('log')
  el.textContent += `\n${msg}`
  el.scrollTop = el.scrollHeight
}

function getQueryParam(name, defaultValue = '') {
  const url = new URL(window.location.href)
  return url.searchParams.get(name) ?? defaultValue
}

function setFormFromQuery() {
  $('merchant').value = getQueryParam('merchant')
  $('dstEid').value = getQueryParam('dstEid')
  $('dstToken').value = getQueryParam('dstToken')
  $('amount').value = getQueryParam('amount')
}

function resolveSrcTokenAddress() {
  const net = $('srcNetwork').value
  const sym = $('srcTokenSelect').value
  const addr = (TOKEN_ADDRESSES[net] && TOKEN_ADDRESSES[net][sym]) || ''
  return addr
}

async function ensureNetwork(targetKey) {
  if (!window.ethereum) throw new Error('MetaMask not found')
  const target = CHAINS[targetKey]
  const current = await window.ethereum.request({ method: 'eth_chainId' })
  if (current.toLowerCase() === target.chainId.toLowerCase()) return
  try {
    await window.ethereum.request({
      method: 'wallet_switchEthereumChain',
      params: [{ chainId: target.chainId }],
    })
  } catch (switchError) {
    if (switchError.code === 4902) {
      await window.ethereum.request({
        method: 'wallet_addEthereumChain',
        params: [target],
      })
    } else {
      throw switchError
    }
  }
}

// ====== Main ======
let provider, signer
let loadedConfig = false

async function loadConfigAndApply() {
  try {
    const res = await fetch('/config')
    if (res.ok) {
      const cfg = await res.json()
      if (cfg?.tokens) TOKEN_ADDRESSES = { ...TOKEN_ADDRESSES, ...cfg.tokens }
      if (cfg?.contracts) OAPP_ADDRESSES = { ...OAPP_ADDRESSES, ...cfg.contracts }
      loadedConfig = true
      // Nothing to set in UI; OAPP address will be resolved at runtime from OAPP_ADDRESSES
    }
  } catch (e) {
    // ignore, fallback to defaults
  }
}

window.addEventListener('load', () => {
  setFormFromQuery()
  loadConfigAndApply()

  // React to network changes if needed (no UI field now)
  $('srcNetwork').addEventListener('change', () => {})

  $('connectBtn').addEventListener('click', async () => {
    try {
      if (!window.ethereum) throw new Error('MetaMask not found')
      await window.ethereum.request({ method: 'eth_requestAccounts' })
      provider = new ethers.providers.Web3Provider(window.ethereum)
      signer = provider.getSigner()
      const addr = await signer.getAddress()
      $('walletInfo').textContent = `Connected: ${addr}`
      log('Wallet connected')
    } catch (e) {
      log(`Connect error: ${e.message || e}`)
    }
  })

  $('approveBtn').addEventListener('click', async () => {
    try {
      const networkKey = $('srcNetwork').value
      await ensureNetwork(networkKey)
      if (!provider) {
        provider = new ethers.providers.Web3Provider(window.ethereum)
        signer = provider.getSigner()
      }
      const srcTokenAddr = resolveSrcTokenAddress()
      if (!srcTokenAddr) {
        alert('No source token address configured for the selected network/token. Please set TOKEN_ADDRESSES in app.js.')
        return
      }
      const oapp = OAPP_ADDRESSES[networkKey]
      if (!oapp) {
        alert(`Missing OApp address for ${networkKey}. Please set OAPP_${networkKey.replace('-', '_')} in .env.`)
        return
      }
      const amount = $('amount').value.trim()
      if (!amount) throw new Error('amount missing')

      const erc20 = new ethers.Contract(srcTokenAddr, ERC20_ABI, signer)
      const tx = await erc20.approve(oapp, amount)
      log(`Approve tx: ${tx.hash}`)
      const rc = await tx.wait()
      log(`Approve confirmed in block ${rc.blockNumber}`)
    } catch (e) {
      log(`Approve error: ${e.message || e}`)
    }
  })

  $('payBtn').addEventListener('click', async () => {
    try {
      const networkKey = $('srcNetwork').value
      await ensureNetwork(networkKey)
      if (!provider) {
        provider = new ethers.providers.Web3Provider(window.ethereum)
        signer = provider.getSigner()
      }

      const oapp = OAPP_ADDRESSES[networkKey]
      const merchant = $('merchant').value.trim()
      const dstEidStr = $('dstEid').value.trim()
      const dstToken = $('dstToken').value.trim()
      const amount = $('amount').value.trim()
      const srcToken = resolveSrcTokenAddress()

      if (!oapp || !merchant || !dstEidStr || !dstToken || !amount || !srcToken) {
        throw new Error('Missing required fields (oappAddress, merchant, dstEid, dstToken, amount, srcToken)')
      }

      const myOApp = new ethers.Contract(oapp, MYOAPP_ABI, signer)
      const dstEid = Number(dstEidStr)

      // Build options hex via backend (mirror hardhat task behavior)
      log('Building LayerZero options (executor receive gas)...')
      const gas = 150000
      const optionsResp = await fetch(`/options?gas=${gas}`)
      if (!optionsResp.ok) throw new Error(`Options API error: ${optionsResp.status}`)
      const { optionsHex } = await optionsResp.json()
      log(`Options: ${optionsHex} (gas=${gas})`)

      log('Quoting fee...')
      const fee = await myOApp.quotePayoutToken(
        dstEid,
        dstToken,
        merchant,
        amount,
        optionsHex,
        false
      )
      log(`Fee.native: ${ethers.utils.formatEther(fee.nativeFee)} ETH`)

      log('Sending requestPayoutToken...')
      const tx = await myOApp.requestPayoutToken(
        dstEid,
        srcToken,
        dstToken,
        merchant,
        amount,
        optionsHex,
        { value: fee.nativeFee }
      )
      log(`Tx: ${tx.hash}`)
      const rc = await tx.wait()
      log(`Confirmed in block ${rc.blockNumber}`)
    } catch (e) {
      log(`Pay error: ${e.message || e}`)
    }
  })
})
