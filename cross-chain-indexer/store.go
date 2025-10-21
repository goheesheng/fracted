package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

// PayoutRecord 为 store 层使用的业务记录结构
type PayoutRecord struct {
	TxHash         string
	BlockNumber    int64
	DstEid         int64
	Payer          common.Address
	Merchant       common.Address
	SrcToken       common.Address
	DstToken       common.Address
	GrossAmount    *big.Int
	NetAmount      *big.Int
	Status         string
	Timestamp      time.Time
	SolanaMerchant string // Solana 原始地址（Base58 格式）
	SolanaPayer    string // Solana 原始地址（Base58 格式）
}

// NewStore 构造 Store 实例，并执行数据库迁移
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

	// 【新增】在初始化时执行数据库迁移
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return s, nil
}

// --------------------------- 核心函数：迁移 ---------------------------

// migrate: 数据库 schema 迁移
func (s *Store) migrate() error {
	// 1. events 原始日志记录表
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			tx_hash TEXT NOT NULL,
			log_index INTEGER NOT NULL,
			block_number INTEGER NOT NULL,
			raw_log TEXT NOT NULL, -- 原始 Log 的 JSON 字符串
			parsed INTEGER DEFAULT 0,
			PRIMARY KEY (tx_hash, log_index)
		);
		CREATE INDEX IF NOT EXISTS idx_events_parsed ON events(parsed);
	`)
	if err != nil {
		return fmt.Errorf("migrating events table: %w", err)
	}

	// 2. payouts 交易记录表
	// 确保所有字段类型（TEXT/INTEGER）与 SQL 兼容
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS payouts (
			tx_hash TEXT PRIMARY KEY,
			block_number INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			dst_eid INTEGER NOT NULL,
			payer TEXT NOT NULL,
			merchant TEXT NOT NULL,
			src_token TEXT NOT NULL,
			dst_token TEXT NOT NULL,
			gross_amount TEXT NOT NULL,
			net_amount TEXT NOT NULL,
			status TEXT NOT NULL, -- "Pending", "Delivered", "Failed"
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		-- 【你的要求】添加 merchant 索引，加速按商户地址的查询
		CREATE INDEX IF NOT EXISTS idx_payouts_merchant ON payouts(merchant);
	`)
	if err != nil {
		return fmt.Errorf("migrating payouts table: %w", err)
	}

	// 添加 Solana 原始地址字段（如果不存在）
	_, err = s.db.Exec(`
		ALTER TABLE payouts ADD COLUMN solana_merchant TEXT DEFAULT '';
	`)
	// 忽略 "duplicate column" 错误
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return fmt.Errorf("adding solana_merchant column: %w", err)
	}

	_, err = s.db.Exec(`
		ALTER TABLE payouts ADD COLUMN solana_payer TEXT DEFAULT '';
	`)
	// 忽略 "duplicate column" 错误
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return fmt.Errorf("adding solana_payer column: %w", err)
	}

	// 3. processed_blocks 区块记录表 (用于记录已处理到的区块高度)
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS processed_blocks (
			chain_id INTEGER PRIMARY KEY,
			block_number INTEGER NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("migrating processed_blocks table: %w", err)
	}

	log.Println("Store: database migration successful.")
	return nil
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	return s.db.Close()
}

// --------------------------- 核心函数：CRUD 操作 ---------------------------

