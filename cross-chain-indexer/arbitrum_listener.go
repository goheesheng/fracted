package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gagliardetto/solana-go"
)

// ArbitrumListener Arbitrum 链监听器
type ArbitrumListener struct {
	wssClient   *ethclient.Client
	httpsClient *ethclient.Client
	store       *Store
	processor   *Processor

	// 合约地址
	contractAddr common.Address

	// 事件 topic
	tokenPayoutTopic common.Hash

	// 链信息
	chainName string
	chainID   uint32
}

// NewArbitrumListener 创建 Arbitrum 监听器
func NewArbitrumListener(wssURL, httpsURL, contractAddr string, store *Store) (*ArbitrumListener, error) {
	// 连接 WSS
	wssClient, err := ethclient.Dial(wssURL)
	if err != nil {
		log.Printf("ArbitrumListener: WSS connection failed (%s): %v", wssURL, err)
		wssClient = nil // 允许继续，稍后重试
	}

	// 连接 HTTPS（必需）
	httpsClient, err := ethclient.Dial(httpsURL)
	if err != nil {
		return nil, fmt.Errorf("ArbitrumListener: HTTPS connection failed (%s): %v", httpsURL, err)
	}

	// 创建 processor
	processor := NewProcessor(httpsClient, store)

	listener := &ArbitrumListener{
		wssClient:        wssClient,
		httpsClient:      httpsClient,
		store:            store,
		processor:        processor,
		contractAddr:     common.HexToAddress(contractAddr),
		tokenPayoutTopic: common.HexToHash("0xd892a21f8b815c577e9ce52aa66d230fa1b28664b1286de9e4b85acfac750c31"),
		chainName:        "Arbitrum Sepolia",
		chainID:          40231, // EID_ARB_SEPOLIA
	}

	return listener, nil
}

// Start 启动监听器
func (al *ArbitrumListener) Start(ctx context.Context) error {
	log.Printf("ArbitrumListener: Starting for contract %s", al.contractAddr.Hex())

	// 1. 异步回填历史事件（不阻塞启动流程）
	go func() {
		latestBlock := al.backfillHistoricalEvents(ctx)
		log.Printf("ArbitrumListener: Backfill completed, latest block: %d", latestBlock)
	}()

	// 2. 启动实时监听
	if al.wssClient != nil {
		go al.listenForNewEvents(ctx)
	} else {
		log.Println("ArbitrumListener: WSS client not available, will only use polling")
		// 启动轮询模式（从当前区块开始）
		go func() {
			header, err := al.httpsClient.HeaderByNumber(ctx, nil)
			var latestBlock uint64
			if err == nil && header != nil && header.Number != nil {
				latestBlock = header.Number.Uint64()
			}
			al.pollForNewEvents(ctx, latestBlock)
		}()
	}

	return nil
}

// backfillHistoricalEvents 回填历史事件
func (al *ArbitrumListener) backfillHistoricalEvents(ctx context.Context) uint64 {
	// 获取最新区块
	latestBlockObj, err := al.httpsClient.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Printf("ArbitrumListener backfill: cannot get latest block: %v", err)
		return 0
	}
	latestBlock := latestBlockObj.Number.Uint64()

	// 从最近的区块开始扫描（Arbitrum 出块快，需要更大的深度）
	scanDepth := uint64(2000)
	fromBlock := uint64(0)
	if latestBlock > scanDepth {
		fromBlock = latestBlock - scanDepth
	}

	log.Printf("ArbitrumListener backfill: scanning blocks [%d - %d] on %s", fromBlock, latestBlock, al.chainName)

	// 构造查询
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(latestBlock)),
		Addresses: []common.Address{al.contractAddr},
		Topics:    [][]common.Hash{{al.tokenPayoutTopic}},
	}

	logs, err := al.httpsClient.FilterLogs(ctx, query)
	if err != nil {
		log.Printf("ArbitrumListener backfill: FilterLogs error: %v", err)
		return latestBlock
	}

	log.Printf("ArbitrumListener backfill: found %d logs in [%d - %d]", len(logs), fromBlock, latestBlock)

	// 处理日志（倒序以保证时间顺序）
	for i := len(logs) - 1; i >= 0; i-- {
		if err := al.parseAndPersist(ctx, logs[i]); err != nil {
			log.Printf("ArbitrumListener backfill: parse error for tx %s: %v", logs[i].TxHash.Hex(), err)
		}
	}

	return latestBlock
}

