package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	// 如果你的 contract 包路径不同（参见 go.mod 的 module），请修改下面这行

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

var jwtSecret = []byte(getEnvOrDefault("JWT_SECRET", "dev-local-secret-change-me"))

// --------------------------- CONFIG (请按需替换) ---------------------------
const (
	// 【必填】用于实时监听的 WSS RPC（你的 Base Sepolia WSS）
	baseSepoliaWSS = "wss://base-sepolia.publicnode.com"

	// 【必填】用于历史查询的 HTTPS RPC
	baseSepoliaHTTPS = "https://base-sepolia.publicnode.com"

	// 可选：用于目标链交付状态检查（Arbitrum Sepolia）
	arbSepoliaHTTPS = "https://arbitrum-sepolia.publicnode.com"

	// Solana Devnet RPC
	solanaDevnetRPC = "https://api.devnet.solana.com"

	// 【必填】Base Sepolia 合约地址（MyOApp）
	// 旧合约地址（保留用于历史数据）
	oappContractAddress = "0x6689F160b47CbfEBf389c55ae34959296Ef56B8D"

	// 新的 Base Sepolia 合约地址（用于 Base -> Arb/Solana 跨链）
	baseContractAddress = "0xA1D91CdcBD933c3385D7dea34D87357f5E62f6d6"

	// Arbitrum Sepolia 合约地址（MyOApp）
	arbContractAddress = "0x1a9C0a66Cb68D92c598B0D2f10de3C755Eb6D438"

	// Solana Devnet 程序地址 (transfer_contract)
	// Solana transfer_contract 程序地址
	solanaProgramAddress = "GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1"

	// 大多数 USDT/USDC 使用 6 位小数
	TokenDecimals = 6.0

	// LayerZero Endpoint IDs
	EID_BASE_SEPOLIA   = 40245
	EID_ARB_SEPOLIA    = 40231
	EID_SOLANA_DEVNET  = 40168
	EID_SOLANA_MAINNET = 30168
)

// TokenPayoutRequested event topic (请确保该 topic 与合约一致)
var tokenPayoutRequestedTopic = common.HexToHash("0xdd9e34114af31ed8b7896e826d4d77f69661c83c3fb0dfde856e2de117034601")

// --------------------------- 全局状态 ---------------------------
var (
	wssStatus string = "Disconnected"
	mu        sync.Mutex
)

// --------------------------- helper: clear screen ---------------------------
func clearScreen() {
	// 简单清屏（大多数终端支持）
	fmt.Print("\033[H\033[2J")
}

// formatAmount: 把 big.Int 格式化为可读字符串
func formatAmount(amount *big.Int) string {
	if amount == nil {
		return "N/A"
	}
	f := new(big.Float).SetInt(amount)
	divisor := big.NewFloat(math.Pow(10, TokenDecimals))
	f.Quo(f, divisor)
	if f.Cmp(big.NewFloat(1000000)) >= 0 {
		f.Quo(f, big.NewFloat(1000000))
		return fmt.Sprintf("%.2f M", f)
	}
	if f.Cmp(big.NewFloat(1000)) >= 0 {
		f.Quo(f, big.NewFloat(1000))
		return fmt.Sprintf("%.2f K", f)
	}
	return fmt.Sprintf("%.2f", f)
}

// --------------------------- Listener (带重连) ---------------------------
func listenForNewEvents(ctx context.Context, client *ethclient.Client, oappAddress common.Address, eventTopic common.Hash, proc *Processor) {
	backoff := 1 * time.Second
	for {
		select {
		case <-ctx.Done():
			log.Println("listener: context cancelled, exiting")
			return
		default:
		}

		mu.Lock()
		wssStatus = "Connecting"
		mu.Unlock()

		query := ethereum.FilterQuery{
			Addresses: []common.Address{oappAddress},
			Topics:    [][]common.Hash{{eventTopic}},
		}
		logsCh := make(chan types.Log)
		sub, err := client.SubscribeFilterLogs(ctx, query, logsCh)
		if err != nil {
			// 不致命：打日志并重试（指数退避）
			errMsg := err.Error()
			if len(errMsg) > 80 {
				errMsg = errMsg[:80] + "..."
			}
			mu.Lock()
			wssStatus = fmt.Sprintf("Error: %s", errMsg)
			mu.Unlock()
			log.Printf("listener: subscribe error: %v — retrying in %s", err, backoff)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > 60*time.Second {
				backoff = 60 * time.Second
			}
			continue
		}

		// 成功订阅
		backoff = 1 * time.Second
		mu.Lock()
		wssStatus = "Connected"
		mu.Unlock()
		log.Println("listener: subscribed to logs")

		// 接收循环
	loop:
		for {
			select {
			case <-ctx.Done():
				sub.Unsubscribe()
				break loop
			case err := <-sub.Err():
				log.Printf("listener: subscription error: %v — will reconnect", err)
				sub.Unsubscribe()
				break loop
			case vLog := <-logsCh:
				// 非阻塞地把 log 交给 processor 去处理
				go func(l types.Log) {
					if err := proc.ParseAndPersist(context.Background(), l); err != nil {
						log.Printf("processor error: %v", err)
					}
				}(vLog)
			}
		}

		// 小延迟后重连
		time.Sleep(2 * time.Second)
	}
}