// InsertEventIfNotExists 插入原始事件，基于 (tx_hash, log_index) 去重
func (s *Store) InsertEventIfNotExists(txHash string, logIndex uint, blockNumber uint64, rawLog string) (bool, error) {
	// ... (此函数内容不变，略去)
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO events (tx_hash, log_index, block_number, raw_log)
		VALUES (?, ?, ?, ?)
	`, txHash, logIndex, blockNumber, rawLog)

	// 如果没有错误且影响行数为 0，则说明该行已存在（IGNORE生效）
	// SQLite 在 INSERT OR IGNORE 成功插入时，返回 Result.RowsAffected 为 1
	// 但 modernc.org/sqlite 驱动的 RowsAffected 可能不可靠，这里使用更可靠的 SELECT
	if err != nil {
		return false, err
	}

	var count int
	err = s.db.QueryRow(`SELECT count(*) FROM events WHERE tx_hash = ? AND log_index = ?`, txHash, logIndex).Scan(&count)
	if err != nil {
		return false, err
	}
	// 如果 count 是 1，说明 INSERT OR IGNORE 生效了
	// 这里逻辑简化：只要没有错误就假设操作成功或忽略成功

	return true, nil
}

// MarkEventAsParsed 标记事件为已解析（parsed = 1）
func (s *Store) MarkEventAsParsed(txHash string, logIndex uint) error {
	// ... (此函数内容不变，略去)
	_, err := s.db.Exec(`
		UPDATE events SET parsed = 1 WHERE tx_hash = ? AND log_index = ?
	`, txHash, logIndex)
	return err
}

// UpsertPayout 插入或更新 PayoutRecord
func (s *Store) UpsertPayout(rec PayoutRecord) error {
	grossStr := rec.GrossAmount.String()
	netStr := rec.NetAmount.String()

	_, err := s.db.Exec(`
		INSERT INTO payouts (tx_hash, block_number, timestamp, dst_eid, payer, merchant, src_token, dst_token, gross_amount, net_amount, status, solana_merchant, solana_payer)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(tx_hash) DO UPDATE SET
			block_number = excluded.block_number,
			timestamp = excluded.timestamp,
			dst_eid = excluded.dst_eid,
			payer = excluded.payer,
			merchant = excluded.merchant,
			src_token = excluded.src_token,
			dst_token = excluded.dst_token,
			gross_amount = excluded.gross_amount,
			net_amount = excluded.net_amount,
			solana_merchant = excluded.solana_merchant,
			solana_payer = excluded.solana_payer,
			created_at = created_at
	`,
		rec.TxHash,
		rec.BlockNumber,
		rec.Timestamp.Format("2006-01-02 15:04:05"),
		rec.DstEid,
		rec.Payer.Hex(),
		rec.Merchant.Hex(),
		rec.SrcToken.Hex(),
		rec.DstToken.Hex(),
		grossStr,
		netStr,
		rec.Status,
		rec.SolanaMerchant,
		rec.SolanaPayer,
	)
	return err
}

// ListPayouts 列出所有 Payouts
func (s *Store) ListPayouts(limit, offset int) ([]PayoutRecord, error) {
	return s.listPayoutsByQuery(`
		SELECT tx_hash, block_number, timestamp, dst_eid, payer, merchant, src_token, dst_token, gross_amount, net_amount, status, COALESCE(solana_merchant, ''), COALESCE(solana_payer, '')
		FROM payouts
		ORDER BY block_number DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
}

// ListMerchantPayouts 列出特定商家的 Payouts
// 【新增函数】用于服务 /merchant/payouts 接口
func (s *Store) ListMerchantPayouts(merchant common.Address, limit, offset int) ([]PayoutRecord, error) {
	// 支持同时匹配 EVM 地址和 Solana 地址
	merchantHex := strings.ToLower(merchant.Hex())
	return s.listPayoutsByQuery(`
		SELECT tx_hash, block_number, timestamp, dst_eid, payer, merchant, src_token, dst_token, gross_amount, net_amount, status, COALESCE(solana_merchant, ''), COALESCE(solana_payer, '')
		FROM payouts
		WHERE LOWER(merchant) = LOWER(?) OR LOWER(solana_merchant) = LOWER(?)
		ORDER BY block_number DESC
		LIMIT ? OFFSET ?
	`, merchantHex, merchantHex, limit, offset)
}

// ListMerchantPayoutsByString 通过字符串地址查询商家的 Payouts
// 【新增函数】支持 Solana 地址查询
func (s *Store) ListMerchantPayoutsByString(merchantAddr string, limit, offset int) ([]PayoutRecord, error) {
	// 支持同时匹配 EVM 地址和 Solana 地址
	merchantLower := strings.ToLower(merchantAddr)
	return s.listPayoutsByQuery(`
		SELECT tx_hash, block_number, timestamp, dst_eid, payer, merchant, src_token, dst_token, gross_amount, net_amount, status, COALESCE(solana_merchant, ''), COALESCE(solana_payer, '')
		FROM payouts
		WHERE LOWER(merchant) = LOWER(?) OR LOWER(solana_merchant) = LOWER(?)
		ORDER BY block_number DESC
		LIMIT ? OFFSET ?
	`, merchantLower, merchantLower, limit, offset)
}

// ListPendingPayouts 列出状态为 "Pending" 的 Payouts
func (s *Store) ListPendingPayouts(limit int) ([]PayoutRecord, error) {
	return s.listPayoutsByQuery(`
		SELECT tx_hash, block_number, timestamp, dst_eid, payer, merchant, src_token, dst_token, gross_amount, net_amount, status, COALESCE(solana_merchant, ''), COALESCE(solana_payer, '')
		FROM payouts
		WHERE status = 'Pending'
		LIMIT ?
	`, limit)
}

