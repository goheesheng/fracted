package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"  // PostgreSQL 驱动
	_ "modernc.org/sqlite" // SQLite 驱动
)

// PayoutRequest 与 store.go 中一致
type PayoutRequest struct {
	TxHash      string
	BlockNumber uint64
	Timestamp   time.Time
	DstEid      uint32
	Payer       string
	Merchant    string
	SrcToken    string
	DstToken    string
	GrossAmount string
	NetPayout   string
	Status      string
	CreatedAt   time.Time
}

func run() {
	// 1️⃣ 打开 SQLite
	sqliteDB, err := sql.Open("sqlite", "indexer.db")
	if err != nil {
		log.Fatalf("打开 SQLite 失败: %v", err)
	}
	defer sqliteDB.Close()

	// 2️⃣ 打开 PostgreSQL
	// 替换 user/pass/dbname
	pgConnStr := "postgres://postgres:password@localhost:5432/crosschain_indexer?sslmode=disable"
	pgDB, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatalf("连接 PostgreSQL 失败: %v", err)
	}
	defer pgDB.Close()

	// 3️⃣ 查询 SQLite payouts
	rows, err := sqliteDB.Query(`SELECT tx_hash, block_number, timestamp, dst_eid, payer, merchant, src_token, dst_token, gross_amount, net_amount, status, created_at FROM payouts`)
	if err != nil {
		log.Fatalf("查询 SQLite 失败: %v", err)
	}
	defer rows.Close()

	var payouts []PayoutRequest
	for rows.Next() {
		var p PayoutRequest
		var ts string
		var created string
		err := rows.Scan(&p.TxHash, &p.BlockNumber, &ts, &p.DstEid, &p.Payer, &p.Merchant, &p.SrcToken, &p.DstToken, &p.GrossAmount, &p.NetPayout, &p.Status, &created)
		if err != nil {
			log.Fatalf("读取 SQLite 行失败: %v", err)
		}
		// SQLite timestamp 转 time.Time
		p.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		payouts = append(payouts, p)
	}

	fmt.Printf("共读取 %d 条 payouts\n", len(payouts))

	// 4️⃣ 插入 PostgreSQL
	insertSQL := `
	INSERT INTO payouts (
		tx_hash, block_number, timestamp, dst_eid, payer, merchant, src_token, dst_token, gross_amount, net_amount, status, created_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	ON CONFLICT (tx_hash) DO NOTHING
	`

	for _, p := range payouts {
		_, err := pgDB.Exec(insertSQL, p.TxHash, p.BlockNumber, p.Timestamp, p.DstEid, p.Payer, p.Merchant, p.SrcToken, p.DstToken, p.GrossAmount, p.NetPayout, p.Status, p.CreatedAt)
		if err != nil {
			log.Printf("插入 PostgreSQL 失败 tx_hash=%s: %v", p.TxHash, err)
		}
	}

	fmt.Println("迁移完成 ✅")
}
