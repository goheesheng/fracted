#!/bin/bash
# Cross-Chain Indexer - Linux 停止脚本

echo "====================================="
echo " Cross-Chain Indexer - 停止服务"
echo "====================================="
echo ""

# 查找并停止进程
if pgrep -x "cross-chain-indexer" > /dev/null; then
    echo "正在停止服务..."
    pkill -9 cross-chain-indexer
    sleep 1
    
    # 验证是否已停止
    if pgrep -x "cross-chain-indexer" > /dev/null; then
        echo "警告: 进程仍在运行"
        exit 1
    else
        echo "✓ 服务已停止"
    fi
else
    echo "没有找到运行中的进程"
fi

echo ""
echo "====================================="
echo " 停止完成"
echo "====================================="
echo ""
echo "重新启动服务:"
echo "  ./scripts/start.sh"
echo ""

