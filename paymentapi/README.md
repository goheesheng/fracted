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

**请求:**
```
GET /generate-link?merchant=0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6&dstEid=40245&dstToken=0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d&amount=123000000
```

**响应:**
```json
{
  "success": true,
  "paymentId": "1234567890123456789",
  "paymentLink": "https://demo.fracted.xyz/payment/1234567890123456789",
  "parameters": {
    "merchant": "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
    "dstEid": 40245,
    "dstToken": "0x75faf114eafb1BDbe2F0316DF893fd58CE46AA4d",
    "amount": 123000000
  }
}
```

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
| merchant_address | TEXT | 商户地址 |
| dst_eid | INTEGER | 目标链 ID |
| dst_token | TEXT | 目标代币地址 |
| amount | TEXT | 支付金额 |
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

### Ethereum 网络
- Arbitrum Sepolia
- Base Sepolia
- Solana Devnet

### 支持的钱包
- MetaMask (Ethereum 网络)
- Phantom (Solana 网络)

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
OAPP_arbitrum_sepolia=your_contract_address
OAPP_base_sepolia=your_contract_address
TOKEN_arbitrum_sepolia_USDC=your_token_address
TOKEN_base_sepolia_USDT=your_token_address
```

### 生产环境
```bash
npm start
```

## 更新日志

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
