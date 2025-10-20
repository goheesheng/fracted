# 📖 用户指南

完整的使用说明、登录指南、Dashboard使用和最佳实践。

## 📋 目录

1. [快速开始](#快速开始)
2. [登录指南](#登录指南)
3. [Dashboard使用](#dashboard使用)
4. [Solana功能](#solana功能)
5. [白名单管理](#白名单管理)
6. [数据导出](#数据导出)
7. [故障排查](#故障排查)
8. [最佳实践](#最佳实践)
9. [常见问题](#常见问题)

---

## 快速开始

### 5分钟快速上手

#### 步骤1: 启动服务（30秒）

**Windows**:
```powershell
.\scripts\start.ps1
```

**Linux**:
```bash
chmod +x scripts/*.sh
./scripts/start.sh
```

**期望输出**:
```
✅ 编译成功
🚀 启动服务器...
main: Solana listener created
main: starting API at :8080
```

#### 步骤2: 访问Dashboard（30秒）

打开浏览器访问:
```
http://localhost:8080/dashboard/
```

#### 步骤3: 登录（1分钟）

**管理员登录**:
```
地址: 0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6
角色: admin
```

**Solana商家登录**:
```
地址: 6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp
角色: merchant
```

#### 步骤4: 查看交易

- Solana交易带有绿色"Solana Devnet"标签
- EVM交易带有相应链的标签
- 点击任意交易查看详情

### 停止服务

**Windows**:
```powershell
.\scripts\stop.ps1
```

**Linux**:
```bash
./scripts/stop.sh
```

---

## 登录指南

### 支持的地址格式

#### EVM地址（Base, Arbitrum等）
- **格式**: `0x` + 40个十六进制字符
- **示例**: `0x77Ed7f6455FE291728A48785090292e3D10F53Bb`
- **长度**: 42字符（包括0x）
- **用途**: EVM链的管理员和商家

#### Solana地址
- **格式**: 32-44个Base58字符
- **示例**: `6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp`
- **字符集**: 1-9, A-H, J-N, P-Z, a-k, m-z（不包含0、O、I、l）
- **用途**: Solana链的商家

### 登录流程

#### 管理员登录（查看所有数据）

1. 访问: http://localhost:8080/dashboard/
2. 在弹出的登录框中输入管理员地址
3. 点击"Login as Admin"
4. 登录成功后可以查看所有商家的交易

**默认管理员**:
```
0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6
```

#### 商家登录（查看个人数据）

##### 方式1: 在管理员Dashboard登录
1. 访问: http://localhost:8080/dashboard/
2. 输入商家地址（EVM或Solana）
3. 选择角色: merchant
4. 查看个人交易

##### 方式2: 使用商家专用登录页
1. 访问: http://localhost:8080/dashboard/login.html
2. 输入商家地址
3. 点击"Access Dashboard"
4. 自动跳转到商家Dashboard

**可用的商家地址**:

| 类型 | 地址 | 交易数 |
|------|------|-------|
| Solana | `6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp` | 7+ |
| Solana | `A9QYh2sTEN3XFFk95WZr2hsLFMC2781oPwKexPySNJrt` | 1+ |
| Solana | `AWuN8Gk6X3xKR73YBRw2H8WXC6QGbJRveHS5DgEJX3ZS` | 3+ |
| EVM | `0x77Ed7f6455FE291728A48785090292e3D10F53Bb` | - |

---

## Dashboard使用

### 管理员Dashboard

#### 概览卡片
- **Total Inflow (Gross)**: 所有交易的总流入
- **Total Outflow (Net)**: 所有交易的净流出

#### Merchant Total Received
按商家统计的总收入（降序排列）。

#### Payer Total Spent
按付款方统计的总支出（降序排列）。

#### Transactions交易列表
所有交易的详细列表，包含：
- **Identity**: 商家地址（Solana显示Base58，EVM显示0x）
- **Time**: 相对时间（如"2h ago"）
- **Value**: 交易金额（USD）
- **Destination**: 目标链（Solana为绿色标签）
- **Tokens**: 代币类型（USDC/USDT）
- **Activity**: 交易描述

**搜索功能**: 在搜索框中输入地址可筛选交易。

**查看详情**: 点击任意交易行查看完整JSON数据。

### 商家Dashboard

#### 商家信息卡片
- **Merchant Address**: 您的钱包地址
- **Total Transactions**: 总交易数
- **Total Received**: 总收入
- **Last Activity**: 最后活动时间

#### Recent Transactions
最近的5笔交易。

#### Token Summary
按代币类型统计的收入。

#### All Transactions
所有交易的完整列表，支持搜索。

---

## Solana功能

### Solana交易特征

在Dashboard中，Solana交易具有以下特征：

1. **绿色标签**: Destination显示为"Solana Devnet"（绿色徽章）
2. **Base58地址**: Merchant和Payer显示为Solana格式
3. **交易签名**: TxHash是Solana交易签名（Base58）
4. **Slot编号**: BlockNumber显示为Slot编号

### 查看Solana交易详情

点击Solana交易后显示的信息：

```json
{
  "TxHash": "5ogaMvNqF1QY1uba8F8xM2PnwMzFmHoGwXCrCe8xVZHek...",
  "BlockNumber": 415675503,  // Slot编号
  "DstChain": "Solana Devnet",
  "Merchant": "6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp",
  "Payer": "6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp",
  "NetAmountUSD": "1.00",
  "Status": "Delivered"
}
```

### 在Solana Explorer验证

点击交易哈希可以在Solana Explorer中验证：
```
https://explorer.solana.com/tx/<SIGNATURE>?cluster=devnet
```

### Solana商家登录

1. 访问登录页面
2. 输入Solana地址（Base58格式）
3. 系统自动识别并验证
4. 登录后只能看到自己的交易

**注意事项**:
- ✅ 直接粘贴Base58地址
- ✅ 不需要0x前缀
- ✅ 保持原始大小写（系统会自动标准化）
- ❌ 不要手动添加任何前缀或后缀

---

## 白名单管理

### 查看白名单

**管理员白名单**:
```bash
curl -X GET http://localhost:8080/admin/admins \
  -H "Authorization: Bearer $TOKEN"
```

**商家白名单**:
```bash
curl -X GET http://localhost:8080/admin/merchants \
  -H "Authorization: Bearer $TOKEN"
```

### 添加地址

**添加EVM商家**:
```bash
curl -X POST http://localhost:8080/admin/merchants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"address": "0x1234567890123456789012345678901234567890"}'
```

**添加Solana商家**:
```bash
curl -X POST http://localhost:8080/admin/merchants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"address": "NewSolanaBase58Address..."}'
```

### 移除地址

```bash
curl -X DELETE "http://localhost:8080/admin/merchants/AddressToRemove" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 数据导出

### 导出所有交易

```bash
sqlite3 indexer.db -header -csv \
  "SELECT * FROM payouts ORDER BY timestamp DESC;" \
  > transactions_export.csv
```

### 导出Solana交易

```bash
sqlite3 indexer.db -header -csv \
  "SELECT * FROM payouts WHERE dst_eid = 40168 ORDER BY timestamp DESC;" \
  > solana_transactions.csv
```

### 导出特定商家数据

```bash
sqlite3 indexer.db -header -csv \
  "SELECT * FROM payouts WHERE solana_merchant = '6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp';" \
  > merchant_data.csv
```

---

## 故障排查

### 常见问题

#### Q1: 服务无法启动

**症状**: 端口被占用
```
panic: listen tcp :8080: bind: address already in use
```

**解决**:
```bash
# Windows
Get-Process | Where-Object {$_.ProcessName -like "*cross-chain*"} | Stop-Process -Force
netstat -ano | findstr :8080

# Linux
lsof -i :8080
kill -9 <PID>
```

#### Q2: Solana监听器无响应

**症状**: solana_log.txt为空

**检查**:
1. 服务是否正常启动（查看控制台日志）
2. RPC连接是否正常（https://api.devnet.solana.com）
3. 程序地址是否正确

**解决**:
```bash
# 重启服务
.\scripts\stop.ps1  # Windows
.\scripts\start.ps1

./scripts/stop.sh   # Linux
./scripts/start.sh
```

#### Q3: Dashboard登录失败

**症状**: "Address not authorized"

**原因**:
1. 地址不在白名单中
2. 地址格式错误

**解决**:
1. 检查地址是否在`config.go`的白名单中
2. 验证地址格式（EVM: 42字符，Solana: 32-44字符）
3. 使用管理员API添加地址

#### Q4: 看不到交易数据

**症状**: Dashboard显示"No transactions"

**检查**:
1. 是否已登录（右上角应显示用户信息）
2. 清除缓存: `localStorage.clear(); location.reload()`
3. 检查数据库是否有数据
4. 查看浏览器控制台是否有错误

#### Q5: Solana地址显示为0x格式

**原因**: 旧数据或缓存问题

**解决**:
1. 刷新浏览器（F5）
2. 清除缓存并重新登录
3. 确认数据库有solana_merchant字段（重启服务会自动迁移）

#### Q6: 如何监控服务状态

**健康检查**:
```bash
curl http://localhost:8080/health
```

**期望响应**:
```json
{
  "ok": true,
  "db": true,
  "wssStatus": "Connected"
}
```

---

## 最佳实践

### 安全建议

1. **生产环境必须更改JWT_SECRET**
2. **定期审查白名单**
3. **使用HTTPS**（通过Nginx反向代理）
4. **定期备份数据库**
5. **监控日志异常**

### 运维建议

1. **日志轮转**: 定期归档solana_log.txt
2. **数据库维护**: 定期备份indexer.db
3. **监控服务**: 使用systemd或supervisor
4. **性能监控**: 关注RPC连接状态和响应时间

### 开发建议

1. **本地测试**: 使用测试网络（Sepolia, Devnet）
2. **代码审查**: 添加新功能前进行测试
3. **文档更新**: 修改后及时更新文档
4. **版本控制**: 使用Git管理代码变更

---

## 常见问题

### 如何添加新的Solana商家？

**方法1: 环境变量**
```bash
MERCHANT_ADDRESSES=existing...,NewSolanaAddress...
```

**方法2: 管理员API**
```bash
curl -X POST http://localhost:8080/admin/merchants \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"address": "NewSolanaAddress..."}'
```

### Solana商家登录失败？

**检查清单**:
- [ ] 地址格式正确（32-44字符，Base58）
- [ ] 地址在白名单中
- [ ] 没有额外的空格或特殊字符
- [ ] 网络连接正常

**调试方法**:
```javascript
// 在浏览器控制台测试
fetch('/auth/login', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    address: '6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp',
    role: 'merchant'
  })
})
.then(r => r.json())
.then(d => console.log(d));
```

### Dashboard上看不到Solana交易？

**原因**: 可能需要登录或刷新

**解决**:
1. 确保已登录（管理员或相应的商家）
2. 清除缓存: `localStorage.clear(); location.reload()`
3. 检查是否有交易: 查看solana_log.txt

### 如何查看Solana监听器状态？

**查看日志**:
```bash
# Windows
Get-Content solana_log.txt -Tail 20

# Linux
tail -f solana_log.txt
```

**检查数据库**:
```sql
SELECT COUNT(*) FROM payouts WHERE dst_eid = 40168;
```

---

## 支持的链

| 链名称 | 网络 | EID | 监听类型 | 地址格式 |
|--------|------|-----|---------|---------|
| Base Sepolia | 测试网 | 40245 | WSS事件 | 0x |
| Arbitrum Sepolia | 测试网 | 40231 | 状态查询 | 0x |
| Solana Devnet | 测试网 | 40168 | WS交易 | Base58 |

---

## 快速命令参考

### 启动/停止
```bash
# 启动
.\scripts\start.ps1        # Windows
./scripts/start.sh         # Linux

# 停止
.\scripts\stop.ps1         # Windows
./scripts/stop.sh          # Linux
```

### 查看日志
```bash
# Solana日志
Get-Content solana_log.txt -Wait -Tail 20  # Windows
tail -f solana_log.txt                      # Linux
```

### 数据库查询
```bash
# 统计各链交易数
sqlite3 indexer.db "SELECT dst_eid, COUNT(*) FROM payouts GROUP BY dst_eid;"

# 查看Solana交易
sqlite3 indexer.db "SELECT * FROM payouts WHERE dst_eid=40168 LIMIT 5;"
```

### 测试登录
```bash
# EVM管理员
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"address":"0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6","role":"admin"}'

# Solana商家
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"address":"6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp","role":"merchant"}'
```

---

## 性能指标

### 监听延迟
- **EVM链（Base）**: 1-3秒
- **Solana**: 2-5秒
- **Dashboard刷新**: 5秒（自动）

### 资源占用
- **内存**: ~50-100 MB
- **CPU**: <5%（空闲时）
- **网络**: ~1-5 KB/s
- **磁盘**: 数据库随交易增长

### 处理能力
- **回填速度**: ~10笔/秒
- **实时处理**: 无延迟
- **并发**: 支持多客户端同时访问

---

## 联系支持

- **技术问题**: 查看[技术文档](TECHNICAL.md)
- **功能建议**: 提交Issue
- **安全问题**: 私密报告

---

**最后更新**: 2025-10-20  
**版本**: v2.0 - 多链完整版