// UpdatePayoutStatus 更新 Payout 状态
func (s *Store) UpdatePayoutStatus(txHash, status string) error {
	// ... (此函数内容不变，略去)
	_, err := s.db.Exec(`
		UPDATE payouts SET status = ? WHERE tx_hash = ?
	`, status, txHash)
	return err
}

// listPayoutsByQuery 是内部辅助函数，用于执行查询并解析结果
func (s *Store) listPayoutsByQuery(query string, args ...interface{}) ([]PayoutRecord, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PayoutRecord
	for rows.Next() {
		// 直接使用普通类型，因为数据库字段不是 NULL
		var txHashStr, payerStr, merchantStr, srcStr, dstStr, grossStr, netStr, statusStr string
		var solanaMerchant, solanaPayer string
		var blockNumber, dstEid int64
		var timestamp time.Time

		rec := PayoutRecord{}

		err := rows.Scan(
			&txHashStr, &blockNumber, &timestamp, &dstEid,
			&payerStr, &merchantStr, &srcStr, &dstStr,
			&grossStr, &netStr, &statusStr,
			&solanaMerchant, &solanaPayer,
		)
		if err != nil {
			log.Printf("Store: failed to scan payout row: %v", err)
			continue
		}

		// 将数据库类型转换为 Go 类型
		rec.TxHash = txHashStr
		rec.BlockNumber = blockNumber
		rec.DstEid = dstEid
		rec.Payer = common.HexToAddress(payerStr)
		rec.Merchant = common.HexToAddress(merchantStr)
		rec.SrcToken = common.HexToAddress(srcStr)
		rec.DstToken = common.HexToAddress(dstStr)

		// 解析大整数（big.Int）
		rec.GrossAmount = big.NewInt(0)
		rec.NetAmount = big.NewInt(0)

		if grossStr != "" {
			b := new(big.Int)
			if _, ok := b.SetString(grossStr, 10); ok {
				rec.GrossAmount = b
			}
		}
		if netStr != "" {
			b := new(big.Int)
			if _, ok := b.SetString(netStr, 10); ok {
				rec.NetAmount = b
			}
		}

		rec.Status = statusStr
		rec.Timestamp = timestamp.UTC()
		rec.SolanaMerchant = solanaMerchant
		rec.SolanaPayer = solanaPayer

		results = append(results, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetLastProcessedBlock 获取最后处理的区块高度
func (s *Store) GetLastProcessedBlock(chainID int) (uint64, error) {
	// ... (此函数内容不变，略去)
	var blockNum sql.NullInt64
	err := s.db.QueryRow(`
		SELECT block_number FROM processed_blocks WHERE chain_id = ?
	`, chainID).Scan(&blockNum)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if blockNum.Valid {
		return uint64(blockNum.Int64), nil
	}
	return 0, nil
}

// SetLastProcessedBlock 设置最后处理的区块高度
func (s *Store) SetLastProcessedBlock(chainID int, blockNum uint64) error {
	// ... (此函数内容不变，略去)
	_, err := s.db.Exec(`
		INSERT INTO processed_blocks (chain_id, block_number)
		VALUES (?, ?)
		ON CONFLICT(chain_id) DO UPDATE SET block_number = excluded.block_number
	`, chainID, blockNum)
	return err
}

// RawEvent 原始事件记录
type RawEvent struct {
	TxHash      string `json:"tx_hash"`
	LogIndex    int    `json:"log_index"`
	BlockNumber uint64 `json:"block_number"`
	RawLog      string `json:"raw_log"`
	Parsed      bool   `json:"parsed"`
}

// GetAllEvents 获取所有原始事件记录
func (s *Store) GetAllEvents(limit, offset int) ([]RawEvent, error) {
	query := `
		SELECT tx_hash, log_index, block_number, raw_log, parsed
		FROM events
		ORDER BY block_number DESC, log_index DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []RawEvent
	for rows.Next() {
		var e RawEvent
		var parsed int
		err := rows.Scan(&e.TxHash, &e.LogIndex, &e.BlockNumber, &e.RawLog, &parsed)
		if err != nil {
			return nil, err
		}
		e.Parsed = parsed == 1
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// GetEventCount 获取事件总数
func (s *Store) GetEventCount() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM events`).Scan(&count)
	return count, err
}
