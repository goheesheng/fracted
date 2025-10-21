// Test script for Solana address validation
// Run: node test-solana-address.js

// Copy the validation function from server.js
function isValidAddress(address, dstEid) {
  const eid = Number(dstEid)
  
  // Solana Devnet (EID 40168)
  if (eid === 40168) {
    // Solana address: base58 encoded, typically 32-44 characters
    return /^[1-9A-HJ-NP-Za-km-z]{32,44}$/.test(address)
  }
  
  // EVM chains (Ethereum, Base, Arbitrum, etc.)
  return /^0x[a-fA-F0-9]{40}$/.test(address)
}

console.log('=== Solana Address Validation Tests ===\n')

// Test cases
const tests = [
  // EVM addresses
  {
    address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    eid: 40245,
    expected: true,
    description: 'Valid EVM address for Base Sepolia'
  },
  {
    address: '0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d',
    eid: 40231,
    expected: true,
    description: 'Valid EVM address for Arbitrum Sepolia'
  },
  {
    address: '0xinvalid',
    eid: 40245,
    expected: false,
    description: 'Invalid EVM address (too short)'
  },
  
  // Solana addresses
  {
    address: '7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU',
    eid: 40168,
    expected: true,
    description: 'Valid Solana address'
  },
  {
    address: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v',
    eid: 40168,
    expected: true,
    description: 'Valid Solana token address (USDC)'
  },
  {
    address: 'Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB',
    eid: 40168,
    expected: true,
    description: 'Valid Solana token address (USDT)'
  },
  
  // Cross-chain validation (should fail)
  {
    address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
    eid: 40168,
    expected: false,
    description: 'EVM address used for Solana (should fail)'
  },
  {
    address: '7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU',
    eid: 40245,
    expected: false,
    description: 'Solana address used for EVM (should fail)'
  },
  
  // Invalid formats
  {
    address: 'invalid',
    eid: 40168,
    expected: false,
    description: 'Invalid Solana address (too short)'
  },
  {
    address: '0OOl1111222233334444555566667777888899',
    eid: 40168,
    expected: false,
    description: 'Invalid Solana address (contains O, 0, l)'
  }
]

// Run tests
let passed = 0
let failed = 0

tests.forEach((test, index) => {
  const result = isValidAddress(test.address, test.eid)
  const status = result === test.expected ? '✅ PASS' : '❌ FAIL'
  
  console.log(`Test ${index + 1}: ${status}`)
  console.log(`  Description: ${test.description}`)
  console.log(`  Address: ${test.address}`)
  console.log(`  EID: ${test.eid}`)
  console.log(`  Expected: ${test.expected}, Got: ${result}`)
  console.log()
  
  if (result === test.expected) {
    passed++
  } else {
    failed++
  }
})

// Summary
console.log('=== Test Summary ===')
console.log(`Total: ${tests.length}`)
console.log(`✅ Passed: ${passed}`)
console.log(`❌ Failed: ${failed}`)
console.log(`Success Rate: ${(passed / tests.length * 100).toFixed(1)}%`)

// Exit with error code if any test failed
if (failed > 0) {
  process.exit(1)
}

