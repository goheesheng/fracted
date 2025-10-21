package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	const arbSepoliaHTTPS = "https://sepolia-rollup.arbitrum.io/rpc"
	const arbContractAddr = "0x1a9C0a66Cb68D92c598B0D2f10de3C755Eb6D438"

	// 正确的topic
	const correctTopic = "0xd892a21f8b815c577e9ce52aa66d230fa1b28664b1286de9e4b85acfac750c31"

	client, err := ethclient.Dial(arbSepoliaHTTPS)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 查询包含新交易的区块
	targetBlock := uint64(206870909)

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(targetBlock) - 10),
		ToBlock:   big.NewInt(int64(targetBlock) + 10),
		Addresses: []common.Address{common.HexToAddress(arbContractAddr)},
		Topics:    [][]common.Hash{{common.HexToHash(correctTopic)}},
	}

	fmt.Println("查询Arbitrum事件...")
	fmt.Printf("区块范围: %d - %d\n", targetBlock-10, targetBlock+10)
	fmt.Printf("合约地址: %s\n", arbContractAddr)
	fmt.Printf("Event Topic: %s\n", correctTopic)
	fmt.Println()

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("找到 %d 个事件\n\n", len(logs))

	for i, vLog := range logs {
		fmt.Printf("事件 #%d:\n", i+1)
		fmt.Printf("  TxHash: %s\n", vLog.TxHash.Hex())
		fmt.Printf("  Block: %d\n", vLog.BlockNumber)
		fmt.Printf("  Topics: %d\n", len(vLog.Topics))
		if len(vLog.Topics) >= 4 {
			dstEid := uint32(vLog.Topics[1].Big().Uint64())
			payer := common.BytesToAddress(vLog.Topics[2].Bytes())
			merchant := common.BytesToAddress(vLog.Topics[3].Bytes())
			fmt.Printf("  DstEid: %d\n", dstEid)
			fmt.Printf("  Payer: %s\n", payer.Hex())
			fmt.Printf("  Merchant: %s\n", merchant.Hex())
		}
		fmt.Println()
	}
}