// listenForNewEvents 实时监听新事件（WSS）
func (al *ArbitrumListener) listenForNewEvents(ctx context.Context) {
	log.Printf("ArbitrumListener: Starting real-time listener for %s", al.chainName)

	for {
		select {
		case <-ctx.Done():
			log.Println("ArbitrumListener: Context cancelled, stopping listener")
			return
		default:
		}

		// 更新状态
		mu.Lock()
		arbWssStatus = "Connecting"
		mu.Unlock()

		// 订阅日志
		query := ethereum.FilterQuery{
			Addresses: []common.Address{al.contractAddr},
			Topics:    [][]common.Hash{{al.tokenPayoutTopic}},
		}

		logsCh := make(chan types.Log)
		sub, err := al.wssClient.SubscribeFilterLogs(ctx, query, logsCh)
		if err != nil {
			log.Printf("ArbitrumListener: SubscribeFilterLogs error: %v (will retry)", err)
			mu.Lock()
			arbWssStatus = "Disconnected"
			mu.Unlock()
			time.Sleep(10 * time.Second)
			continue
		}

		mu.Lock()
		arbWssStatus = "Connected"
		mu.Unlock()
		log.Printf("ArbitrumListener: WSS subscription active for %s", al.chainName)

		// 处理日志
		for {
			select {
			case <-ctx.Done():
				sub.Unsubscribe()
				return
			case err := <-sub.Err():
				log.Printf("ArbitrumListener: subscription error: %v (reconnecting)", err)
				mu.Lock()
				arbWssStatus = "Disconnected"
				mu.Unlock()
				sub.Unsubscribe()
				time.Sleep(5 * time.Second)
				goto reconnect
			case vLog := <-logsCh:
				if err := al.parseAndPersist(ctx, vLog); err != nil {
					log.Printf("ArbitrumListener: parse error for tx %s: %v", vLog.TxHash.Hex(), err)
				}
			}
		}
	reconnect:
	}
}

// pollForNewEvents 轮询模式（当 WSS 不可用时）
func (al *ArbitrumListener) pollForNewEvents(ctx context.Context, startBlock uint64) {
	log.Printf("ArbitrumListener: Starting polling mode from block %d", startBlock)

	ticker := time.NewTicker(15 * time.Second) // Arbitrum 出块快，可以更频繁轮询
	defer ticker.Stop()

	currentBlock := startBlock

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 获取最新区块
			latestBlockObj, err := al.httpsClient.HeaderByNumber(ctx, nil)
			if err != nil {
				log.Printf("ArbitrumListener poll: cannot get latest block: %v", err)
				continue
			}
			latestBlock := latestBlockObj.Number.Uint64()

			if latestBlock <= currentBlock {
				continue
			}

			// 查询新区块
			query := ethereum.FilterQuery{
				FromBlock: big.NewInt(int64(currentBlock + 1)),
				ToBlock:   big.NewInt(int64(latestBlock)),
				Addresses: []common.Address{al.contractAddr},
				Topics:    [][]common.Hash{{al.tokenPayoutTopic}},
			}

			logs, err := al.httpsClient.FilterLogs(ctx, query)
			if err != nil {
				log.Printf("ArbitrumListener poll: FilterLogs error: %v", err)
				continue
			}

			if len(logs) > 0 {
				log.Printf("ArbitrumListener poll: found %d new logs in blocks [%d - %d]", len(logs), currentBlock+1, latestBlock)
				for _, vLog := range logs {
					if err := al.parseAndPersist(ctx, vLog); err != nil {
						log.Printf("ArbitrumListener poll: parse error for tx %s: %v", vLog.TxHash.Hex(), err)
					}
				}
			}

			currentBlock = latestBlock
		}
	}
}

