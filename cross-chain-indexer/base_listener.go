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

// BaseListener Base Sepolia 链监听器
type BaseListener struct {
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

// NewBaseListener 创建 Base Sepolia 监听器
func NewBaseListener(wssURL, httpsURL, contractAddr string, store *Store) (*BaseListener, error) {
	// 连接 WSS
	wssClient, err := ethclient.Dial(wssURL)
	if err != nil {
		log.Printf("BaseListener: WSS connection failed (%s): %v", wssURL, err)
		wssClient = nil // 允许继续，稍后重试
	}

	// 连接 HTTPS（必需）
	httpsClient, err := ethclient.Dial(httpsURL)
	if err != nil {
		return nil, fmt.Errorf("BaseListener: HTTPS connection failed (%s): %v", httpsURL, err)
	}

	// 创建 processor
	processor := NewProcessor(httpsClient, store)

	listener := &BaseListener{
		wssClient:        wssClient,
		httpsClient:      httpsClient,
		store:            store,
		processor:        processor,
		contractAddr:     common.HexToAddress(contractAddr),
		tokenPayoutTopic: common.HexToHash("0xd892a21f8b815c577e9ce52aa66d230fa1b28664b1286de9e4b85acfac750c31"),
		chainName:        "Base Sepolia",
		chainID:          40245, // EID_BASE_SEPOLIA
	}

	return listener, nil
}

// Start 启动监听器
func (bl *BaseListener) Start(ctx context.Context) error {
	log.Printf("BaseListener: Starting for contract %s", bl.contractAddr.Hex())

	// 1. 回填历史事件
	latestBlock := bl.backfillHistoricalEvents(ctx)
	log.Printf("BaseListener: Backfill completed, latest block: %d", latestBlock)

	// 2. 启动实时监听
	if bl.wssClient != nil {
		go bl.listenForNewEvents(ctx)
	} else {
		log.Println("BaseListener: WSS client not available, will only use polling")
		// 启动轮询模式
		go bl.pollForNewEvents(ctx, latestBlock)
	}

	return nil
}

// backfillHistoricalEvents 回填历史事件
func (bl *BaseListener) backfillHistoricalEvents(ctx context.Context) uint64 {
	// 获取最新区块
	latestBlockObj, err := bl.httpsClient.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Printf("BaseListener backfill: cannot get latest block: %v", err)
		return 0
	}
	latestBlock := latestBlockObj.Number.Uint64()

	// 从最近的区块开始扫描
	scanDepth := uint64(50000)
	fromBlock := uint64(0)
	if latestBlock > scanDepth {
		fromBlock = latestBlock - scanDepth
	}

	log.Printf("BaseListener backfill: scanning blocks [%d - %d] on %s", fromBlock, latestBlock, bl.chainName)

	// 构造查询
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(latestBlock)),
		Addresses: []common.Address{bl.contractAddr},
		Topics:    [][]common.Hash{{bl.tokenPayoutTopic}},
	}

	logs, err := bl.httpsClient.FilterLogs(ctx, query)
	if err != nil {
		log.Printf("BaseListener backfill: FilterLogs error: %v", err)
		return latestBlock
	}

	log.Printf("BaseListener backfill: found %d logs in [%d - %d]", len(logs), fromBlock, latestBlock)

	// 处理日志（倒序以保证时间顺序）
	for i := len(logs) - 1; i >= 0; i-- {
		if err := bl.parseAndPersist(ctx, logs[i]); err != nil {
			log.Printf("BaseListener backfill: parse error for tx %s: %v", logs[i].TxHash.Hex(), err)
		}
	}

	return latestBlock
}

