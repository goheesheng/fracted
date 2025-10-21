import express from 'express'
import path from 'path'
import dotenv from 'dotenv'
import { fileURLToPath } from 'url'
import { Options } from '@layerzerolabs/lz-v2-utilities'
import PaymentDatabase from './database.js'

dotenv.config()

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const app = express()
const PORT = process.env.PORT || 8080

// Initialize database
const paymentDB = new PaymentDatabase()

// Middleware for parsing JSON
app.use(express.json())

// Serve static assets from /public with no-cache headers
app.use(express.static(path.join(__dirname, 'public'), {
  setHeaders: (res, path) => {
    if (path.endsWith('.html') || path.endsWith('.css') || path.endsWith('.js')) {
      res.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate')
      res.setHeader('Pragma', 'no-cache')
      res.setHeader('Expires', '0')
    }
  }
}))

// Build config from environment variables and expose to frontend
// Expected env vars (optional except those you use):
// - OAPP_base_sepolia, OAPP_arbitrum_sepolia
// - TOKEN_base_sepolia_USDT, TOKEN_base_sepolia_USDC, TOKEN_base_sepolia_XUSD
// - TOKEN_arbitrum_sepolia_USDT, TOKEN_arbitrum_sepolia_USDC, TOKEN_arbitrum_sepolia_XUSD
function envConfig() {
  const networks = ['base-sepolia', 'arbitrum-sepolia', 'solana-devnet']
  const symbols = ['USDT', 'USDC', 'XUSD']
  const contracts = {}
  const tokens = {}
  
  // Load contract addresses
  for (const net of networks) {
    const key = `OAPP_${net.replace('-', '_')}`
    if (process.env[key]) contracts[net] = process.env[key]
    tokens[net] = {}
    for (const sym of symbols) {
      const tKey = `TOKEN_${net.replace('-', '_')}_${sym}`
      if (process.env[tKey]) tokens[net][sym] = process.env[tKey]
    }
  }
  
  // Load EID to chain name mappings
  const eidToChain = {}
  for (const [key, value] of Object.entries(process.env)) {
    if (key.startsWith('EID_TO_CHAIN_')) {
      const eid = key.replace('EID_TO_CHAIN_', '')
      eidToChain[eid] = value
    }
  }
  
  // Load token address to symbol mappings
  const tokenToSymbol = {}
  for (const [key, value] of Object.entries(process.env)) {
    if (key.startsWith('TOKEN_SYMBOL_')) {
      const address = key.replace('TOKEN_SYMBOL_', '')
      tokenToSymbol[address] = value
    }
  }
  
  return { contracts, tokens, eidToChain, tokenToSymbol }
}

app.get('/config', (_req, res) => {
  res.json(envConfig())
})

// Build and return a valid LayerZero V2 options hex for executor receive gas
// Usage: GET /options?gas=150000
app.get('/options', (req, res) => {
  try {
    const gas = Number(req.query.gas ?? 150000)
    const optionsHex = Options.newOptions().addExecutorLzReceiveOption(gas, 0).toHex()
    res.json({ optionsHex, gas })
  } catch (e) {
    res.status(400).json({ error: e?.message || String(e) })
  }
})