// parseAndPersist 解析并持久化事件
func (al *ArbitrumListener) parseAndPersist(ctx context.Context, vLog types.Log) error {
	// 使用 processor 解析事件
	// 注意：需要标记这是从 Arbitrum 来的交易

	// 获取交易详情（用于验证交易是否确认）
	_, pending, err := al.httpsClient.TransactionByHash(ctx, vLog.TxHash)
	if err != nil {
		return fmt.Errorf("get transaction failed: %v", err)
	}
	if pending {
		return fmt.Errorf("transaction is pending")
	}

	// 解析事件
	// event TokenPayoutRequested(
	//     uint32 indexed dstEid,
	//     address indexed payer,
	//     address indexed merchant,
	//     address srcToken,
	//     address dstToken,
	//     uint256 grossAmount,
	//     uint256 netAmount,
	//     uint256 feeAmount
	// )

	if len(vLog.Topics) < 4 {
		return fmt.Errorf("invalid log topics length: %d", len(vLog.Topics))
	}

	// Topics[0] = event signature
	// Topics[1] = dstEid (indexed)
	// Topics[2] = payer (indexed)
	// Topics[3] = merchant (indexed)

	dstEid := uint32(vLog.Topics[1].Big().Uint64())
	payer := common.BytesToAddress(vLog.Topics[2].Bytes())

	// merchant 地址处理：根据目标链类型决定如何解析
	var merchant common.Address
	var solanaMerchant string

	// 检查目标链是否为 Solana (EID 40168 = Solana Devnet, 30168 = Solana Mainnet)
	isSolanaDestination := (dstEid == 40168 || dstEid == 30168)

	if isSolanaDestination {
		// Solana 地址：Topics[3] 是完整的 32 字节 Solana 公钥
		merchantBytes32 := vLog.Topics[3]

		// 转换为 Solana Base58 地址
		solPubkey := solana.PublicKeyFromBytes(merchantBytes32.Bytes())
		solanaMerchant = solPubkey.String()

		// 为了数据库兼容性，也存储一个映射的 EVM 地址（使用后20字节）
		merchant = common.BytesToAddress(merchantBytes32.Bytes())

		log.Printf("ArbitrumListener: Solana merchant address: %s (mapped to %s)",
			solanaMerchant, merchant.Hex())
	} else {
		// EVM 地址：Topics[3] 是标准的 EVM 地址（取后20字节）
		merchant = common.BytesToAddress(vLog.Topics[3].Bytes())
	}

	// 解析 data
	// data = (srcToken, dstToken, grossAmount, netAmount, feeAmount)
	if len(vLog.Data) < 160 { // 5 * 32 bytes
		return fmt.Errorf("invalid log data length: %d", len(vLog.Data))
	}

	srcToken := common.BytesToAddress(vLog.Data[0:32])
	dstToken := common.BytesToAddress(vLog.Data[32:64])
	grossAmount := new(big.Int).SetBytes(vLog.Data[64:96])
	netAmount := new(big.Int).SetBytes(vLog.Data[96:128])
	// feeAmount := new(big.Int).SetBytes(vLog.Data[128:160]) // 暂时不使用

	// 获取区块信息
	_, err = al.httpsClient.TransactionReceipt(ctx, vLog.TxHash)
	if err != nil {
		return fmt.Errorf("get receipt failed: %v", err)
	}

	header, err := al.httpsClient.HeaderByNumber(ctx, big.NewInt(int64(vLog.BlockNumber)))
	if err != nil {
		return fmt.Errorf("get block header failed: %v", err)
	}

	timestamp := time.Unix(int64(header.Time), 0).UTC()

	// 构造 PayoutRecord 对象
	record := &PayoutRecord{
		TxHash:         strings.ToLower(vLog.TxHash.Hex()),
		BlockNumber:    int64(vLog.BlockNumber),
		DstEid:         int64(dstEid),
		Payer:          payer,
		Merchant:       merchant,
		SrcToken:       srcToken,
		DstToken:       dstToken,
		GrossAmount:    grossAmount,
		NetAmount:      netAmount,
		Status:         "Pending",
		Timestamp:      timestamp,
		SolanaMerchant: solanaMerchant, // 保存 Solana 原始地址
	}

	// 保存到数据库（UpsertPayout 接受值类型，不是指针）
	if err := al.store.UpsertPayout(*record); err != nil {
		return fmt.Errorf("save payout failed: %v", err)
	}

	log.Printf("ArbitrumListener: Saved payout from %s: tx=%s, payer=%s, merchant=%s, amount=%s -> EID:%d",
		al.chainName,
		record.TxHash[:10]+"...",
		record.Payer.Hex()[:8]+"...",
		record.Merchant.Hex()[:8]+"...",
		formatAmount(grossAmount),
		dstEid,
	)

	return nil
}

// getChainNameByEID 根据 EID 获取链名称
func getChainNameByEID(eid uint32) string {
	switch eid {
	case 40245:
		return "base-sepolia"
	case 40231:
		return "arbitrum-sepolia"
	case 40168:
		return "solana-devnet"
	case 30168:
		return "solana-mainnet"
	default:
		return fmt.Sprintf("unknown-eid-%d", eid)
	}
}

// 全局 Arbitrum WSS 状态
var arbWssStatus string = "Disconnected"
