package main

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

// Solana 日志文件
const solanaLogFile = "solana_log.txt"

// convertToWebSocketURL 将 HTTP(S) URL 转换为 WS(S) URL
func convertToWebSocketURL(httpURL string) string {
	wsURL := strings.Replace(httpURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	return wsURL
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// logToFile 将日志写入文件
func logToFile(message string) {
	f, err := os.OpenFile(solanaLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)
	if _, err := f.WriteString(logLine); err != nil {
		log.Printf("Error writing to log file: %v", err)
	}
}

// SolanaListener 负责监听 Solana 程序的交易
type SolanaListener struct {
	rpcURL      string
	programAddr solana.PublicKey
	store       *Store
}

// TransferOutInstruction transfer_out 指令数据结构
type TransferOutInstruction struct {
	Discriminator [8]byte // Anchor 指令 discriminator
	Amount        uint64  // 转账金额
}

// NewSolanaListener 创建 Solana 监听器
func NewSolanaListener(rpcURL string, programAddrStr string, store *Store) (*SolanaListener, error) {
	programAddr, err := solana.PublicKeyFromBase58(programAddrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Solana program address: %w", err)
	}

	return &SolanaListener{
		rpcURL:      rpcURL,
		programAddr: programAddr,
		store:       store,
	}, nil
}

// BackfillHistoricalTransactions 回填历史交易
func (l *SolanaListener) BackfillHistoricalTransactions(ctx context.Context, limit int) error {
	log.Printf("Solana backfill: starting (limit: %d), program: %s", limit, l.programAddr.String())

	client := rpc.New(l.rpcURL)

	// 获取程序的签名列表
	sigs, err := client.GetSignaturesForAddress(ctx, l.programAddr)
	if err != nil {
		log.Printf("Solana backfill ERROR: failed to get signatures: %v", err)
		return fmt.Errorf("failed to get signatures: %w", err)
	}

	log.Printf("Solana backfill: found %d total transactions", len(sigs))

	successCount := 0
	errorCount := 0
	parsedCount := 0

	// 处理每个交易
	for i, sig := range sigs {
		if i >= limit {
			break
		}

		if sig.Err != nil {
			// 跳过失败的交易
			errorCount++
			continue
		}

		// 获取交易详情
		maxVer := uint64(0)
		tx, err := client.GetTransaction(ctx, sig.Signature, &rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase64,
			MaxSupportedTransactionVersion: &maxVer,
		})
		if err != nil {
			log.Printf("Solana backfill: failed to get transaction %s: %v", sig.Signature, err)
			errorCount++
			continue
		}

		if tx == nil || tx.Meta == nil {
			errorCount++
			continue
		}

		// 解析交易并保存
		if err := l.parseAndStore(tx, sig); err != nil {
			if err.Error() != "not a relevant transaction" {
				log.Printf("Solana backfill: parse error tx %s: %v", sig.Signature, err)
			}
			continue
		}

		parsedCount++
		successCount++

		// 进度显示
		if (i+1)%10 == 0 {
			log.Printf("Solana backfill: processed %d/%d, indexed: %d", i+1, len(sigs), parsedCount)
		}
	}

	log.Printf("Solana backfill: completed! Total: %d, Indexed: %d, Errors: %d", len(sigs), parsedCount, errorCount)
	return nil
}

// ListenForNewTransactions 实时监听新交易
func (l *SolanaListener) ListenForNewTransactions(ctx context.Context) error {
	log.Println("Solana listener: starting WebSocket connection")

	// 将 HTTPS URL 转换为 WSS URL
	wsURL := convertToWebSocketURL(l.rpcURL)
	log.Printf("Solana listener: connecting to %s", wsURL)

	// 创建 WebSocket 客户端
	wsClient, err := ws.Connect(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect WebSocket: %w", err)
	}
	defer wsClient.Close()

	// 订阅程序日志
	sub, err := wsClient.LogsSubscribeMentions(l.programAddr, rpc.CommitmentFinalized)
	if err != nil {
		return fmt.Errorf("failed to subscribe logs: %w", err)
	}
	defer sub.Unsubscribe()

	log.Println("Solana listener: subscribed to program logs")

	// 接收循环
	for {
		select {
		case <-ctx.Done():
			return nil
		case result := <-sub.Response():
			if result == nil {
				continue
			}

			// 处理日志
			go l.handleLogNotification(result)
		case err := <-sub.Err():
			if err != nil {
				log.Printf("Solana listener: subscription error: %v", err)
				return err
			}
		}
	}
}

// handleLogNotification 处理实时日志通知
func (l *SolanaListener) handleLogNotification(result *ws.LogResult) {
	if result.Value.Err != nil {
		// 跳过失败的交易
		return
	}

	// 获取交易签名
	sig := result.Value.Signature

	// 获取交易详情
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := rpc.New(l.rpcURL)
	maxVer := uint64(0)
	tx, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Encoding:                       solana.EncodingBase64,
		MaxSupportedTransactionVersion: &maxVer,
	})
	if err != nil {
		log.Printf("Solana listener: failed to get transaction %s: %v", sig, err)
		return
	}

	if tx == nil {
		return
	}

	// 解析并保存
	sigInfo := &rpc.TransactionSignature{
		Signature: sig,
		BlockTime: tx.BlockTime,
		Slot:      tx.Slot,
	}

	if err := l.parseAndStore(tx, sigInfo); err != nil {
		log.Printf("Solana listener: failed to parse tx %s: %v", sig, err)
	}
}

