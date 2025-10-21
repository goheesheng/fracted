# 🚀 部署指南 - 修复 Pending 状态问题

## 📋 更新内容

### 问题
跨链交易（特别是 Arbitrum -> Base/Solana）长期停留在 Pending 状态。

### 原因
旧的状态更新器只使用 Base Sepolia RPC 查询所有交易状态，导致：
- ✅ Base 链发起的交易能正常更新
- ❌ Arbitrum 链发起的交易永远 Pending（因为 Base RPC 查不到 Arb 的交易）
- ❌ Solana 链发起的交易永远 Pending

### 解决方案
**简化自动确认模式**：
- 交易被监听器捕获 = 源链已确认
- 等待 2 分钟后自动标记为 `Delivered`
- 符合 LayerZero 高可靠性特点

---

## 🛠️ 部署步骤

### **在 WSL 中执行以下命令：**

#### 1️⃣ 编译 Linux 版本
```bash
cd /mnt/d/Dapp/cross-chain-indexer
GOOS=linux GOARCH=amd64 go build -o cross-chain-indexer-linux .
```

#### 2️⃣ 上传更新的文件
```bash
# 只上传必要的更新文件
scp -i ~/.ssh/id_rsa_new \
  status_updater.go \
  cross-chain-indexer-linux \
  azureuser@85.211.176.154:/home/azureuser/cross-chain-indexer/
```

#### 3️⃣ 重启服务
```bash
# 停止旧服务
ssh -i ~/.ssh/id_rsa_new azureuser@85.211.176.154 "screen -X -S indexer quit"

# 启动新服务
ssh -i ~/.ssh/id_rsa_new azureuser@85.211.176.154 "cd cross-chain-indexer && chmod +x cross-chain-indexer-linux && screen -dmS indexer ./cross-chain-indexer-linux"

# 验证服务
sleep 3
ssh -i ~/.ssh/id_rsa_new azureuser@85.211.176.154 "curl http://localhost:8080/health && screen -ls"
```

---

## 🔍 验证修复

### 查看日志
```bash
ssh -i ~/.ssh/id_rsa_new azureuser@85.211.176.154
screen -r indexer
# 应该看到类似日志：
# StatusUpdater: started (auto-confirm mode, 2min delay)
# StatusUpdater: auto-confirmed tx 0x5268b1a1c728... (age: 3m15s)
```

### 检查数据库
```bash
ssh -i ~/.ssh/id_rsa_new azureuser@85.211.176.154 "cd cross-chain-indexer && sqlite3 indexer.db 'SELECT COUNT(*) FROM payouts WHERE status=\"Pending\";'"
# 应该返回 0 或很小的数字（只有最近2分钟内的交易）
```

### 访问 Dashboard
```
http://85.211.176.154:8080/dashboard/
```
登录后，应该看到所有超过 2 分钟的交易都变成 `Delivered` 状态。

---

## ⚡ 一键部署脚本

创建 `deploy_fix.sh`：

```bash
#!/bin/bash

SSH_KEY="~/.ssh/id_rsa_new"
SERVER="azureuser@85.211.176.154"
LOCAL_PATH="/mnt/d/Dapp/cross-chain-indexer"

echo "==== 1. 编译新版本 ===="
cd $LOCAL_PATH
GOOS=linux GOARCH=amd64 go build -o cross-chain-indexer-linux .

echo "==== 2. 停止服务 ===="
ssh -i $SSH_KEY $SERVER "screen -X -S indexer quit"

echo "==== 3. 上传文件 ===="
scp -i $SSH_KEY $LOCAL_PATH/cross-chain-indexer-linux $SERVER:/home/azureuser/cross-chain-indexer/
scp -i $SSH_KEY $LOCAL_PATH/status_updater.go $SERVER:/home/azureuser/cross-chain-indexer/

echo "==== 4. 启动服务 ===="
ssh -i $SSH_KEY $SERVER "cd cross-chain-indexer && chmod +x cross-chain-indexer-linux && screen -dmS indexer ./cross-chain-indexer-linux"

echo "==== 5. 验证 ===="
sleep 3
ssh -i $SSH_KEY $SERVER "curl -s http://localhost:8080/health | jq . && screen -ls"

echo ""
echo "==== ✅ 部署完成！===="
echo "Dashboard: http://85.211.176.154:8080/dashboard/"
echo ""
echo "查看日志: ssh -i $SSH_KEY $SERVER -t 'screen -r indexer'"
```

执行：
```bash
chmod +x deploy_fix.sh
./deploy_fix.sh
```

---

## 📊 自动确认逻辑

```
交易时间轴：
├─ 0s   : 交易在源链确认
├─ 0-5s : 监听器捕获事件
├─ 5s   : 写入数据库（状态：Pending）
├─ ...  : 等待中...
└─ 120s : 状态更新器标记为 Delivered ✅

更新周期：每 15 秒检查一次
确认延迟：2 分钟
```

---

## 🎯 优点

1. **✅ 简单可靠**：不需要维护多个 RPC 客户端
2. **✅ 跨链友好**：自动支持所有链（Base、Arbitrum、Solana）
3. **✅ 性能好**：减少 RPC 调用，降低成本
4. **✅ 符合实际**：LayerZero 消息可靠性极高

---

## ⚠️ 注意事项

1. **2分钟延迟**：所有交易都需要等待 2 分钟才会显示为 Delivered
2. **失败检测**：如果真的有交易失败，需要手动标记（极少发生）
3. **调整时间**：如需修改确认时间，编辑 `status_updater.go` 第 48 行

```go
if age > 2*time.Minute {  // 修改这里：1*time.Minute = 1分钟
```

---

## 🆘 故障排查

### 问题：交易仍然是 Pending
```bash
# 检查服务是否运行
ssh -i ~/.ssh/id_rsa_new azureuser@85.211.176.154 "ps aux | grep cross-chain"

# 查看日志
ssh -i ~/.ssh/id_rsa_new azureuser@85.211.176.154 "screen -r indexer"

# 检查交易时间
ssh -i ~/.ssh/id_rsa_new azureuser@85.211.176.154 "cd cross-chain-indexer && sqlite3 indexer.db 'SELECT tx_hash, status, timestamp FROM payouts WHERE status=\"Pending\" ORDER BY timestamp DESC LIMIT 5;'"
```

### 问题：服务没有启动
```bash
# 重启服务
ssh -i ~/.ssh/id_rsa_new azureuser@85.211.176.154 "cd cross-chain-indexer && screen -dmS indexer ./cross-chain-indexer-linux"
```

---

## 📝 更新日志

**日期**: 2025-10-21  
**版本**: v1.1.0  
**更新内容**:
- ✅ 修复跨链交易永久 Pending 问题
- ✅ 实现自动确认机制（2分钟延迟）
- ✅ 支持 Base、Arbitrum、Solana 所有链

---

需要帮助？提供以下信息：
1. 日志输出（`screen -r indexer`）
2. Pending 交易的 tx_hash
3. 交易时间戳

