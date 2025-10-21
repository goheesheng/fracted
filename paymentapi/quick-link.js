// ===== Quick Payment Link Generator =====
// Modify the configuration below, then run: node quick-link.js

// 🔧 Configure your merchant information
const MERCHANT_ADDRESS = '0xB7aa464b19037CF3dB7F723504dFafE7b63aAb84'  // Your merchant address (EVM: 0x..., Solana: base58)
const DESTINATION_EID = 40231  // Destination chain ID (40245=Base Sepolia, 40231=Arbitrum Sepolia, 40168=Solana Devnet)
const DESTINATION_TOKEN = '0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d'  // Destination token address (EVM: 0x..., Solana: base58)
const AMOUNT = 1000000  // Payment amount (smallest units)
const SERVER_URL = 'https://demo.fracted.xyz'  // Server URL

// 📝 Solana Example Configuration:
// const MERCHANT_ADDRESS = '7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU'
// const DESTINATION_EID = 40168
// const DESTINATION_TOKEN = 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v'  // USDC on Solana

// 🚀 Generate payment link (using new Snowflake ID system)
async function generateLink() {
  try {
    const params = new URLSearchParams({
      merchant: MERCHANT_ADDRESS,
      dstEid: DESTINATION_EID.toString(),
      dstToken: DESTINATION_TOKEN,
      amount: AMOUNT.toString()
    })
    
    // Call the API to generate payment link with Snowflake ID
    const response = await fetch(`${SERVER_URL}/generate-link?${params.toString()}`)
    const data = await response.json()
    
    if (data.success) {
      return data.paymentLink
    } else {
      throw new Error(data.error || 'Failed to generate payment link')
    }
  } catch (error) {
    console.error('Error generating payment link:', error.message)
    throw error
  }
}

// 📋 Display results
async function main() {
  try {
    const paymentLink = await generateLink()
    console.log('🔗 Payment Link:')
    console.log(paymentLink)
    console.log('\n📋 Configuration:')
    console.log(`Merchant Address: ${MERCHANT_ADDRESS}`)
    console.log(`Destination Chain EID: ${DESTINATION_EID}`)
    console.log(`Destination Token: ${DESTINATION_TOKEN}`)
    console.log(`Payment Amount: ${AMOUNT}`)
    console.log(`\n💡 Copy the link and open it in your browser to start payment!`)
  } catch (error) {
    console.error('❌ Error:', error.message)
    process.exit(1)
  }
}

// Run the script
main()