// --------------------------- Backfill (历史回溯) ---------------------------
func backfillHistoricalEvents(client *ethclient.Client, oappAddress common.Address, eventTopic common.Hash, proc *Processor) uint64 {
	log.Println("backfill: starting historical scan (50000 blocks)")
	ctx := context.Background()
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Printf("backfill: HeaderByNumber error: %v", err)
		return 0
	}
	latestBlock := header.Number.Uint64()
	const backfillBlocks = 50000
	var fromBlock uint64
	if latestBlock > backfillBlocks {
		fromBlock = latestBlock - backfillBlocks
	} else {
		fromBlock = 0
	}

	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(latestBlock),
		Addresses: []common.Address{oappAddress},
		Topics:    [][]common.Hash{{eventTopic}},
	}
	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		log.Printf("backfill: FilterLogs error: %v", err)
		return latestBlock
	}
	log.Printf("backfill: found %d logs in [%d - %d]", len(logs), fromBlock, latestBlock)
	// 倒序处理以保证时间顺序
	for i := len(logs) - 1; i >= 0; i-- {
		if err := proc.ParseAndPersist(ctx, logs[i]); err != nil {
			log.Printf("backfill: parse error for tx %s idx %d: %v", logs[i].TxHash.Hex(), logs[i].Index, err)
		}
	}
	return latestBlock
}

// --------------------------- Delivery worker (占位) ---------------------------
// 这里保留最小实现：周期扫描 DB 的 Pending（真实环境应到目标链检查）
func trackDeliveryStatus(store *Store) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// MVP: 仅打印提醒。后续可实现：扫描 payouts WHERE status='Pending'
		// 对接目标链(arbitrum)事件并更新为 Delivered/Failed。
		log.Println("worker: tick - delivery checker placeholder")
	}
}

// --------------------------- Dashboard 简化渲染 ---------------------------
func renderDashboard(latestBlock uint64) {
	clearScreen()
	mu.Lock()
	status := wssStatus
	mu.Unlock()
	var colored string
	switch status {
	case "Connected":
		colored = "\033[32mConnected\033[0m"
	case "Disconnected":
		colored = "\033[31mDisconnected\033[0m"
	case "Connecting":
		colored = "\033[33mConnecting\033[0m"
	default:
		colored = status
	}
	fmt.Println(strings.Repeat("=", 110))
	fmt.Printf("LayerZero Cross-Chain Indexer | WSS: %s | Contract: %s\n", colored, oappContractAddress)
	fmt.Printf("Latest Block (Base Sepolia HTTPS): %d\n", latestBlock)
	fmt.Println(strings.Repeat("=", 110))
	fmt.Printf("API: http://localhost:8080/health  |  Payouts DB: indexer.db\n")
	fmt.Println()
	fmt.Println("Logs printed below (press Ctrl+C to stop):")
}