// listenForNewEvents 实时监听新事件（WSS）
func (bl *BaseListener) listenForNewEvents(ctx context.Context) {
	log.Printf("BaseListener: Starting real-time listener for %s", bl.chainName)

	for {
		select {
		case <-ctx.Done():
			log.Println("BaseListener: Context cancelled, stopping listener")
			return
		default:
		}

		// 更新状态
		mu.Lock()
		baseWssStatus = "Connecting"
		mu.Unlock()

		// 订阅日志
		query := ethereum.FilterQuery{
			Addresses: []common.Address{bl.contractAddr},
			Topics:    [][]common.Hash{{bl.tokenPayoutTopic}},
		}

		logsCh := make(chan types.Log)
		sub, err := bl.wssClient.SubscribeFilterLogs(ctx, query, logsCh)
		if err != nil {
			log.Printf("BaseListener: SubscribeFilterLogs error: %v (will retry)", err)
			mu.Lock()
			baseWssStatus = "Disconnected"
			mu.Unlock()
			time.Sleep(10 * time.Second)
			continue
		}

		mu.Lock()
		baseWssStatus = "Connected"
		mu.Unlock()
		log.Printf("BaseListener: WSS subscription active for %s", bl.chainName)

		// 处理日志
		for {
			select {
			case <-ctx.Done():
				sub.Unsubscribe()
				return
			case err := <-sub.Err():
				log.Printf("BaseListener: subscription error: %v (reconnecting)", err)
				mu.Lock()
				baseWssStatus = "Disconnected"
				mu.Unlock()
				sub.Unsubscribe()
				time.Sleep(5 * time.Second)
				goto reconnect
			case vLog := <-logsCh:
				if err := bl.parseAndPersist(ctx, vLog); err != nil {
					log.Printf("BaseListener: parse error for tx %s: %v", vLog.TxHash.Hex(), err)
				}
			}
		}
	reconnect:
	}
}

// pollForNewEvents 轮询模式（当 WSS 不可用时）
func (bl *BaseListener) pollForNewEvents(ctx context.Context, startBlock uint64) {
	log.Printf("BaseListener: Starting polling mode from block %d", startBlock)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	currentBlock := startBlock

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 获取最新区块
			latestBlockObj, err := bl.httpsClient.HeaderByNumber(ctx, nil)
			if err != nil {
				log.Printf("BaseListener poll: cannot get latest block: %v", err)
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
				Addresses: []common.Address{bl.contractAddr},
				Topics:    [][]common.Hash{{bl.tokenPayoutTopic}},
			}

			logs, err := bl.httpsClient.FilterLogs(ctx, query)
			if err != nil {
				log.Printf("BaseListener poll: FilterLogs error: %v", err)
				continue
			}

			if len(logs) > 0 {
				log.Printf("BaseListener poll: found %d new logs in blocks [%d - %d]", len(logs), currentBlock+1, latestBlock)
				for _, vLog := range logs {
					if err := bl.parseAndPersist(ctx, vLog); err != nil {
						log.Printf("BaseListener poll: parse error for tx %s: %v", vLog.TxHash.Hex(), err)
					}
				}
			}

			currentBlock = latestBlock
		}
	}
}

// parseAndPersist 解析并持久化事件
func (bl *BaseListener) parseAndPersist(ctx context.Context, vLog types.Log) error {
	// 获取交易详情（用于验证交易是否确认）
	_, pending, err := bl.httpsClient.TransactionByHash(ctx, vLog.TxHash)
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

		log.Printf("BaseListener: Solana merchant address: %s (mapped to %s)",
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
	_, err = bl.httpsClient.TransactionReceipt(ctx, vLog.TxHash)
	if err != nil {
		return fmt.Errorf("get receipt failed: %v", err)
	}

	header, err := bl.httpsClient.HeaderByNumber(ctx, big.NewInt(int64(vLog.BlockNumber)))
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
	if err := bl.store.UpsertPayout(*record); err != nil {
		return fmt.Errorf("save payout failed: %v", err)
	}

	log.Printf("BaseListener: Saved payout from %s: tx=%s, payer=%s, merchant=%s, amount=%s -> EID:%d",
		bl.chainName,
		record.TxHash[:10]+"...",
		record.Payer.Hex()[:8]+"...",
		record.Merchant.Hex()[:8]+"...",
		formatAmount(grossAmount),
		dstEid,
	)

	return nil
}

// 全局 Base WSS 状态
var baseWssStatus string = "Disconnected"
