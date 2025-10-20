package main

import (
	"context"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// statusUpdater 周期性扫描 store 中 Pending 的 tx，并用 sourceChainClient 检查 tx receipt。
// 如果回执存在且 receipt.Status == 1 => Delivered；如果存在且 == 0 => Failed。
// 请在 main.go 中用 go statusUpdater(store, httpsClient, 15*time.Second) 启动。
// ------------------------------------------------------------------
func statusUpdater(s *Store, sourceChainClient *ethclient.Client, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for range ticker.C {
		// 1) 获取待处理项（限制 N 条以免一次性太多）
		pending, err := s.ListPendingPayouts(200)
		if err != nil {
			log.Printf("statusUpdater: ListPendingPayouts error: %v", err)
			continue
		}
		if len(pending) == 0 {
			// nothing to do
			continue
		}
		// 2) 遍历并检查 tx 回执
		for _, row := range pending {
			txHash := row.TxHash
			// Query receipt (non-blocking context with small timeout)
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			receipt, err := sourceChainClient.TransactionReceipt(ctx, common.HexToHash(txHash))
			cancel()
			if err != nil {
				// 回执不存在或 RPC 超时：不把它标为失败，继续下次重试
				// 记录调试日志（但避免过于频繁日志）
				// log.Printf("statusUpdater: receipt not ready for tx %s: %v", txHash, err)
				continue
			}
			// 有回执，根据 receipt.Status 更新
			if receipt.Status == 1 {
				// 成功
				if err := s.UpdatePayoutStatus(txHash, "Delivered"); err != nil {
					log.Printf("statusUpdater: UpdatePayoutStatus Delivered error tx=%s: %v", txHash, err)
				} else {
					log.Printf("statusUpdater: marked Delivered tx=%s", txHash)
				}
			} else {
				// 失败
				if err := s.UpdatePayoutStatus(txHash, "Failed"); err != nil {
					log.Printf("statusUpdater: UpdatePayoutStatus Failed error tx=%s: %v", txHash, err)
				} else {
					log.Printf("statusUpdater: marked Failed tx=%s", txHash)
				}
			}
		}
	}
}
