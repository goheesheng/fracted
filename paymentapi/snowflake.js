// Simple Snowflake ID Generator
// Based on Twitter's Snowflake algorithm with fixes for negative numbers

class Snowflake {
  constructor(options = {}) {
    this.machineId = options.machineId || 1
    // Use a more recent epoch to avoid large timestamp differences
    this.epoch = options.epoch || 1700000000000 // 2023-11-15 00:00:00 UTC
    this.sequence = 0
    this.lastTimestamp = 0
  }

  generate() {
    let timestamp = Date.now()
    
    if (timestamp < this.lastTimestamp) {
      throw new Error('Clock moved backwards. Refusing to generate id')
    }
    
    if (this.lastTimestamp === timestamp) {
      this.sequence = (this.sequence + 1) & 4095 // 12 bits for sequence
      if (this.sequence === 0) {
        timestamp = this.waitForNextMillis(this.lastTimestamp)
      }
    } else {
      this.sequence = 0
    }
    
    this.lastTimestamp = timestamp
    
    // Calculate timestamp difference
    const timestampDiff = timestamp - this.epoch
    
    // Ensure timestamp difference is positive and within safe range
    if (timestampDiff < 0) {
      throw new Error('Timestamp difference is negative. Check epoch setting.')
    }
    
    // Use a simpler approach to avoid overflow
    // Combine timestamp (milliseconds since epoch), machine ID, and sequence
    const id = timestampDiff * 1000000 + this.machineId * 1000 + this.sequence
    
    return id.toString()
  }
  
  waitForNextMillis(lastTimestamp) {
    let timestamp = Date.now()
    while (timestamp <= lastTimestamp) {
      timestamp = Date.now()
    }
    return timestamp
  }
}

export default Snowflake
