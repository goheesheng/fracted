# 🔧 技术文档

完整的API文档、架构设计、部署指南和开发指南。

## 📋 目录

1. [API文档](#api文档)
2. [架构设计](#架构设计)
3. [Solana集成](#solana集成)
4. [数据库设计](#数据库设计)
5. [部署指南](#部署指南)
6. [开发指南](#开发指南)
7. [性能优化](#性能优化)
8. [安全配置](#安全配置)
9. [故障排查](#故障排查)

---

## API文档

### 认证端点

#### POST /auth/login
用户登录，支持EVM和Solana地址。

**请求**:
```json
{
  "address": "0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6",  // 或Solana Base58地址
  "role": "admin"  // 或"merchant"
}
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "address": "0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6",
  "role": "admin"
}
```

#### GET /auth/me
获取当前用户信息。

**Header**: `Authorization: Bearer <token>`

**响应**:
```json
{
  "address": "0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6",
  "role": "admin"
}
```

### 商家端点

#### GET /merchant/payouts
查询商家的交易记录（需要认证）。

**Header**: `Authorization: Bearer <token>`

**Query参数**:
- `limit`: 每页数量（默认50，最大500）
- `offset`: 偏移量（默认0）

**响应**:
```json
[
  {
    "TxHash": "5ogaMvNqF1QY1uba8F8xM2PnwMzFmHoGwXCrCe8xVZHekpqpKbCaiMc9BXQ9GkEu4s93SNurzZQfw6Wi7z652s6L",
    "DstChain": "Solana Devnet",
    "Merchant": "6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp",
    "Payer": "6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp",
    "NetAmountUSD": "1.00",
    "Status": "Delivered",
    "Timestamp": "2025-10-19T14:46:56Z"
  }
]
```

### 管理员端点

#### GET /admin/payouts
查询所有交易记录（需要管理员权限）。

**URL参数认证**: `?token=<jwt_token>`

#### POST /admin/backfill
触发历史数据回填。

**请求**:
```json
{
  "from_block": 0,
  "to_block": 0  // 0表示使用默认值
}
```

#### GET /admin/merchants
列出所有商家地址。

#### POST /admin/merchants
添加商家地址（支持EVM和Solana）。

**请求**:
```json
{
  "address": "6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp"
}
```

#### DELETE /admin/merchants/{address}
移除商家地址。

#### GET /admin/admins
列出所有管理员地址。

#### POST /admin/admins
添加管理员地址。

#### DELETE /admin/admins/{address}
移除管理员地址。

---

## 架构设计

### 组件架构

```
┌─────────────────────────────────────────────────┐
│         HTTP API Server (api.go)                │
│  /auth  /merchant  /admin  /dashboard           │
└───────────────┬─────────────────────────────────┘
                │
    ┌───────────┴───────────┐
    │                       │
┌───▼────┐            ┌────▼─────┐
│  Store │            │ Listeners │
│(SQLite)│            │           │
└────────┘            ├───────────┤
                      │ Base WSS  │
                      │ Arb Query │
                      │ Solana WS │
                      └───────────┘
```

### 数据流

```
区块链交易
    ↓
监听器捕获（实时/查询）
    ↓
解析数据（根据链类型）
    ↓
提取关键信息
    ↓
地址处理（保存双格式）
    ↓
保存到数据库
    ↓
API查询
    ↓
智能地址转换
    ↓
Dashboard显示
```

### 地址处理策略

```
存储层（数据库）:
  - merchant/payer: EVM格式（用于索引和查询）
  - solana_merchant/solana_payer: Base58格式（用于显示）

API层:
  - 根据dst_eid判断链类型
  - Solana链(40168/30168) → 返回Base58地址
  - EVM链 → 返回0x地址

前端层:
  - 直接显示API返回的地址
  - 自动识别格式无需转换
```

---

## Solana集成

### 合约信息

- **程序ID**: `GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1`
- **网络**: Solana Devnet
- **指令**: transfer_out

### transfer_out指令结构

#### 账户顺序（7个账户）
```rust
0: config (PDA)
1: authority (Signer) - 调用方
2: vault_authority (PDA)
3: vault_token_account (mut)
4: recipient_token_account (mut) - 接收方
5: mint - 代币类型
6: token_program
```

#### 指令数据
```
[0-7]   bytes: Discriminator (Anchor自动生成)
[8-15]  bytes: Amount (u64, little-endian)
```

### 监听器工作原理

1. **启动时回填**: 扫描最近100笔交易
2. **实时监听**: WebSocket订阅程序交易
3. **解析指令**: 识别transfer_out并提取数据
4. **提取信息**:
   - Authority → Payer
   - Recipient token account owner → Merchant
   - Amount → 转账金额
   - Mint → 代币类型
5. **保存数据**: 存入数据库（dst_eid = 40168）
6. **地址处理**: 同时保存Solana Base58和EVM转换格式

### 地址转换

Solana公钥（32字节）→ EVM地址（20字节）:
- 使用Solana公钥的前20字节作为EVM地址
- 保证唯一性和一致性

### 日志文件

所有Solana交易处理日志写入 `solana_log.txt`：

```
[2025-10-20 01:15:13] Processing tx 5ogaMvNq... (slot: 415675503)
[2025-10-20 01:15:13] Found transfer_out instruction #0: amount=1000000
[2025-10-20 01:15:13] ✅ Indexed transfer_out: tx=5ogaMvNq...
```

### Solana Explorer

查看程序交易：
```
https://explorer.solana.com/address/GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1?cluster=devnet
```

---

## 数据库设计

### payouts表

| 字段 | 类型 | 说明 |
|------|------|------|
| tx_hash | TEXT | 交易哈希（主键）|
| block_number | INTEGER | 区块号/Slot |
| timestamp | DATETIME | 交易时间 |
| dst_eid | INTEGER | 目标链EID（40168=Solana Devnet）|
| payer | TEXT | 付款方（EVM格式）|
| merchant | TEXT | 收款方（EVM格式）|
| src_token | TEXT | 源代币地址 |
| dst_token | TEXT | 目标代币地址 |
| gross_amount | TEXT | 总金额 |
| net_amount | TEXT | 净金额 |
| status | TEXT | 状态（Pending/Delivered/Failed）|
| solana_merchant | TEXT | Solana原始地址（Base58）|
| solana_payer | TEXT | Solana原始地址（Base58）|
| created_at | DATETIME | 创建时间 |

**索引**:
- `idx_payouts_merchant` - 商家地址索引
- `idx_payouts_dst_eid` - 目标链索引
- `idx_payouts_timestamp` - 时间戳索引

### events表

存储原始事件日志（用于调试和重放）。

### processed_blocks表

记录每条链已处理到的区块高度。

---

## 部署指南

### 开发环境

#### 手动启动
```bash
# 编译
go build -o cross-chain-indexer .

# 运行
./cross-chain-indexer
```

#### 使用脚本（推荐）

**Windows**:
```powershell
.\scripts\start.ps1
```

**Linux**:
```bash
chmod +x scripts/*.sh
./scripts/start.sh
```

### 生产环境

#### Docker部署

**构建镜像**:
```bash
docker build -t cross-chain-indexer .
```

**运行容器**:
```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/indexer.db:/app/indexer.db \
  -e JWT_SECRET=your-production-secret \
  -e ADMIN_ADDRESSES=0x... \
  -e MERCHANT_ADDRESSES=0x...,SolanaAddr... \
  --name indexer \
  cross-chain-indexer
```

**Docker Compose**:
```bash
# 启动完整环境
docker-compose up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f cross-chain-indexer

# 停止服务
docker-compose down
```

#### 系统服务部署（Linux）

```bash
# 复制服务文件
sudo cp scripts/cross-chain-indexer.service /etc/systemd/system/

# 启用并启动
sudo systemctl enable cross-chain-indexer
sudo systemctl start cross-chain-indexer

# 查看状态
sudo systemctl status cross-chain-indexer

# 查看日志
sudo journalctl -u cross-chain-indexer -f
```

### 环境变量

| 变量名 | 说明 | 默认值 | 示例 |
|--------|------|--------|------|
| `JWT_SECRET` | JWT签名密钥 | `dev-local-secret-change-me` | `your-secret-key-32-chars-min` |
| `ADMIN_ADDRESSES` | 管理员地址（逗号分隔） | 见config.go | `0xAddr1,0xAddr2` |
| `MERCHANT_ADDRESSES` | 商家地址（逗号分隔，支持EVM和Solana） | 见config.go | `0xEVM,SolanaBase58` |

---

## 开发指南

### 项目结构

```
cross-chain-indexer/
├── main.go              # 主入口，初始化各组件
├── api.go               # HTTP API路由和处理器
├── config.go            # 配置管理和白名单
├── store.go             # 数据库操作
├── processor.go         # EVM链事件处理
├── solana_listener.go   # Solana链监听器
├── status_updater.go    # 状态更新器
└── *_test.go            # 测试文件
```

### 添加新链支持

#### 1. 定义EID
在 `main.go` 中添加：
```go
const (
    EID_NEW_CHAIN = 12345
)
```

#### 2. 更新链名称映射
在 `api.go` 中：
```go
func getChainName(eid int64) string {
    case 12345:
        return "New Chain Name"
}
```

#### 3. 实现监听器
参考 `solana_listener.go` 实现新的监听器。

### 添加新的Solana商家

#### 方法1: 环境变量
```bash
MERCHANT_ADDRESSES=existing...,NewSolanaAddress...
```

#### 方法2: 代码配置
在 `config.go` 中：
```go
config.MerchantAddresses["new_solana_address_lowercase"] = true
```

#### 方法3: 运行时添加（管理员API）
```bash
curl -X POST http://localhost:8080/admin/merchants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"address": "NewSolanaAddress..."}'
```

### 智能地址系统

系统自动处理两种地址格式：

```go
// 验证
if isValidEVMAddress(addr) {
    // EVM地址处理
} else if isValidSolanaAddress(addr) {
    // Solana地址处理
}

// 显示（API响应）
if isSolanaChain(eid) && solanaMerchant != "" {
    response.Merchant = solanaMerchant  // Base58
} else {
    response.Merchant = merchant.Hex()  // 0x
}
```

---

## 性能优化

### RPC配置

**使用高性能RPC**:
- Solana: Helius, QuickNode, Alchemy
- Base: Alchemy, Infura, QuickNode

**在main.go中配置**:
```go
const (
    baseSepoliaWSS = "wss://your-premium-rpc"
    solanaDevnetRPC = "https://your-premium-rpc"
)
```

### 数据库优化

**添加索引**:
```sql
CREATE INDEX IF NOT EXISTS idx_payouts_timestamp ON payouts(timestamp);
CREATE INDEX IF NOT EXISTS idx_payouts_status ON payouts(status);
```

**定期清理**:
```sql
-- 删除30天前的已完成交易
DELETE FROM payouts 
WHERE status = 'Delivered' 
AND timestamp < datetime('now', '-30 days');
```

### 监听器优化

**调整回填数量**:
```go
// main.go
solanaListener.BackfillHistoricalTransactions(ctx, 100)  // 默认100
```

**调整轮询间隔**:
```go
// main.go
go statusUpdater(store, httpsClient, 15*time.Second)  // 默认15秒
```

---

## 安全配置

### JWT密钥

**生产环境必须更改**:
```bash
# 生成强密钥（32+字符）
openssl rand -base64 32

# 设置环境变量
export JWT_SECRET="生成的密钥"
```

### HTTPS配置

**使用Nginx反向代理**:
```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 白名单管理

**环境变量**:
```bash
# 支持混合地址（EVM和Solana）
MERCHANT_ADDRESSES=0x77Ed...,6H7AYK...,AWuN8G...
```

**运行时管理**:
```bash
# 添加商家
curl -X POST http://localhost:8080/admin/merchants \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"address": "NewAddress"}'

# 移除商家
curl -X DELETE http://localhost:8080/admin/merchants/OldAddress \
  -H "Authorization: Bearer $TOKEN"
```

---

## 故障排查

### Solana监听器问题

#### 症状1: solana_log.txt为空
**原因**: 没有transfer_out交易或监听器未启动

**解决**:
```bash
# 检查服务日志
tail -f solana_log.txt

# 查看最近的链上交易
# 访问Solana Explorer
```

#### 症状2: 交易未被索引
**检查**:
- 交易是否成功（失败交易会被跳过）
- 是否是transfer_out指令
- 账户数量是否为7个

**调试**:
查看日志中的解析信息，确认各字段提取是否成功。

### Dashboard问题

#### 症状: 登录后看不到数据
**原因**: 缓存或token问题

**解决**:
```javascript
// 浏览器控制台
localStorage.clear();
location.reload();
```

#### 症状: Solana地址显示为0x
**原因**: 数据库中缺少solana_merchant字段

**解决**: 数据库会自动迁移，重启服务即可。

---

## 监控和维护

### 日志文件

- **solana_log.txt**: Solana交易处理日志
- 控制台输出: 所有链的事件日志

### 定期维护

**数据库备份**:
```bash
# 每日备份
cp indexer.db indexer.db.backup.$(date +%Y%m%d)
```

**日志轮转**:
```bash
# 压缩旧日志
gzip solana_log.txt
mv solana_log.txt.gz logs/solana_log_$(date +%Y%m%d).txt.gz
touch solana_log.txt
```

---

## 开发工具

### 编译
```bash
# 开发编译
go build -o cross-chain-indexer .

# 生产编译（优化）
go build -ldflags="-s -w" -o cross-chain-indexer .

# 跨平台编译
GOOS=linux GOARCH=amd64 go build -o cross-chain-indexer-linux .
GOOS=windows GOARCH=amd64 go build -o cross-chain-indexer.exe .
```

### 测试
```bash
# 运行所有测试
go test -v ./...

# 运行特定测试
go test -v -run TestAdminSecurity

# 测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 调试

**查看Solana交易**:
```bash
# 实时监控
tail -f solana_log.txt

# 查看特定交易
https://explorer.solana.com/tx/<SIGNATURE>?cluster=devnet
```

**查询数据库**:
```bash
# 统计
sqlite3 indexer.db "SELECT dst_eid, COUNT(*) FROM payouts GROUP BY dst_eid;"

# 查看最新Solana交易
sqlite3 indexer.db "SELECT * FROM payouts WHERE dst_eid=40168 ORDER BY timestamp DESC LIMIT 5;"
```

---

## API速查表

### 认证
```bash
# 登录（EVM）
POST /auth/login {"address":"0x...","role":"admin"}

# 登录（Solana）
POST /auth/login {"address":"6H7AYK...","role":"merchant"}
```

### 查询
```bash
# 商家交易
GET /merchant/payouts?limit=50
Header: Authorization: Bearer <token>

# 所有交易（管理员）
GET /admin/payouts?token=<token>&limit=100
```

### 管理
```bash
# 添加商家（支持Solana）
POST /admin/merchants {"address":"6H7AYK..."}

# 列出商家
GET /admin/merchants
```

---

**更新时间**: 2025-10-20  
**版本**: v2.0 - 多链完整版
