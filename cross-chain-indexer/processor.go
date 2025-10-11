package main

import (
	"context"
	"log"
	"math/big"
	"time"

	"cross-chain-indexer/contract"

	"github.com/ethereum/go-ethereum/common"
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
	}

	topic := ""
	if len(vLog.Topics) > 0 {
		topic = vLog.Topics[0].Hex()
	}

	// 1) 插入事件去重（events 表）
	inserted, err := p.store.InsertEventIfNotExists(vLog.TxHash.Hex(), uint(vLog.Index), vLog.BlockNumber, topic, vLog.Data)
	if err != nil {
		log.Printf("processor: InsertEventIfNotExists error: %v", err)
		return err
	}
	if !inserted {
		// 已存在，说明已经处理或正在处理中，跳过
		return nil
	}

	// 2) 通过合约 binding 解析事件
	oapp, err := contract.NewMyOApp(common.HexToAddress(oappContractAddress), p.client)
	if err != nil {
		log.Printf("processor: NewMyOApp error: %v", err)
		// 返回错误，让调用方决定是否重试
		return err
	}

	event, err := oapp.ParseTokenPayoutRequested(vLog)
	if err != nil {
		log.Printf("processor: ParseTokenPayoutRequested failed for tx %s idx %d: %v", vLog.TxHash.Hex(), vLog.Index, err)
		return err
	}

	// 3) 尝试获取区块时间作为 timestamp（若失败则使用 now）
	var ts time.Time
	header, err := p.client.HeaderByNumber(ctx, new(big.Int).SetUint64(vLog.BlockNumber))
	if err != nil {
		// 取不到区块头时使用当前时间（并记录日志）
		log.Printf("processor: warning couldn't get header for block %d: %v", vLog.BlockNumber, err)
		ts = time.Now().UTC()
	} else {
		ts = time.Unix(int64(header.Time), 0).UTC()
	}

	// 4) 构造 PayoutRecord
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

	// 5) Upsert payout 到 DB
	if err := p.store.UpsertPayout(rec); err != nil {
		log.Printf("processor: UpsertPayout error for tx %s: %v", rec.TxHash, err)
		return err
	}

	// 6) 标记 event parsed
	if err := p.store.MarkEventParsed(vLog.TxHash.Hex(), uint(vLog.Index)); err != nil {
		log.Printf("processor: MarkEventParsed error for tx %s idx %d: %v", vLog.TxHash.Hex(), vLog.Index, err)
		// 已经写到 payouts，但 mark parsed 失败。返回错误以便调用者重试或记录。
		return err
	}

	log.Printf("processor: persisted payout tx=%s block=%d dstEid=%d net=%s", rec.TxHash, rec.BlockNumber, rec.DstEid, rec.NetAmount.String())
	return nil
}