// --------------------------- main ---------------------------
func main() {
	// 随机种子（若 later 使用随机模拟）
	rand.Seed(time.Now().UnixNano())

	// 1) 初始化 Store (SQLite)
	dbPath := "indexer.db"
	store, err := NewStore(dbPath)
	if err != nil {
		log.Fatalf("main: NewStore error: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("main: store close error: %v", err)
		}
	}()

	// 2) 初始化 ETH clients
	wssClient, err := ethclient.Dial(baseSepoliaWSS)
	if err != nil {
		log.Printf("main: cannot connect WSS RPC (%s): %v", baseSepoliaWSS, err)
		// 不 fatal，listener 会重试
		wssClient = nil
	} else {
		log.Println("main: WSS client ready")
	}

	httpsClient, err := ethclient.Dial(baseSepoliaHTTPS)
	if err != nil {
		log.Fatalf("main: cannot connect HTTPS RPC (%s): %v", baseSepoliaHTTPS, err)
	}
	log.Println("main: HTTPS client ready")

	// 3) 初始化 Base Sepolia 监听器（新合约地址）
	log.Println("main: Initializing Base Sepolia listener...")
	baseListener, err := NewBaseListener(
		baseSepoliaWSS,
		baseSepoliaHTTPS,
		baseContractAddress,
		store,
	)
	if err != nil {
		log.Printf("main: failed to create Base listener: %v", err)
		baseListener = nil
	} else {
		log.Println("main: Base listener created successfully")
	}

	// 4) 初始化 Arbitrum 监听器
	log.Println("main: Initializing Arbitrum listener...")
	arbListener, err := NewArbitrumListener(
		"wss://arbitrum-sepolia.publicnode.com",
		arbSepoliaHTTPS,
		arbContractAddress,
		store,
	)
	if err != nil {
		log.Printf("main: failed to create Arbitrum listener: %v", err)
		arbListener = nil
	} else {
		log.Println("main: Arbitrum listener created successfully")
	}

	// 5) Processor
	proc := NewProcessor(httpsClient, store)

	// 6) Backfill Base Sepolia events once (旧合约地址，保留用于历史数据)
	oappAddr := common.HexToAddress(oappContractAddress)
	latestBlock := backfillHistoricalEvents(httpsClient, oappAddr, tokenPayoutRequestedTopic, proc)

	// 7) Start Base Sepolia WSS listener (if wss client created) - 旧合约
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if wssClient != nil {
		go listenForNewEvents(ctx, wssClient, oappAddr, tokenPayoutRequestedTopic, proc)
	} else {
		mu.Lock()
		wssStatus = "Disconnected"
		mu.Unlock()
	}

	// 8) Start Base Sepolia listener (新合约) - if created successfully
	if baseListener != nil {
		log.Println("main: Starting Base Sepolia listener...")
		if err := baseListener.Start(ctx); err != nil {
			log.Printf("main: Base listener start error: %v", err)
		}
	}

	// 9) Start Arbitrum listener (if created successfully)
	if arbListener != nil {
		log.Println("main: Starting Arbitrum listener...")
		if err := arbListener.Start(ctx); err != nil {
			log.Printf("main: Arbitrum listener start error: %v", err)
		}
	}

	// 10) 初始化并启动 Solana 监听器
	solanaListener, err := NewSolanaListener(solanaDevnetRPC, solanaProgramAddress, store)
	if err != nil {
		log.Printf("main: failed to create Solana listener: %v", err)
	} else {
		log.Println("main: Solana listener created")

		// 回填 Solana 历史交易（最近 100 笔）
		go func() {
			if err := solanaListener.BackfillHistoricalTransactions(ctx, 100); err != nil {
				log.Printf("main: Solana backfill error: %v", err)
			}
		}()

		// 启动实时监听（带重连机制）
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if err := solanaListener.ListenForNewTransactions(ctx); err != nil {
					log.Printf("main: Solana listener error: %v, reconnecting in 5s...", err)
					time.Sleep(5 * time.Second)
				}
			}
		}()
	}

	// 11) Start delivery worker (placeholder)
	go trackDeliveryStatus(store)
	// 启动状态更新器：每 15 秒检查一次 Pending（你可以根据需要调整间隔）
	go statusUpdater(store, httpsClient, 15*time.Second)

	// 12) Start API server (api.go must provide NewServer)
	server := NewServer(store, httpsClient, oappAddr, tokenPayoutRequestedTopic, proc)
	go func() {
		addr := ":8080"
		log.Printf("main: starting API at %s", addr)
		if err := http.ListenAndServe(addr, server.routes()); err != nil {
			log.Fatalf("main: API ListenAndServe error: %v", err)
		}
	}()

	// 13) Dashboard refresh loop
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// try to refresh latest block
		header, err := httpsClient.HeaderByNumber(context.Background(), nil)
		if err == nil && header != nil && header.Number != nil {
			latestBlock = header.Number.Uint64()
		}
		renderDashboard(latestBlock)

		// Print a small memory hint / pid for debugging (optional)
		fmt.Printf("pid=%d  time=%s\n", os.Getpid(), time.Now().Format("2006-01-02 15:04:05"))
	}
}