// Generate payment link with payment ID
// Usage: GET /generate-link?merchant=0x...&dstEid=40245&dstToken=0x...&amount=1000000
app.get('/generate-link', async (req, res) => {
  try {
    const { merchant, dstEid, dstToken, amount } = req.query
    
    // Validate required parameters
    if (!merchant || !dstEid || !dstToken || !amount) {
      return res.status(400).json({ 
        error: 'Missing required parameters: merchant, dstEid, dstToken, amount' 
      })
    }
    
    // Validate merchant address format
    if (!/^0x[a-fA-F0-9]{40}$/.test(merchant)) {
      return res.status(400).json({ 
        error: 'Invalid merchant address format' 
      })
    }
    
    // Validate dstEid is a number
    const dstEidNum = Number(dstEid)
    if (isNaN(dstEidNum) || dstEidNum <= 0) {
      return res.status(400).json({ 
        error: 'Invalid dstEid: must be a positive number' 
      })
    }
    
    // Validate dstToken address format
    if (!/^0x[a-fA-F0-9]{40}$/.test(dstToken)) {
      return res.status(400).json({ 
        error: 'Invalid dstToken address format' 
      })
    }
    
    // Validate amount is a positive number
    const amountNum = Number(amount)
    if (isNaN(amountNum) || amountNum <= 0) {
      return res.status(400).json({ 
        error: 'Invalid amount: must be a positive number' 
      })
    }
    
    // Create payment record in database
    const paymentId = await paymentDB.createPayment(merchant, dstEidNum, dstToken, amountNum)
    
    // Generate the new payment link with payment ID
    const baseUrl = 'https://demo.fracted.xyz'
    const paymentLink = `${baseUrl}/payment/${paymentId}`
    
    res.json({
      success: true,
      paymentId,
      paymentLink,
      parameters: {
        merchant,
        dstEid: dstEidNum,
        dstToken,
        amount: amountNum
      },
      message: 'Payment link generated successfully'
    })
    
  } catch (e) {
    res.status(500).json({ error: e?.message || String(e) })
  }
})

// Get payment information by payment ID
// Usage: GET /api/payment/:paymentId
app.get('/api/payment/:paymentId', async (req, res) => {
  try {
    const { paymentId } = req.params
    
    if (!paymentId) {
      return res.status(400).json({ 
        error: 'Payment ID is required' 
      })
    }
    
    const payment = await paymentDB.getPayment(paymentId)
    
    if (!payment) {
      return res.status(404).json({ 
        error: 'Payment not found' 
      })
    }
    
    res.json({
      success: true,
      payment: {
        id: payment.id,
        merchant: payment.merchant_address,
        dstEid: payment.dst_eid,
        dstToken: payment.dst_token,
        amount: payment.amount,
        status: payment.status,
        createdAt: payment.created_at,
        updatedAt: payment.updated_at
      }
    })
    
  } catch (e) {
    res.status(500).json({ error: e?.message || String(e) })
  }
})

// Update payment status
// Usage: POST /api/payment/:paymentId/status
app.post('/api/payment/:paymentId/status', async (req, res) => {
  try {
    const { paymentId } = req.params
    const { status } = req.body
    
    if (!paymentId) {
      return res.status(400).json({ 
        error: 'Payment ID is required' 
      })
    }
    
    if (!status) {
      return res.status(400).json({ 
        error: 'Status is required' 
      })
    }
    
    const validStatuses = ['pending', 'processing', 'completed', 'failed', 'cancelled']
    if (!validStatuses.includes(status)) {
      return res.status(400).json({ 
        error: `Invalid status. Must be one of: ${validStatuses.join(', ')}` 
      })
    }
    
    const updated = await paymentDB.updatePaymentStatus(paymentId, status)
    
    if (updated === 0) {
      return res.status(404).json({ 
        error: 'Payment not found' 
      })
    }
    
    res.json({
      success: true,
      message: 'Payment status updated successfully'
    })
    
  } catch (e) {
    res.status(500).json({ error: e?.message || String(e) })
  }
})

// Get all payments (for admin/debugging)
// Usage: GET /api/payments
app.get('/api/payments', async (req, res) => {
  try {
    const payments = await paymentDB.getAllPayments()
    
    res.json({
      success: true,
      payments: payments.map(payment => ({
        id: payment.id,
        merchant: payment.merchant_address,
        dstEid: payment.dst_eid,
        dstToken: payment.dst_token,
        amount: payment.amount,
        status: payment.status,
        createdAt: payment.created_at,
        updatedAt: payment.updated_at
      }))
    })
    
  } catch (e) {
    res.status(500).json({ error: e?.message || String(e) })
  }
})

// Initiate payment page route
app.get('/initiatebyuser', (_req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'initiate.html'))
})

// Root serves the generator page
app.get('/', (_req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'))
})

// Payment page route with payment ID
app.get('/payment/:paymentId', (_req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'payment.html'))
})

// Legacy payment page route (for backward compatibility)
app.get('/payment', (_req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'payment.html'))
})

app.listen(PORT, () => {
  console.log(`Payment API running at http://localhost:${PORT}`)
})
