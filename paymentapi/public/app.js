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

// Token decimals mapping (most stablecoins use 6 decimals, but some use 18)
const TOKEN_DECIMALS = {
  'USDT': 6,
  'USDC': 6,
  'XUSD': 18
}

// EID to Chain Name mapping (loaded from server)
let EID_TO_CHAIN = {}

// Token address to symbol mapping (loaded from server)
let TOKEN_ADDRESS_TO_SYMBOL = {}

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
  
  console.log('setFormFromQuery called, loadedConfig:', loadedConfig)
  
  // Update amount display after setting form values
  updateAmountDisplay()
  
  // Update display names if config is already loaded
  if (loadedConfig) {
    updateDisplayNames()
  }
}

function resolveSrcTokenAddress() {
  const net = $('srcNetwork').value
  const sym = $('srcTokenSelect').value
  const addr = (TOKEN_ADDRESSES[net] && TOKEN_ADDRESSES[net][sym]) || ''
  return addr
}

function formatAmountToUSD(amountWei, tokenSymbol) {
  if (!amountWei || amountWei === '0') return '$0.00'
  
  const decimals = TOKEN_DECIMALS[tokenSymbol] || 6
  const amount = ethers.utils.formatUnits(amountWei, decimals)
  const usdAmount = parseFloat(amount).toFixed(2)
  return `$${usdAmount}`
}

function updateAmountDisplay() {
  const amountWei = $('amount').value
  const dstToken = $('dstToken').value
  
  // Get token symbol from server mapping
  const tokenSymbol = TOKEN_ADDRESS_TO_SYMBOL[dstToken] || 'USDC'
  
  const usdAmount = formatAmountToUSD(amountWei, tokenSymbol)
  
  // Update amount display
  $('amountDisplay').textContent = usdAmount
  
  // Update amount details
  $('amountDetails').textContent = `${amountWei} ${tokenSymbol} (smallest units)`
}

