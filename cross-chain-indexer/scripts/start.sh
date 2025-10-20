#!/bin/bash
# Cross-Chain Indexer - Linux 启动脚本
# 功能: 编译、启动服务、健康检查

echo "====================================="
echo " Cross-Chain Indexer - 启动服务"
echo "====================================="
echo ""

# 检查是否在项目根目录
if [ ! -f "main.go" ]; then
    echo "错误: 请在项目根目录执行此脚本"
    exit 1
fi

# 1. 停止现有进程
echo "1. 检查现有进程..."
if pgrep -x "cross-chain-indexer" > /dev/null; then
    echo "   发现运行中的进程，正在停止..."
    pkill -9 cross-chain-indexer
    sleep 1
    echo "   已停止"
else
    echo "   无运行中的进程"
fi

# 2. 编译项目
echo ""
echo "2. 编译项目..."
if go build -o cross-chain-indexer .; then
    echo "   ✓ 编译成功"
else
    echo "   ✗ 编译失败"
    exit 1
fi

# 3. 启动服务
echo ""
echo "3. 启动服务..."
nohup ./cross-chain-indexer > /dev/null 2>&1 &
INDEXER_PID=$!
echo "   进程 ID: $INDEXER_PID"
sleep 3

# 4. 健康检查
echo ""
echo "4. 健康检查..."
max_retries=5
retry_count=0

while [ $retry_count -lt $max_retries ]; do
    if curl -s -f http://localhost:8080/health > /dev/null 2>&1; then
        echo "   ✓ 服务运行正常"
        healthy=true
        break
    else
        retry_count=$((retry_count + 1))
        if [ $retry_count -lt $max_retries ]; then
            echo "   等待服务启动... ($retry_count/$max_retries)"
            sleep 2
        fi
    fi
done

if [ "$healthy" != "true" ]; then
    echo "   ✗ 服务启动超时，请检查日志"
    exit 1
fi

# 5. 显示信息
echo ""
echo "====================================="
echo " 服务启动成功！"
echo "====================================="
echo ""
echo "Dashboard:"
echo "  http://localhost:8080/dashboard/"
echo ""
echo "健康检查:"
echo "  http://localhost:8080/health"
echo ""
echo "进程 ID: $INDEXER_PID"
echo ""
echo "停止服务:"
echo "  ./scripts/stop.sh"
echo ""
echo "查看 Solana 日志:"
echo "  tail -f solana_log.txt"
echo ""

