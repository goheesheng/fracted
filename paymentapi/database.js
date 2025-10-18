import sqlite3 from 'sqlite3'
import Snowflake from './snowflake.js'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

// Initialize Snowflake ID generator
const snowflake = new Snowflake({
  machineId: 1, // You can set this to a unique machine ID
  epoch: 1640995200000 // Custom epoch (2022-01-01 00:00:00 UTC)
})

class PaymentDatabase {
  constructor() {
    this.db = null
    this.init()
  }

  init() {
    const dbPath = path.join(__dirname, 'payments.db')
    this.db = new sqlite3.Database(dbPath, (err) => {
      if (err) {
        console.error('Error opening database:', err.message)
      } else {
        console.log('Connected to SQLite database')
        this.createTable()
      }
    })
  }

  createTable() {
    const createTableSQL = `
      CREATE TABLE IF NOT EXISTS payments (
        id TEXT PRIMARY KEY,
        merchant_address TEXT NOT NULL,
        dst_eid INTEGER NOT NULL,
        dst_token TEXT NOT NULL,
        amount TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        status TEXT DEFAULT 'pending'
      )
    `

    this.db.run(createTableSQL, (err) => {
      if (err) {
        console.error('Error creating table:', err.message)
      } else {
        console.log('Payments table created or already exists')
      }
    })
  }

  generatePaymentId() {
    return snowflake.generate().toString()
  }

  createPayment(merchantAddress, dstEid, dstToken, amount) {
    return new Promise((resolve, reject) => {
      const paymentId = this.generatePaymentId()
      const sql = `
        INSERT INTO payments (id, merchant_address, dst_eid, dst_token, amount)
        VALUES (?, ?, ?, ?, ?)
      `
      
      this.db.run(sql, [paymentId, merchantAddress, dstEid, dstToken, amount], function(err) {
        if (err) {
          reject(err)
        } else {
          resolve(paymentId)
        }
      })
    })
  }

  getPayment(paymentId) {
    return new Promise((resolve, reject) => {
      const sql = 'SELECT * FROM payments WHERE id = ?'
      
      this.db.get(sql, [paymentId], (err, row) => {
        if (err) {
          reject(err)
        } else {
          resolve(row)
        }
      })
    })
  }

  updatePaymentStatus(paymentId, status) {
    return new Promise((resolve, reject) => {
      const sql = `
        UPDATE payments 
        SET status = ?, updated_at = CURRENT_TIMESTAMP 
        WHERE id = ?
      `
      
      this.db.run(sql, [status, paymentId], function(err) {
        if (err) {
          reject(err)
        } else {
          resolve(this.changes)
        }
      })
    })
  }

  getAllPayments() {
    return new Promise((resolve, reject) => {
      const sql = 'SELECT * FROM payments ORDER BY created_at DESC'
      
      this.db.all(sql, [], (err, rows) => {
        if (err) {
          reject(err)
        } else {
          resolve(rows)
        }
      })
    })
  }

  close() {
    if (this.db) {
      this.db.close((err) => {
        if (err) {
          console.error('Error closing database:', err.message)
        } else {
          console.log('Database connection closed')
        }
      })
    }
  }
}

export default PaymentDatabase
