// Test script for payment ID system
import fetch from 'node-fetch'

const BASE_URL = 'http://localhost:8080'

async function testPaymentIdSystem() {
  console.log('Testing Payment ID System...\n')
  
  try {
    // Test 1: Generate payment link
    console.log('1. Testing payment link generation...')
    const generateResponse = await fetch(`${BASE_URL}/generate-link?merchant=0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6&dstEid=40245&dstToken=0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d&amount=123000000`)
    
    if (!generateResponse.ok) {
      throw new Error(`Generate link failed: ${generateResponse.status}`)
    }
    
    const generateData = await generateResponse.json()
    console.log('Generated payment data:', generateData)
    
    const paymentId = generateData.paymentId
    const paymentLink = generateData.paymentLink
    
    console.log(`Payment ID: ${paymentId}`)
    console.log(`Payment Link: ${paymentLink}\n`)
    
    // Test 2: Get payment information
    console.log('2. Testing payment information retrieval...')
    const getResponse = await fetch(`${BASE_URL}/api/payment/${paymentId}`)
    
    if (!getResponse.ok) {
      throw new Error(`Get payment failed: ${getResponse.status}`)
    }
    
    const getData = await getResponse.json()
    console.log('Retrieved payment data:', getData)
    
    // Test 3: Update payment status
    console.log('\n3. Testing payment status update...')
    const updateResponse = await fetch(`${BASE_URL}/api/payment/${paymentId}/status`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ status: 'processing' })
    })
    
    if (!updateResponse.ok) {
      throw new Error(`Update status failed: ${updateResponse.status}`)
    }
    
    const updateData = await updateResponse.json()
    console.log('Update status result:', updateData)
    
    // Test 4: Get all payments
    console.log('\n4. Testing get all payments...')
    const allResponse = await fetch(`${BASE_URL}/api/payments`)
    
    if (!allResponse.ok) {
      throw new Error(`Get all payments failed: ${allResponse.status}`)
    }
    
    const allData = await allResponse.json()
    console.log('All payments:', allData)
    
    console.log('\n✅ All tests passed!')
    
  } catch (error) {
    console.error('❌ Test failed:', error.message)
  }
}

// Run the test
testPaymentIdSystem()