// parseAndStore 解析 Solana 交易并存储到数据库
func (l *SolanaListener) parseAndStore(tx *rpc.GetTransactionResult, sig *rpc.TransactionSignature) error {
	if tx == nil || tx.Meta == nil {
		return fmt.Errorf("invalid transaction data")
	}

	// 提取基本信息
	txHash := sig.Signature.String()
	blockTime := time.Unix(int64(*sig.BlockTime), 0).UTC()
	slot := sig.Slot

	logMsg := fmt.Sprintf("Processing tx %s (slot: %d)", txHash[:min(20, len(txHash))], slot)
	log.Println("Solana: " + logMsg)
	logToFile(logMsg)

	// 解析交易数据
	txParsed, err := tx.Transaction.GetTransaction()
	if err != nil {
		return fmt.Errorf("failed to parse transaction: %w", err)
	}
	if txParsed == nil {
		return fmt.Errorf("transaction is nil")
	}

	// 遍历所有指令，查找 transfer_out
	message := txParsed.Message
	accounts := message.AccountKeys

	var transferOutFound bool
	var authority string
	var recipient string
	var mint string
	var amount uint64

	// 解析指令
	for idx, instruction := range message.Instructions {
		// 检查是否是我们的程序
		if int(instruction.ProgramIDIndex) >= len(accounts) {
			continue
		}
		programID := accounts[instruction.ProgramIDIndex]
		if programID.String() != l.programAddr.String() {
			continue
		}

		// 解析指令数据
		instructionData := instruction.Data
		if len(instructionData) < 16 {
			// transfer_out 指令至少需要 8字节 discriminator + 8字节 amount
			continue
		}

		// 检查是否是 transfer_out 指令
		// Anchor 的 transfer_out discriminator 可以从 IDL 计算，或通过日志识别
		// 这里我们先通过日志消息确认
		isTransferOut := false
		for _, logLine := range tx.Meta.LogMessages {
			if strings.Contains(logLine, "Instruction: TransferOut") ||
				strings.Contains(logLine, "Program log: transfer_out") {
				isTransferOut = true
				break
			}
		}

		if !isTransferOut {
			// 也可以通过检查账户数量和结构来判断
			// transfer_out 需要 7 个账户
			if len(instruction.Accounts) == 7 {
				// 可能是 transfer_out
				isTransferOut = true
			} else {
				continue
			}
		}

		transferOutFound = true

		// 解析金额（跳过 8 字节 discriminator）
		if len(instructionData) >= 16 {
			amount = binary.LittleEndian.Uint64(instructionData[8:16])
		}

		// 解析账户（按照 TransferOut 结构）
		// 0: config (PDA)
		// 1: authority (Signer) - 调用方
		// 2: vault_authority (PDA)
		// 3: vault_token_account (mut)
		// 4: recipient_token_account (mut)
		// 5: mint
		// 6: token_program

		if len(instruction.Accounts) >= 7 {
			authorityIdx := instruction.Accounts[1]
			recipientTokenAccountIdx := instruction.Accounts[4]
			mintIdx := instruction.Accounts[5]

			if int(authorityIdx) < len(accounts) {
				authority = accounts[authorityIdx].String()
			}
			if int(mintIdx) < len(accounts) {
				mint = accounts[mintIdx].String()
			}

			// recipient 是 recipient_token_account 的 owner
			// 需要从 Token Balance 中获取
			if tx.Meta.PostTokenBalances != nil {
				for _, bal := range tx.Meta.PostTokenBalances {
					if bal.AccountIndex == uint16(recipientTokenAccountIdx) {
						if bal.Owner != nil {
							recipient = bal.Owner.String()
						}
						// 如果指令数据中的金额为0，从 balance 变化中获取
						if amount == 0 && bal.UiTokenAmount.Amount != "" {
							if amt, ok := new(big.Int).SetString(bal.UiTokenAmount.Amount, 10); ok {
								amount = amt.Uint64()
							}
						}
						break
					}
				}
			}
		}

		logMsg := fmt.Sprintf("Found transfer_out instruction #%d: amount=%d, authority=%s, recipient=%s, mint=%s",
			idx, amount, authority[:min(10, len(authority))], recipient[:min(10, len(recipient))], mint[:min(10, len(mint))])
		log.Println("Solana: " + logMsg)
		logToFile(logMsg)

		break // 只处理第一个 transfer_out 指令
	}

	if !transferOutFound {
		return fmt.Errorf("not a transfer_out transaction")
	}

	// 验证必需字段
	if recipient == "" {
		// 尝试从 Token Balance 变化中推断 recipient
		if tx.Meta.PreTokenBalances != nil && tx.Meta.PostTokenBalances != nil {
			for i, postBal := range tx.Meta.PostTokenBalances {
				if i < len(tx.Meta.PreTokenBalances) {
					preBal := tx.Meta.PreTokenBalances[i]
					if postBal.UiTokenAmount.UiAmount != nil && preBal.UiTokenAmount.UiAmount != nil {
						diff := *postBal.UiTokenAmount.UiAmount - *preBal.UiTokenAmount.UiAmount
						if diff > 0 && postBal.Owner != nil {
							// 余额增加的账户就是接收方
							recipient = postBal.Owner.String()
							if amount == 0 && postBal.UiTokenAmount.Amount != "" {
								if amt, ok := new(big.Int).SetString(postBal.UiTokenAmount.Amount, 10); ok {
									// 这里获取的是总余额，不是转账金额
									// 需要计算差值
									if preAmt, ok2 := new(big.Int).SetString(preBal.UiTokenAmount.Amount, 10); ok2 {
										diff := new(big.Int).Sub(amt, preAmt)
										amount = diff.Uint64()
									}
								}
							}
							break
						}
					}
				}
			}
		}
	}

	if recipient == "" {
		logMsg := fmt.Sprintf("Cannot extract recipient from tx %s", txHash[:min(20, len(txHash))])
		log.Println("Solana: " + logMsg)
		logToFile(logMsg)
		return fmt.Errorf("no recipient found")
	}

	if amount == 0 {
		logMsg := fmt.Sprintf("Warning: zero amount in tx %s", txHash[:min(20, len(txHash))])
		log.Println("Solana: " + logMsg)
		logToFile(logMsg)
	}

	// 将 Solana 地址转换为伪 EVM 地址（用于存储）
	merchantAddr := solanaAddressToEVMAddress(recipient)
	payerAddr := solanaAddressToEVMAddress(authority)

	// 金额转换
	amountBig := big.NewInt(0)
	if amount > 0 {
		amountBig.SetUint64(amount)
	}

	// 构造 PayoutRecord
	rec := PayoutRecord{
		TxHash:         txHash,
		BlockNumber:    int64(slot),
		DstEid:         EID_SOLANA_DEVNET,
		Payer:          payerAddr,
		Merchant:       merchantAddr,
		SrcToken:       common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"), // Base USDC
		DstToken:       common.HexToAddress(mint),                                         // Solana token mint
		GrossAmount:    amountBig,
		NetAmount:      amountBig,
		Status:         "Delivered",
		Timestamp:      blockTime,
		SolanaMerchant: recipient, // 保存原始 Solana 地址
		SolanaPayer:    authority, // 保存原始 Solana 地址
	}

	// 保存到数据库
	if err := l.store.UpsertPayout(rec); err != nil {
		logMsg := fmt.Sprintf("Failed to save tx %s: %v", txHash[:min(20, len(txHash))], err)
		log.Println("Solana: " + logMsg)
		logToFile(logMsg)
		return fmt.Errorf("failed to upsert payout: %w", err)
	}

	logMsg = fmt.Sprintf("✅ Indexed transfer_out: tx=%s, recipient=%s, amount=%d, slot=%d",
		txHash[:min(20, len(txHash))], recipient[:min(10, len(recipient))], amount, slot)
	log.Println("Solana: " + logMsg)
	logToFile(logMsg)

	return nil
}

