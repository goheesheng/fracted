# Fracted Payment API

## 概述

Fracted Payment API 是一个支持多链支付的系统，包括 Ethereum、Arbitrum、Base 和 Solana 网络。系统使用 Snowflake 算法生成唯一的支付 ID，并将支付信息存储在 SQLite 数据库中。

## 功能特性

- ✅ 多链支付支持 (Ethereum, Arbitrum, Base, Solana)
- ✅ Snowflake ID 生成器
- ✅ SQLite 数据库存储
- ✅ RESTful API 接口
- ✅ 移动端适配
- ✅ 支付状态跟踪

## 安装和运行

### 1. 安装依赖

```bash
npm install
```

### 2. 启动服务器

```bash
npm start
```

服务器将在 `http://localhost:8080` 启动。

## API 文档

### 1. 生成支付链接

**EVM 网络示例:**
```
GET /generate-link?merchant=0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6&dstEid=40245&dstToken=0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d&amount=123000000
```

**Solana 网络示例:**
```
GET /generate-link?merchant=7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU&dstEid=40168&dstToken=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v&amount=1000000
```

**响应:**
```json
{
  "success": true,
  "paymentId": "1734567890123456789",
  "paymentLink": "https://demo.fracted.xyz/payment/1734567890123456789",
  "parameters": {
    "merchant": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
    "dstEid": 40245,
    "dstToken": "0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d",
    "amount": 123000000
  },
  "message": "Payment link generated successfully"
}
```

**地址格式说明:**
- **EVM 网络** (Base, Arbitrum): 使用 `0x` 前缀的以太坊地址格式（40 个十六进制字符）
- **Solana 网络**: 使用 base58 编码的地址格式（32-44 个字符）

### 2. 获取支付信息

**请求:**
```
GET /api/payment/{paymentId}
```

**响应:**
```json
{
  "success": true,
  "payment": {
    "id": "1234567890123456789",
    "merchant": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
    "dstEid": 40245,
    "dstToken": "0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d",
    "amount": "123000000",
    "status": "pending",
    "createdAt": "2024-01-01 12:00:00",
    "updatedAt": "2024-01-01 12:00:00"
  }
}
```

### 3. 更新支付状态

**请求:**
```
POST /api/payment/{paymentId}/status
Content-Type: application/json

{
  "status": "processing"
}
```

**响应:**
```json
{
  "success": true,
  "message": "Payment status updated successfully"
}
```

### 4. 获取所有支付记录

**请求:**
```
GET /api/payments
```

**响应:**
```json
{
  "success": true,
  "payments": [
    {
      "id": "1234567890123456789",
      "merchant": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
      "dstEid": 40245,
      "dstToken": "0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d",
      "amount": "123000000",
      "status": "pending",
      "createdAt": "2024-01-01 12:00:00",
      "updatedAt": "2024-01-01 12:00:00"
    }
  ]
}
```

## 数据库结构

### payments 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT | 主键，Snowflake ID |
| merchant_address | TEXT | 商户地址（支持 EVM 和 Solana 格式） |
| dst_eid | INTEGER | 目标链 ID (40245=Base, 40231=Arbitrum, 40168=Solana) |
| dst_token | TEXT | 目标代币地址（支持 EVM 和 Solana 格式） |
| amount | TEXT | 支付金额（最小单位） |
| status | TEXT | 支付状态 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

## 支付状态

- `pending`: 待处理
- `processing`: 处理中
- `completed`: 已完成
- `failed`: 失败
- `cancelled`: 已取消

## 支持的网络

### EVM 网络
- **Arbitrum Sepolia** (EID: 40231)
  - RPC: https://sepolia-rollup.arbitrum.io/rpc
  - 浏览器: https://sepolia.arbiscan.io/
- **Base Sepolia** (EID: 40245)
  - RPC: https://sepolia.base.org
  - 浏览器: https://sepolia.basescan.org/

