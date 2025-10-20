package main

import (
	"context"
	"encoding/json" // 新增
	"fmt"           // 新增
	"log"
	"math/big"
	"time"

	"cross-chain-indexer/contract"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Processor 负责解析 raw logs -> 业务记录，然后写入 Store（SQLite）
type Processor struct {
	client *ethclient.Client
	store  *Store
}

// NewProcessor 返回一个 Processor 实例
func NewProcessor(client *ethclient.Client, store *Store) *Processor {
	return &Processor{
		client: client,
		store:  store,
	}
}

// ParseAndPersist:
// 1) 将 raw log 写入 events 表（若已存在则跳过）
// 2) 使用合约 binding 解析 TokenPayoutRequested 事件
// 3) 将解析得到的业务记录 upsert 到 payouts 表
// 4) 将 events 标记为 parsed
func (p *Processor) ParseAndPersist(ctx context.Context, vLog types.Log) error {
	// 防御：必须有 topic
	if len(vLog.Topics) == 0 {
		log.Printf("processor: log has no topics, tx=%s idx=%d", vLog.TxHash.Hex(), vLog.Index)
		return nil // 忽略无 topic 的 log
	}

	// 1) 序列化 Log 为 JSON 字符串，作为原始数据存储
	rawLogBytes, err := json.Marshal(vLog)
	if err != nil {
		return fmt.Errorf("failed to marshal log to json: %w", err)
	}
	rawLog := string(rawLogBytes)

	// 2) 将原始 log 写入 events 表（若已存在则跳过）
	inserted, err := p.store.InsertEventIfNotExists(
		vLog.TxHash.Hex(),
		vLog.Index,
		vLog.BlockNumber,
		rawLog,
	)
	if err != nil {
		return fmt.Errorf("insert event failed: %w", err)
	}

	// 如果 log 已经存在（被跳过），则无需重复解析和处理
	if !inserted {
		return nil
	}

	// 3) 解析事件
	event, err := p.parseEvent(vLog)
	if err != nil {
		// 解析失败，但原始事件已存。继续处理下一个 log。
		log.Printf("processor: failed to parse log (tx=%s, idx=%d): %v", vLog.TxHash.Hex(), vLog.Index, err)
		return nil
	}

	// 4) 获取区块时间戳
	header, err := p.client.HeaderByNumber(ctx, new(big.Int).SetUint64(vLog.BlockNumber))
	var ts time.Time
	if err != nil || header == nil {
		// 取不到区块头时使用当前时间（并记录日志）
		log.Printf("processor: warning couldn't get header for block %d: %v", vLog.BlockNumber, err)
		ts = time.Now().UTC()
	} else {
		ts = time.Unix(int64(header.Time), 0).UTC()
	}

	// 5) 构造 PayoutRecord
	rec := PayoutRecord{
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: int64(vLog.BlockNumber),
		DstEid:      int64(event.DstEid),
		Payer:       event.Payer,
		Merchant:    event.Merchant,
		SrcToken:    event.SrcToken,
		DstToken:    event.DstToken,
		GrossAmount: event.GrossAmount,
		NetAmount:   event.NetAmount,
		Status:      "Pending",
		Timestamp:   ts,
	}

	// safety: if NetAmount nil, still proceed but log
	if rec.NetAmount == nil {
		log.Printf("processor: warning net amount nil for tx %s", rec.TxHash)
		rec.NetAmount = big.NewInt(0)
	}
	if rec.GrossAmount == nil {
		rec.GrossAmount = big.NewInt(0)
	}

	// 6) Upsert payout 到 DB
	if err := p.store.UpsertPayout(rec); err != nil {
		log.Printf("processor: failed to upsert payout for tx %s: %v", rec.TxHash, err)
		return fmt.Errorf("upsert payout failed: %w", err)
	}

	// 7) 标记 events 表中的 log 为已解析
	if err := p.store.MarkEventAsParsed(vLog.TxHash.Hex(), vLog.Index); err != nil {
		log.Printf("processor: failed to mark event as parsed for tx %s: %v", vLog.TxHash.Hex(), err)
	}

	return nil
}

// parseEvent 解析具体的 TokenPayoutRequested 事件
func (p *Processor) parseEvent(vLog types.Log) (*contract.MyOAppTokenPayoutRequested, error) {
	// ContractAddress hardcode，但在 main.go 中统一配置可能更好
	oappAddr := vLog.Address

	// 实例化合约绑定
	oapp, err := contract.NewMyOApp(oappAddr, p.client)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate contract: %w", err)
	}

	// 解析事件
	event, err := oapp.ParseTokenPayoutRequested(vLog)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TokenPayoutRequested event: %w", err)
	}

	return event, nil
}
