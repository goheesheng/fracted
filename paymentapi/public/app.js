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
  'solana-devnet': {
    chainId: '0x103', // 259 (Solana Devnet)
    chainName: 'Solana Devnet',
    nativeCurrency: { name: 'SOL', symbol: 'SOL', decimals: 9 },
    rpcUrls: ['https://api.devnet.solana.com'],
    blockExplorerUrls: ['https://explorer.solana.com/?cluster=devnet'],
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
  'solana-devnet': {
    USDT: 'Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB', // USDT on Solana (example)
    USDC: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v', // USDC on Solana (example)
    XUSD: '', // Custom token address will be added here
  },
}

// MyOApp contract addresses per network, loaded from server /config
let OAPP_ADDRESSES = {
  'arbitrum-sepolia': '',
  'base-sepolia': '',
  'solana-devnet': '', // Solana program address will be added here
}

// Token decimals mapping (most stablecoins use 6 decimals, but some use 18)
// Note: Solana SPL tokens typically use 6 decimals for stablecoins
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

function getPaymentIdFromUrl() {
  const path = window.location.pathname
  // Match both positive and negative numbers, and also handle very large numbers
  const match = path.match(/\/payment\/(-?\d+)/)
  return match ? match[1] : null
}

async function setFormFromQuery() {
  const paymentId = getPaymentIdFromUrl()
  
  if (paymentId) {
    // Load payment data from API using payment ID
    try {
      log(`Loading payment data for ID: ${paymentId}`)
      const response = await fetch(`/api/payment/${paymentId}`)
      
      if (!response.ok) {
        throw new Error(`Failed to load payment: ${response.status}`)
      }
      
      const data = await response.json()
      const payment = data.payment
      
      // Set form values from API data
      $('merchant').value = payment.merchant
      $('dstEid').value = payment.dstEid
      $('dstToken').value = payment.dstToken
      $('amount').value = payment.amount
      
      log(`Payment data loaded: ${JSON.stringify(payment)}`)
      
    } catch (error) {
      log(`Error loading payment data: ${error.message}`)
      alert(`Error loading payment data: ${error.message}`)
      return
    }
  } else {
    // Legacy mode: load from URL parameters
    $('merchant').value = getQueryParam('merchant')
    $('dstEid').value = getQueryParam('dstEid')
    $('dstToken').value = getQueryParam('dstToken')
    $('amount').value = getQueryParam('amount')
    
    log('Using legacy URL parameters')
  }
  
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
  // Handle Solana network
  if (targetKey === 'solana-devnet') {
    if (!window.solana || !window.solana.isPhantom) {
      throw new Error('Phantom wallet not found. Please install Phantom wallet for Solana.')
    }
    
    try {
      // Connect to Phantom wallet
      const response = await window.solana.connect()
      log(`Connected to Phantom wallet: ${response.publicKey.toString()}`)
      
      // Switch to devnet
      await window.solana.request({
        method: 'sol_requestAccounts',
        params: {
          onlyIfTrusted: false
        }
      })
      
      // Set network to devnet
      await window.solana.request({
        method: 'sol_requestAccounts',
        params: {
          onlyIfTrusted: false
        }
      })
      
      log('Switched to Solana devnet')
      return
    } catch (error) {
      throw new Error(`Failed to connect to Solana: ${error.message}`)
    }
  }
  
  // Handle Ethereum networks
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
  // 强制清除缓存
  if ('caches' in window) {
    const cacheNames = await caches.keys()
    await Promise.all(cacheNames.map(cacheName => caches.delete(cacheName)))
  }
  
  await setFormFromQuery()
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
  $('srcNetwork').addEventListener('change', () => {
    const selectedNetwork = $('srcNetwork').value
    if (selectedNetwork === 'solana-devnet') {
      $('walletInfo').textContent = 'Solana Devnet selected. Please install Phantom wallet for Solana support.'
      log('Solana Devnet selected - Phantom wallet required')
    } else {
      // Reset wallet info for Ethereum networks
      if (isMobile()) {
        if (!isMetaMaskInstalled()) {
          $('walletInfo').textContent = 'Tap to open MetaMask app'
        } else {
          $('walletInfo').textContent = 'MetaMask browser extension detected'
        }
      } else {
        if (!isMetaMaskInstalled()) {
          $('walletInfo').textContent = 'MetaMask not found. Please install MetaMask extension.'
        } else {
          $('walletInfo').textContent = 'MetaMask detected. Click to connect.'
        }
      }
    }
  })

  // Confirm order button event listener (mobile only)
  $('confirmOrderBtn').addEventListener('click', () => {
    if (isMobileFlow) {
      log('Order confirmed, switching to payment page')
      switchToPaymentPage()
    }
  })

  // Back to order button event listener (mobile only)
  $('backToOrderBtn').addEventListener('click', () => {
    if (isMobileFlow) {
      log('Back to order page')
      switchToOrderPage()
    }
  })

  // Mobile detection function
  function isMobile() {
    const mobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
    const isSmallScreen = window.innerWidth <= 768
    log(`Mobile detection: userAgent=${mobile}, screenWidth=${window.innerWidth}, isSmallScreen=${isSmallScreen}`)
    return mobile || isSmallScreen
  }

  // Mobile flow functions
  function initMobileFlow() {
    if (isMobile()) {
      isMobileFlow = true
      const container = document.querySelector('.payment-container')
      container.classList.add('mobile-flow')
      currentPage = 'order'
      log('Mobile flow initialized - showing order page first')
    } else {
      log('Desktop detected - showing both sides simultaneously')
    }
  }

  function switchToPaymentPage() {
    if (!isMobileFlow) return
    
    const container = document.querySelector('.payment-container')
    container.classList.add('show-payment')
    currentPage = 'payment'
    log('Switched to payment page - hiding order, showing payment')
  }

  function switchToOrderPage() {
    if (!isMobileFlow) return
    
    const container = document.querySelector('.payment-container')
    container.classList.remove('show-payment')
    currentPage = 'order'
    log('Switched to order page - showing order, hiding payment')
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
      const selectedNetwork = $('srcNetwork').value
      
      // Handle Solana network
      if (selectedNetwork === 'solana-devnet') {
        if (!window.solana || !window.solana.isPhantom) {
          throw new Error('Phantom wallet not found. Please install Phantom wallet for Solana.')
        }
        
        log('Connecting to Phantom wallet...')
        $('walletInfo').textContent = 'Connecting to Phantom...'
        
        const response = await window.solana.connect({ onlyIfTrusted: false })
        const publicKey = response.publicKey.toString()
        
        $('walletInfo').textContent = `Connected: ${publicKey.slice(0, 6)}...${publicKey.slice(-4)}`
        log(`Phantom wallet connected: ${publicKey}`)
        
        // Update button states
        updateWalletButtonStates(true)
        return
      }
      
      // Handle Ethereum networks
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
      
      // Request account access
      await window.ethereum.request({ method: 'eth_requestAccounts' })
      provider = new ethers.providers.Web3Provider(window.ethereum)
      signer = provider.getSigner()
      const addr = await signer.getAddress()
      $('walletInfo').textContent = `Connected: ${addr.slice(0, 6)}...${addr.slice(-4)}`
      log('Wallet connected successfully')
      
      // Update button states
      updateWalletButtonStates(true)
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
      
      // Handle Solana network
      if (networkKey === 'solana-devnet') {
        log('Solana approval not needed - SPL tokens use different approval mechanism')
        alert('Solana SPL tokens use a different approval mechanism. This step is not needed for Solana.')
        return
      }
      
      // Handle Ethereum networks
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
      
      // Record start time for transaction timing
      const startTime = Date.now()
      
      const networkKey = $('srcNetwork').value
      
      // Handle Solana network
      if (networkKey === 'solana-devnet') {
        log('Solana payment not yet implemented - this is a placeholder for future Solana integration')
        alert('Solana payment functionality is not yet implemented. This is a placeholder for future Solana smart contract integration.')
        setButtonLoading(false)
        return
      }
      
      // Handle Ethereum networks
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
      
      // Calculate transaction time
      const endTime = Date.now()
      const transactionTime = ((endTime - startTime) / 1000).toFixed(2)
      
      // Show success modal with transaction details
      showPaymentSuccess(tx.hash, transactionTime, rc.blockNumber, networkKey)
      
    } catch (e) {
      log(`Pay error: ${e.message || e}`)
      alert(`Payment failed: ${e.message || e}`)
    } finally {
      setButtonLoading(false)
    }
  })

  // Wallet button state management
  function updateWalletButtonStates(isConnected) {
    const connectBtn = document.getElementById('connectBtn')
    const disconnectBtn = document.getElementById('disconnectBtn')
    const switchWalletBtn = document.getElementById('switchWalletBtn')
    
    if (isConnected) {
      connectBtn.style.display = 'none'
      disconnectBtn.style.display = 'block'
      switchWalletBtn.style.display = 'block'
    } else {
      connectBtn.style.display = 'block'
      disconnectBtn.style.display = 'none'
      switchWalletBtn.style.display = 'none'
    }
  }
  
  // Disconnect wallet function
  async function disconnectWallet() {
    try {
      const selectedNetwork = $('srcNetwork').value
      
      if (selectedNetwork === 'solana-devnet') {
        // For Solana, we can't truly disconnect, but we can clear the state
        if (window.solana && window.solana.disconnect) {
          await window.solana.disconnect()
        }
        log('Solana wallet disconnected')
      } else {
        // For Ethereum, clear the provider and signer
        provider = null
        signer = null
        log('Ethereum wallet disconnected')
      }
      
      // Reset wallet info
      $('walletInfo').textContent = 'Wallet disconnected'
      
      // Update button states
      updateWalletButtonStates(false)
      
      // Reset to initial state
      setTimeout(() => {
        updateWalletInfoDisplay()
      }, 1000)
      
    } catch (e) {
      log(`Disconnect error: ${e.message || e}`)
      $('walletInfo').textContent = `Disconnect error: ${e.message || e}`
    }
  }
  
  // Switch wallet function
  async function switchWallet() {
    try {
      const selectedNetwork = $('srcNetwork').value
      
      if (selectedNetwork === 'solana-devnet') {
        // For Solana, request to connect again (this will show account selection)
        if (window.solana && window.solana.connect) {
          const response = await window.solana.connect({ onlyIfTrusted: false })
          const publicKey = response.publicKey.toString()
          $('walletInfo').textContent = `Connected: ${publicKey.slice(0, 6)}...${publicKey.slice(-4)}`
          log(`Switched to Solana wallet: ${publicKey}`)
        }
      } else {
        // For Ethereum, try to switch accounts using MetaMask's account switching
        try {
          // Try to trigger account switching by requesting permissions
          await window.ethereum.request({
            method: 'wallet_requestPermissions',
            params: [{ eth_accounts: {} }]
          })
        } catch (e) {
          // If permission request fails, try direct account request
          log('Permission request failed, trying direct account request')
          await window.ethereum.request({ method: 'eth_requestAccounts' })
        }
        
        // Update provider and signer with new account
        provider = new ethers.providers.Web3Provider(window.ethereum)
        signer = provider.getSigner()
        const addr = await signer.getAddress()
        $('walletInfo').textContent = `Connected: ${addr.slice(0, 6)}...${addr.slice(-4)}`
        log(`Switched to Ethereum wallet: ${addr}`)
      }
      
    } catch (e) {
      log(`Switch wallet error: ${e.message || e}`)
      if (e.code === 4001) {
        $('walletInfo').textContent = 'Account switch rejected by user'
      } else {
        $('walletInfo').textContent = `Switch error: ${e.message || e}`
      }
    }
  }
  
  // Event listeners for new buttons
  document.getElementById('disconnectBtn').addEventListener('click', disconnectWallet)
  document.getElementById('switchWalletBtn').addEventListener('click', switchWallet)

  // Payment success modal functions
  function showPaymentSuccess(txHash, transactionTime, blockNumber, networkKey) {
    // Update modal content
    document.getElementById('successTxHash').textContent = txHash
    document.getElementById('successTxTime').textContent = `${transactionTime} seconds`
    document.getElementById('successTxBlock').textContent = blockNumber
    
    // Show modal
    document.getElementById('paymentSuccessModal').style.display = 'flex'
    
    // Store network key for explorer link
    document.getElementById('paymentSuccessModal').dataset.network = networkKey
  }
  
  function hidePaymentSuccess() {
    document.getElementById('paymentSuccessModal').style.display = 'none'
  }
  
  // Success modal event listeners
  document.getElementById('viewOnExplorer').addEventListener('click', function() {
    const networkKey = document.getElementById('paymentSuccessModal').dataset.network
    const txHash = document.getElementById('successTxHash').textContent
    
    let explorerUrl = ''
    switch (networkKey) {
      case 'arbitrum-sepolia':
        explorerUrl = `https://sepolia.arbiscan.io/tx/${txHash}`
        break
      case 'base-sepolia':
        explorerUrl = `https://sepolia.basescan.org/tx/${txHash}`
        break
      case 'solana-devnet':
        explorerUrl = `https://explorer.solana.com/tx/${txHash}?cluster=devnet`
        break
      default:
        explorerUrl = `https://etherscan.io/tx/${txHash}`
    }
    
    window.open(explorerUrl, '_blank')
  })
  
  document.getElementById('exitPayment').addEventListener('click', function() {
    hidePaymentSuccess()
    // Optionally redirect to home page or close the payment window
    window.location.href = '/'
  })
  
  // Close modal when clicking outside
  document.getElementById('paymentSuccessModal').addEventListener('click', function(e) {
    if (e.target === this) {
      hidePaymentSuccess()
    }
  })

  // Fracted logo click handler
  const fractedLogo = document.getElementById('fractedLogo')
  if (fractedLogo) {
    fractedLogo.addEventListener('click', function() {
      // Navigate to main Fracted website
      window.location.href = 'https://fracted.xyz/'
      console.log('Navigating to Fracted main website...')
    })
  }
})
