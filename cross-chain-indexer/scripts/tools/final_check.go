package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func main() {
	db, _ := sql.Open("sqlite", "indexer.db")
	defer db.Close()

	// 检查新交易
	var count int
	db.QueryRow("SELECT COUNT(*) FROM payouts WHERE tx_hash = '0xdf9678e4d73ca79ffb65e88f207537af3bb42f1ea62b00129694154361d2faf0'").Scan(&count)

	fmt.Println("检查新交易 (0xdf9678e4...):")
	if count > 0 {
		fmt.Println("  ✅ 已在数据库中")

		var timestamp, merchant, status string
		var dstEid int64
		db.QueryRow(`
			SELECT timestamp, merchant, dst_eid, status 
			FROM payouts 
			WHERE tx_hash = '0xdf9678e4d73ca79ffb65e88f207537af3bb42f1ea62b00129694154361d2faf0'
		`).Scan(&timestamp, &merchant, &dstEid, &status)

		fmt.Printf("  时间: %s\n", timestamp)
		fmt.Printf("  商家: %s\n", merchant)
		fmt.Printf("  目标链: EID %d\n", dstEid)
		fmt.Printf("  状态: %s\n", status)
	} else {
		fmt.Println("  ❌ 未找到")
	}

	// 统计各链交易数
	fmt.Println("\n各链交易统计:")
	rows, _ := db.Query("SELECT dst_eid, COUNT(*) as cnt FROM payouts GROUP BY dst_eid ORDER BY dst_eid")
	defer rows.Close()

	for rows.Next() {
		var dstEid, cnt int64
		rows.Scan(&dstEid, &cnt)
		chainName := "Unknown"
		if dstEid == 40168 {
			chainName = "Solana Devnet"
		} else if dstEid == 40231 {
			chainName = "Arbitrum Sepolia"
		} else if dstEid == 40245 {
			chainName = "Base Sepolia"
		}
		fmt.Printf("  EID %d (%s): %d 笔\n", dstEid, chainName, cnt)
	}
}
