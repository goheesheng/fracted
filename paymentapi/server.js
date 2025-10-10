import express from 'express'
import path from 'path'
import dotenv from 'dotenv'
import { fileURLToPath } from 'url'
import { Options } from '@layerzerolabs/lz-v2-utilities'

dotenv.config()

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const app = express()
const PORT = process.env.PORT || 8080

// Serve static assets from /public
app.use(express.static(path.join(__dirname, 'public')))

// Health check
app.get('/health', (_req, res) => res.json({ ok: true }))

// Build config from environment variables and expose to frontend
// Expected env vars (optional except those you use):
// - OAPP_base_sepolia, OAPP_arbitrum_sepolia
// - TOKEN_base_sepolia_USDT, TOKEN_base_sepolia_USDC, TOKEN_base_sepolia_XUSD
// - TOKEN_arbitrum_sepolia_USDT, TOKEN_arbitrum_sepolia_USDC, TOKEN_arbitrum_sepolia_XUSD
function envConfig() {
  const networks = ['base-sepolia', 'arbitrum-sepolia']
  const symbols = ['USDT', 'USDC', 'XUSD']
  const contracts = {}
  const tokens = {}
  for (const net of networks) {
    const key = `OAPP_${net.replace('-', '_')}`
    if (process.env[key]) contracts[net] = process.env[key]
    tokens[net] = {}
    for (const sym of symbols) {
      const tKey = `TOKEN_${net.replace('-', '_')}_${sym}`
      if (process.env[tKey]) tokens[net][sym] = process.env[tKey]
    }
  }
  return { contracts, tokens }
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

// Root serves the payment page; parameters are read by client from the query string
app.get('/', (_req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'))
})

app.listen(PORT, () => {
  console.log(`Payment API running at http://localhost:${PORT}`)
})
