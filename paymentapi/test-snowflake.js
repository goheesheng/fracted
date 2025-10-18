// Test script for Snowflake ID generator
import Snowflake from './snowflake.js'

const snowflake = new Snowflake({
  machineId: 1,
  epoch: 1700000000000 // 2023-11-15 00:00:00 UTC
})

console.log('Testing Snowflake ID Generator...\n')

// Generate multiple IDs to test
for (let i = 0; i < 10; i++) {
  const id = snowflake.generate()
  console.log(`Generated ID ${i + 1}: ${id}`)
  
  // Check if ID is positive
  if (parseInt(id) < 0) {
    console.error(`❌ Negative ID detected: ${id}`)
  } else {
    console.log(`✅ Positive ID: ${id}`)
  }
  
  // Small delay to ensure different timestamps
  await new Promise(resolve => setTimeout(resolve, 1))
}

console.log('\n✅ Snowflake ID generation test completed!')
