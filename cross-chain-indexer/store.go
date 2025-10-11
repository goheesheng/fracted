package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

// PayoutRecord 为 store 层使用的业务记录结构
type PayoutRecord struct {
	TxHash      string
	BlockNumber int64
	DstEid      int64
	Payer       common.Address
	Merchant    common.Address
	SrcToken    common.Address
	DstToken    common.Address
	GrossAmount *big.Int
	NetAmount   *big.Int
	Status      string
	Timestamp   time.Time
}

func NewStore(path string) (*Store, error) {
	// sqlite3 DSN: 设置 busy timeout 与启用 foreign_keys
	dsn := fmt.Sprintf("%s?_busy_timeout=5000&_foreign_keys=1", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// 保守设置
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = s.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	schema := `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tx_hash TEXT NOT NULL,
  log_index INTEGER NOT NULL,
  block_number INTEGER NOT NULL,
  topic TEXT NOT NULL,
  data BLOB,
  parsed INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(tx_hash, log_index)
);

CREATE TABLE IF NOT EXISTS payouts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tx_hash TEXT NOT NULL UNIQUE,
  block_number INTEGER,
  dst_eid INTEGER,
  payer TEXT,
  merchant TEXT,
  src_token TEXT,
  dst_token TEXT,
  gross_amount TEXT,
  net_amount TEXT,
  status TEXT,
  timestamp DATETIME,
  last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS processed_blocks (
  chain_name TEXT PRIMARY KEY,
  last_block INTEGER,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
	_, err := s.db.Exec(schema)
	return err
}

// InsertEventIfNotExists 插入事件（去重）；返回 true 表示插入成功，false 表示已存在
func (s *Store) InsertEventIfNotExists(txHash string, logIndex uint, blockNumber uint64, topic string, data []byte) (bool, error) {
	sqlStmt := `INSERT OR IGNORE INTO events (tx_hash, log_index, block_number, topic, data) VALUES (?, ?, ?, ?, ?)`
	res, err := s.db.Exec(sqlStmt, txHash, int(logIndex), int64(blockNumber), topic, data)
	if err != nil {
		return false, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return ra > 0, nil
}

func (s *Store) MarkEventParsed(txHash string, logIndex uint) error {
	sqlStmt := `UPDATE events SET parsed = 1 WHERE tx_hash = ? AND log_index = ?`
	_, err := s.db.Exec(sqlStmt, txHash, int(logIndex))
	return err
}

// UpsertPayout 将 payout 写入或更新（基于 tx_hash 唯一约束）
// amount 字段以字符串形式存储（big.Int.String）
func (s *Store) UpsertPayout(p PayoutRecord) error {
	gross := ""
	net := ""
	if p.GrossAmount != nil {
		gross = p.GrossAmount.String()
	}
	if p.NetAmount != nil {
		net = p.NetAmount.String()
	}
	ts := p.Timestamp.UTC().Format(time.RFC3339)
	sqlStmt := `
INSERT INTO payouts (tx_hash, block_number, dst_eid, payer, merchant, src_token, dst_token, gross_amount, net_amount, status, timestamp, last_updated, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT(tx_hash) DO UPDATE SET
  block_number=excluded.block_number,
  dst_eid=excluded.dst_eid,
  payer=excluded.payer,
  merchant=excluded.merchant,
  src_token=excluded.src_token,
  dst_token=excluded.dst_token,
  gross_amount=excluded.gross_amount,
  net_amount=excluded.net_amount,
  status=excluded.status,
  last_updated=CURRENT_TIMESTAMP;
`
	_, err := s.db.Exec(sqlStmt,
		p.TxHash, p.BlockNumber, p.DstEid,
		p.Payer.Hex(), p.Merchant.Hex(), p.SrcToken.Hex(), p.DstToken.Hex(),
		gross, net, p.Status, ts,
	)
	return err
}

// ListPayouts 分页读取 payouts（返回 PayoutRecord 列表）
func (s *Store) ListPayouts(limit, offset int) ([]PayoutRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	sqlStmt := `SELECT tx_hash, block_number, dst_eid, payer, merchant, src_token, dst_token, gross_amount, net_amount, status, timestamp FROM payouts ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(sqlStmt, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PayoutRecord
	for rows.Next() {
		var txHash string
		var blockNumber sql.NullInt64
		var dstEid sql.NullInt64
		var payerStr, merchantStr, srcStr, dstStr sql.NullString
		var grossStr, netStr sql.NullString
		var status sql.NullString
		var ts sql.NullString

		if err := rows.Scan(&txHash, &blockNumber, &dstEid, &payerStr, &merchantStr, &srcStr, &dstStr, &grossStr, &netStr, &status, &ts); err != nil {
			log.Printf("ListPayouts: scan error: %v", err)
			continue
		}

		rec := PayoutRecord{
			TxHash:      txHash,
			BlockNumber: 0,
			DstEid:      0,
			Payer:       common.Address{},
			Merchant:    common.Address{},
			SrcToken:    common.Address{},
			DstToken:    common.Address{},
			GrossAmount: big.NewInt(0),
			NetAmount:   big.NewInt(0),
			Status:      "",
			Timestamp:   time.Time{},
		}
		if blockNumber.Valid {
			rec.BlockNumber = blockNumber.Int64
		}
		if dstEid.Valid {
			rec.DstEid = dstEid.Int64
		}
		if payerStr.Valid {
			rec.Payer = common.HexToAddress(payerStr.String)
		}
		if merchantStr.Valid {
			rec.Merchant = common.HexToAddress(merchantStr.String)
		}
		if srcStr.Valid {
			rec.SrcToken = common.HexToAddress(srcStr.String)
		}
		if dstStr.Valid {
			rec.DstToken = common.HexToAddress(dstStr.String)
		}
		if grossStr.Valid && grossStr.String != "" {
			b := new(big.Int)
			if _, ok := b.SetString(grossStr.String, 10); ok {
				rec.GrossAmount = b
			}
		}
		if netStr.Valid && netStr.String != "" {
			b := new(big.Int)
			if _, ok := b.SetString(netStr.String, 10); ok {
				rec.NetAmount = b
			}
		}
		if status.Valid {
			rec.Status = status.String
		}
		if ts.Valid && ts.String != "" {
			// 首先尝试 RFC3339（这是我们写入时使用的格式）
			if t, err := time.Parse(time.RFC3339, ts.String); err == nil {
				rec.Timestamp = t
			} else if t2, err2 := time.ParseInLocation("2006-01-02 15:04:05", ts.String, time.Local); err2 == nil {
				// 兼容旧版写入格式（如果存在）
				rec.Timestamp = t2.UTC()
			} else {
				// 解析失败：保留 zero value（并可记录日志以便排查）
				// log.Printf("ListPayouts: cannot parse timestamp %q: %v / %v", ts.String, err, err2)
			}
		}
		out = append(out, rec)
	}
	return out, nil
}

// PendingPayoutRow 简化结构用于查询待处理记录
type PendingPayoutRow struct {
	TxHash      string
	BlockNumber int64
	Status      string
}

// ListPendingPayouts 返回当前处于 Pending 状态的 payouts（limit 可控）
func (s *Store) ListPendingPayouts(limit int) ([]PendingPayoutRow, error) {
	if limit <= 0 {
		limit = 100
	}
	sqlStmt := `SELECT tx_hash, block_number, status FROM payouts WHERE status = 'Pending' ORDER BY created_at ASC LIMIT ?`
	rows, err := s.db.Query(sqlStmt, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PendingPayoutRow
	for rows.Next() {
		var r PendingPayoutRow
		if err := rows.Scan(&r.TxHash, &r.BlockNumber, &r.Status); err != nil {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

// UpdatePayoutStatus 更新指定 tx 的状态（并更新时间戳）
// status 应为 "Pending"|"Delivered"|"Failed"
func (s *Store) UpdatePayoutStatus(txHash string, status string) error {
	sqlStmt := `UPDATE payouts SET status = ?, last_updated = CURRENT_TIMESTAMP WHERE tx_hash = ?`
	_, err := s.db.Exec(sqlStmt, status, txHash)
	return err
}

// ------------------ 新增方法结束 ------------------