function updateDisplayNames() {
  const merchant = $('merchant').value
  const dstEid = $('dstEid').value
  const dstToken = $('dstToken').value
  const amount = $('amount').value
  
  console.log('updateDisplayNames called with:', { merchant, dstEid, dstToken, amount })
  console.log('EID_TO_CHAIN:', EID_TO_CHAIN)
  console.log('TOKEN_ADDRESS_TO_SYMBOL:', TOKEN_ADDRESS_TO_SYMBOL)
  
  // Update merchant display
  $('merchantDisplay').textContent = merchant || 'Loading...'
  
  // Update EID display - show friendly name but keep original value for processing
  const chainName = EID_TO_CHAIN[dstEid] || `EID ${dstEid}`
  console.log('Chain name for EID', dstEid, ':', chainName)
  $('dstEidDisplay').textContent = chainName
  // Store original EID in a data attribute and show friendly name
  $('dstEid').setAttribute('data-original-eid', dstEid)
  $('dstEid').value = chainName
  
  // Update token display - show friendly name but keep original value for processing
  const tokenSymbol = TOKEN_ADDRESS_TO_SYMBOL[dstToken] || 'Unknown Token'
  console.log('Token symbol for address', dstToken, ':', tokenSymbol)
  $('dstTokenDisplay').textContent = tokenSymbol
  // Store original token address in a data attribute and show friendly name
  $('dstToken').setAttribute('data-original-token', dstToken)
  $('dstToken').value = tokenSymbol
  
  // Update amount raw display
  $('amountRawDisplay').textContent = amount || 'Loading...'
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

// Mobile flow state
let isMobileFlow = false
let currentPage = 'order' // 'order' or 'payment'

async function loadConfigAndApply() {
  try {
    const res = await fetch('/config')
    if (res.ok) {
      const cfg = await res.json()
      if (cfg?.tokens) TOKEN_ADDRESSES = { ...TOKEN_ADDRESSES, ...cfg.tokens }
      if (cfg?.contracts) OAPP_ADDRESSES = { ...OAPP_ADDRESSES, ...cfg.contracts }
      if (cfg?.eidToChain) EID_TO_CHAIN = { ...EID_TO_CHAIN, ...cfg.eidToChain }
      if (cfg?.tokenToSymbol) TOKEN_ADDRESS_TO_SYMBOL = { ...TOKEN_ADDRESS_TO_SYMBOL, ...cfg.tokenToSymbol }
      loadedConfig = true
      
      // Debug: Log loaded mappings
      console.log('Loaded EID_TO_CHAIN:', EID_TO_CHAIN)
      console.log('Loaded TOKEN_ADDRESS_TO_SYMBOL:', TOKEN_ADDRESS_TO_SYMBOL)
      
      // Update display names after loading config from server
      updateDisplayNames()
    }
  } catch (e) {
    console.error('Failed to load config:', e)
  }
}

// Show loading state
function setButtonLoading(loading) {
  const payBtn = document.getElementById('payBtn')
  
  if (loading) {
    payBtn.disabled = true
    payBtn.textContent = 'Processing...'
    payBtn.style.opacity = '0.7'
  } else {
    payBtn.disabled = false
    payBtn.textContent = 'Pay'
    payBtn.style.opacity = '1'
  }
}

window.addEventListener('load', async () => {
  setFormFromQuery()
  await loadConfigAndApply()

  // Check MetaMask status on page load
  function checkMetaMaskStatus() {
    if (isMobile()) {
      if (!isMetaMaskInstalled()) {
        $('walletInfo').textContent = 'Tap to open MetaMask app'
        log('Mobile detected without MetaMask extension')
      } else {
        $('walletInfo').textContent = 'MetaMask browser extension detected'
        log('Mobile detected with MetaMask extension')
      }
    } else {
      if (!isMetaMaskInstalled()) {
        $('walletInfo').textContent = 'MetaMask not found. Please install MetaMask extension.'
        log('Desktop without MetaMask extension')
      } else {
        $('walletInfo').textContent = 'MetaMask detected. Click to connect.'
        log('Desktop with MetaMask extension')
      }
    }
  }

  // Initialize MetaMask status check
  checkMetaMaskStatus()

  // Initialize mobile flow
  initMobileFlow()

  // React to network changes
  $('srcNetwork').addEventListener('change', () => {})

  // Confirm order button event listener
  $('confirmOrderBtn').addEventListener('click', () => {
    log('Order confirmed, switching to payment page')
    switchToPaymentPage()
  })

  // Back to order button event listener
  $('backToOrderBtn').addEventListener('click', () => {
    log('Back to order page')
    switchToOrderPage()
  })

  // Mobile detection function
  function isMobile() {
    return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
  }

  // Mobile flow functions
  function initMobileFlow() {
    if (isMobile()) {
      isMobileFlow = true
      const container = document.querySelector('.payment-container')
      container.classList.add('mobile-flow')
      currentPage = 'order'
      log('Mobile flow initialized')
    }
  }

  function switchToPaymentPage() {
    if (!isMobileFlow) return
    
    const container = document.querySelector('.payment-container')
    container.classList.add('show-payment')
    currentPage = 'payment'
    log('Switched to payment page')
  }

  function switchToOrderPage() {
    if (!isMobileFlow) return
    
    const container = document.querySelector('.payment-container')
    container.classList.remove('show-payment')
    currentPage = 'order'
    log('Switched to order page')
  }

  // Check if MetaMask is installed
  function isMetaMaskInstalled() {
    return typeof window.ethereum !== 'undefined' && window.ethereum.isMetaMask
  }

  // Open MetaMask app on mobile
  function openMetaMaskApp() {
    const currentUrl = window.location.href
    const metamaskUrl = 'metamask://dapp/' + window.location.host + window.location.pathname + window.location.search
    const metamaskUniversalUrl = 'https://metamask.app.link/dapp/' + window.location.host + window.location.pathname + window.location.search
    
    log('Attempting to open MetaMask app...')
    log('MetaMask URL: ' + metamaskUrl)
    log('Universal URL: ' + metamaskUniversalUrl)
    
    // Try to open MetaMask app directly
    const iframe = document.createElement('iframe')
    iframe.style.display = 'none'
    iframe.src = metamaskUrl
    document.body.appendChild(iframe)
    
    // Fallback: open universal link after a short delay
    setTimeout(() => {
      document.body.removeChild(iframe)
      window.open(metamaskUniversalUrl, '_blank')
    }, 2000)
  }

  $('connectBtn').addEventListener('click', async () => {
    try {
      // Check if we're on mobile
      if (isMobile()) {
        if (!isMetaMaskInstalled()) {
          log('Mobile detected, opening MetaMask app...')
          $('walletInfo').textContent = 'Opening MetaMask app...'
          openMetaMaskApp()
          return
        } else {
          log('Mobile detected with MetaMask browser extension')
        }
      }

      // Desktop or MetaMask already available
      if (!window.ethereum) {
        if (isMobile()) {
          log('MetaMask not found on mobile, opening app...')
          $('walletInfo').textContent = 'Opening MetaMask app...'
          openMetaMaskApp()
          return
        } else {
          throw new Error('MetaMask not found. Please install MetaMask extension.')
        }
      }
      
      log('Requesting account access...')
      $('walletInfo').textContent = 'Connecting...'
      
      await window.ethereum.request({ method: 'eth_requestAccounts' })
      provider = new ethers.providers.Web3Provider(window.ethereum)
      signer = provider.getSigner()
      const addr = await signer.getAddress()
      $('walletInfo').textContent = `Connected: ${addr.slice(0, 6)}...${addr.slice(-4)}`
      log('Wallet connected successfully')
    } catch (e) {
      log(`Connect error: ${e.message || e}`)
      if (e.code === 4001) {
        $('walletInfo').textContent = 'Connection rejected by user'
      } else {
        $('walletInfo').textContent = `Error: ${e.message || e}`
      }
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
      setButtonLoading(true)
      
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

      // Get the original EID number from data attribute or current value
      let dstEid
      const originalEid = $('dstEid').getAttribute('data-original-eid')
      if (originalEid) {
        // Use the original EID number stored in data attribute
        dstEid = Number(originalEid)
      } else if (dstEidStr.startsWith('EID ')) {
        // Extract number from "EID 12345" format
        dstEid = Number(dstEidStr.replace('EID ', ''))
      } else {
        // Try to find EID by chain name (reverse lookup)
        const foundEid = Object.keys(EID_TO_CHAIN).find(eid => EID_TO_CHAIN[eid] === dstEidStr)
        if (foundEid) {
          dstEid = Number(foundEid)
        } else {
          // Try direct conversion
          dstEid = Number(dstEidStr)
        }
      }

      // Validate dstEid is a valid number
      if (isNaN(dstEid) || dstEid <= 0) {
        throw new Error(`Invalid destination EID: ${dstEidStr}. Expected a positive number or valid chain name.`)
      }

      // Get the original token address from data attribute or current value
      let dstTokenAddress = dstToken
      const originalToken = $('dstToken').getAttribute('data-original-token')
      if (originalToken) {
        dstTokenAddress = originalToken
      }

      const myOApp = new ethers.Contract(oapp, MYOAPP_ABI, signer)

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
        dstTokenAddress,
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
        dstTokenAddress,
        merchant,
        amount,
        optionsHex,
        { value: fee.nativeFee }
      )
      log(`Tx: ${tx.hash}`)
      const rc = await tx.wait()
      log(`Confirmed in block ${rc.blockNumber}`)
      
      // Show success message
      alert('Payment successful! Transaction confirmed.')
      
    } catch (e) {
      log(`Pay error: ${e.message || e}`)
      alert(`Payment failed: ${e.message || e}`)
    } finally {
      setButtonLoading(false)
    }
  })
})
