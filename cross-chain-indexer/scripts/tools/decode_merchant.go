package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
)

func main() {
	// 从交易 Event Log Topics[3] 提取的 merchant 值（32字节）
	merchantHex := "6752055c20b3e9d8746656ddf73855507f87ab6d87523e4c76a7fa36096a99eb"
	
	fmt.Println("=== Merchant 地址解析 ===\n")
	fmt.Printf("原始 32 字节 (hex): 0x%s\n\n", merchantHex)
	
	// 方法 1: 当前代码的处理方式（错误）- 只取后 20 字节作为 EVM 地址
	merchantBytes := common.HexToHash("0x" + merchantHex)
	evmAddr := common.BytesToAddress(merchantBytes.Bytes())
	fmt.Printf("❌ 错误处理 (EVM 格式，只取后20字节):\n")
	fmt.Printf("   %s\n\n", evmAddr.Hex())
	
	// 方法 2: 正确的处理方式 - 转换为 Solana Base58 地址
	merchantBytesRaw, err := hex.DecodeString(merchantHex)
	if err != nil {
		log.Fatal(err)
	}
	
	if len(merchantBytesRaw) == 32 {
		solPubkey := solana.PublicKeyFromBytes(merchantBytesRaw)
		fmt.Printf("✅ 正确处理 (Solana Base58 格式):\n")
		fmt.Printf("   %s\n\n", solPubkey.String())
		
		// 验证：将 Solana 地址转回字节，应该匹配原始数据
		verifyBytes := solPubkey.Bytes()
		verifyHex := hex.EncodeToString(verifyBytes)
		
		fmt.Printf("验证:\n")
		fmt.Printf("   原始: 0x%s\n", merchantHex)
		fmt.Printf("   转换: 0x%s\n", verifyHex)
		if merchantHex == verifyHex {
			fmt.Printf("   ✓ 匹配成功\n")
		} else {
			fmt.Printf("   ✗ 不匹配\n")
		}
	} else {
		fmt.Printf("错误：merchant 字节长度不是 32 (%d)\n", len(merchantBytesRaw))
	}
}