### Solana 网络
- **Solana Devnet** (EID: 40168)
  - RPC: https://api.devnet.solana.com
  - 浏览器: https://explorer.solana.com/?cluster=devnet
  - **注意:** Solana 地址格式为 base58 编码（例如：`7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU`）

### 支持的钱包
- **MetaMask** (用于 EVM 网络)
- **Phantom** (用于 Solana 网络)

## 故障排除

### 1. Payment ID 负数问题

**问题:** Payment ID 出现负数，如 `-268431360`

**原因:** Snowflake 算法中的位运算导致整数溢出

**解决方案:** 
- 使用更近期的 epoch 时间
- 改用简单的数学运算而不是位运算
- 确保时间戳差异为正数

**修复后的 ID 格式:** 正数，如 `1234567890123456789`

### 2. 数据库连接问题

**问题:** `Cannot find package 'sqlite3'`

**解决方案:**
```bash
npm install sqlite3
```

如果安装失败，可以尝试：
```bash
npm install better-sqlite3
```

### 3. 移动端适配

系统自动检测移动设备并调整界面：
- 桌面端：左右分栏显示
- 移动端：先显示订单确认，点击确认后显示支付界面

## 测试

### 运行 Snowflake 测试
```bash
node test-snowflake.js
```

### 运行支付系统测试
```bash
node test-payment-id.js
```

## 部署

### 环境变量
创建 `.env` 文件：
```
PORT=8080

# EVM 网络配置
OAPP_arbitrum_sepolia=0x...
OAPP_base_sepolia=0x...
TOKEN_arbitrum_sepolia_USDC=0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d
TOKEN_base_sepolia_USDT=0x036CbD53842c5426634e7929541eC2318f3dCF7e

# Solana 网络配置
OAPP_solana_devnet=YourSolanaProgramAddress
TOKEN_solana_devnet_USDC=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
TOKEN_solana_devnet_USDT=Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB

# EID 映射
EID_TO_CHAIN_40245=Base Sepolia
EID_TO_CHAIN_40231=Arbitrum Sepolia
EID_TO_CHAIN_40168=Solana Devnet

# 代币符号映射
TOKEN_SYMBOL_0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d=USDC
TOKEN_SYMBOL_EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v=USDC
```

### 生产环境
```bash
npm start
```

## 快速生成支付链接

使用 `quick-link.js` 脚本快速生成支付链接：

```bash
node quick-link.js
```

修改配置：
```javascript
// EVM 示例
const MERCHANT_ADDRESS = '0xB7aa464b19037CF3dB7F723504dFafE7b63aAb84'
const DESTINATION_EID = 40231
const DESTINATION_TOKEN = '0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d'
const AMOUNT = 1000000

// Solana 示例
const MERCHANT_ADDRESS = '7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU'
const DESTINATION_EID = 40168
const DESTINATION_TOKEN = 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v'
const AMOUNT = 1000000
```

## 常用代币地址

### Base Sepolia
- USDT: `0x036CbD53842c5426634e7929541eC2318f3dCF7e`
- USDC: `0x036CbD53842c5426634e7929541eC2318f3dCF7e`

### Arbitrum Sepolia
- USDT: `0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d`
- USDC: `0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d`

### Solana Devnet
- USDT: `Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB`
- USDC: `EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v`

## 更新日志

### v1.2.0 (2024-10-21)
- ✨ 新增 Solana Devnet 支持
- ✨ 支持 Solana base58 地址格式验证
- ✨ 更新支付链接生成器界面
- ✨ 动态地址格式提示
- ✨ 增加 Solana 代币地址示例
- 📝 更新文档和示例

### v1.1.0
- 修复 Payment ID 负数问题
- 改进 Snowflake 算法
- 增强错误处理
- 支持负数 ID 解析（向后兼容）

### v1.0.0
- 初始版本
- 支持多链支付
- SQLite 数据库集成
- RESTful API
