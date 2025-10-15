// ===== Quick Payment Link Generator =====
// Modify the configuration below, then run: node quick-link.js

// 🔧 Configure your merchant information
const MERCHANT_ADDRESS = '0xB7aa464b19037CF3dB7F723504dFafE7b63aAb84'  // Your merchant address
const DESTINATION_EID = 40231  // Destination chain ID (40245=Base Sepolia, 40231=Arbitrum Sepolia)
const DESTINATION_TOKEN = '0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d'  // Destination token address
const AMOUNT = 1000000  // Payment amount (smallest units)
const SERVER_URL = 'https://demo.fracted.xyz'  // Server URL

// 🚀 Generate payment link
function generateLink() {
  const params = new URLSearchParams({
    merchant: MERCHANT_ADDRESS,
    dstEid: DESTINATION_EID.toString(),
    dstToken: DESTINATION_TOKEN,
    amount: AMOUNT.toString()
  })
  
  return `${SERVER_URL}/?${params.toString()}`
}

// 📋 Display results
const paymentLink = generateLink()
console.log('🔗 Payment Link:')
console.log(paymentLink)
console.log('\n📋 Configuration:')
console.log(`Merchant Address: ${MERCHANT_ADDRESS}`)
console.log(`Destination Chain EID: ${DESTINATION_EID}`)
console.log(`Destination Token: ${DESTINATION_TOKEN}`)
console.log(`Payment Amount: ${AMOUNT}`)
console.log(`\n💡 Copy the link and open it in your browser to start payment!`)
