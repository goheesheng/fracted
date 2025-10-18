// Simple Snowflake ID Generator
// Based on Twitter's Snowflake algorithm

class Snowflake {
  constructor(options = {}) {
    this.machineId = options.machineId || 1
    this.epoch = options.epoch || 1640995200000 // 2022-01-01 00:00:00 UTC
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
    
    // 41 bits for timestamp, 10 bits for machine ID, 12 bits for sequence
    const id = ((timestamp - this.epoch) << 22) | (this.machineId << 12) | this.sequence
    
    return id
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
