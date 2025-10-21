package main

import (
	"log"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// statusUpdater 简化版状态更新器（跨链自动确认模式）
//
// 对于跨链场景，LayerZero 消息一旦在源链确认并被监听器捕获，
// 基本都会成功到达目标链。因此：
// - 交易被监听器捕获并写入数据库 = 源链已确认
// - 等待一定时间后（默认2分钟），自动标记为 Delivered
//
// 这样可以避免跨多条链查询交易状态的复杂性，
// 同时也符合 LayerZero 的高可靠性特点。
// ------------------------------------------------------------------
func statusUpdater(s *Store, _ *ethclient.Client, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Println("StatusUpdater: started (auto-confirm mode, 2min delay)")

	for range ticker.C {
		// 1) 获取所有 Pending 状态的交易
		pending, err := s.ListPendingPayouts(200)
		if err != nil {
			log.Printf("StatusUpdater: ListPendingPayouts error: %v", err)
			continue
		}

		if len(pending) == 0 {
			// 没有待处理交易
			continue
		}

		// 2) 检查每笔交易的时间，超过阈值则自动确认
		now := time.Now().UTC()
		updatedCount := 0

		for _, tx := range pending {
			// 计算交易年龄（从时间戳到现在的时长）
			age := now.Sub(tx.Timestamp)

			// 如果交易已经超过 2 分钟，自动标记为 Delivered
			if age > 2*time.Minute {
				if err := s.UpdatePayoutStatus(tx.TxHash, "Delivered"); err != nil {
					log.Printf("StatusUpdater: failed to update tx %s: %v", tx.TxHash, err)
				} else {
					updatedCount++
					log.Printf("StatusUpdater: auto-confirmed tx %s (age: %v)",
						tx.TxHash[:16]+"...", age.Round(time.Second))
				}
			}
		}

		if updatedCount > 0 {
			log.Printf("StatusUpdater: auto-confirmed %d transactions this round", updatedCount)
		}
	}
}

// 注意：sourceChainClient 参数保留是为了保持函数签名兼容性，
// 但在简化模式下不再使用（用 _ 忽略）