// solanaAddressToEVMAddress 将 Solana 地址转换为伪 EVM 地址
// 使用 Solana 地址的哈希值作为 EVM 地址
func solanaAddressToEVMAddress(solAddr string) common.Address {
	if solAddr == "" {
		return common.HexToAddress("0x0000000000000000000000000000000000000000")
	}

	// 解析 Solana 地址
	pubkey, err := solana.PublicKeyFromBase58(solAddr)
	if err != nil {
		return common.HexToAddress("0x0000000000000000000000000000000000000000")
	}

	// 使用公钥的前20字节作为 EVM 地址
	var evmAddr common.Address
	copy(evmAddr[:], pubkey[:20])

	return evmAddr
}

// SolanaTransactionInfo 解析后的 Solana 交易信息
type SolanaTransactionInfo struct {
	Signature   string
	Slot        uint64
	BlockTime   time.Time
	Payer       string
	Merchant    string
	Amount      *big.Int
	Token       string
	Status      string
	LogMessages []string
}

// ParseSolanaLogs 解析 Solana 程序日志
// 这个函数需要根据你的 Solana 程序的实际日志格式来实现
func ParseSolanaLogs(logs []string) (*SolanaTransactionInfo, error) {
	info := &SolanaTransactionInfo{
		LogMessages: logs,
	}

	// TODO: 根据实际的 Solana 程序日志格式解析
	// 示例：
	// - 查找包含 "Transfer" 的日志
	// - 提取金额、发送方、接收方等信息
	// - 解析 base64 编码的数据

	for _, logMsg := range logs {
		// 示例解析逻辑
		// if strings.Contains(logMsg, "Program log: Transfer") {
		//     // 解析转账信息
		// }

		// 这里需要根据你的合约定义具体实现
		_ = logMsg
	}

	return info, nil
}

// DecodeSolanaInstruction 解码 Solana 指令数据
func DecodeSolanaInstruction(data []byte) (map[string]interface{}, error) {
	// 根据你的 Solana 程序的指令格式解码
	// 这需要知道程序的 IDL (Interface Definition Language)

	if len(data) == 0 {
		return nil, fmt.Errorf("empty instruction data")
	}

	result := make(map[string]interface{})

	// TODO: 实现具体的解码逻辑
	// 示例：
	// - 第一个字节通常是指令类型
	// - 后续字节是参数
	result["instruction_type"] = data[0]
	result["data"] = base64.StdEncoding.EncodeToString(data)

	return result, nil
}
