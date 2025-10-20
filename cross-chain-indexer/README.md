# 🔗 LayerZero 跨链支付索引器

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Production%20Ready-brightgreen.svg)]()

一个基于LayerZero协议的跨链支付索引器，支持Base、Arbitrum和Solana多链交易监听，为商家和管理员提供安全的交易数据访问服务。

## 🎯 项目概述

本项目是一个跨链支付索引器，主要功能包括：

- 🔍 **实时监听** Base、Arbitrum和Solana链上的交易事件
- 🌐 **多链支持** Base Sepolia ↔ Arbitrum Sepolia ↔ Solana Devnet
- 🔐 **安全认证** 基于JWT的身份验证和授权
- 👥 **角色管理** 支持管理员和商家两种角色
- 🛡️ **白名单机制** 严格的地址白名单访问控制
- 📊 **数据展示** 现代化的Web Dashboard界面
- 🗄️ **数据存储** SQLite数据库持久化存储
- ⚡ **智能地址** 自动识别和转换EVM与Solana地址格式

## ✨ 核心特性

### 🔒 安全机制
- **JWT认证**: 基于JSON Web Token的安全认证
- **角色权限**: 管理员和商家分离的权限体系
- **地址白名单**: 严格的地址访问控制（支持EVM和Solana）
- **动态管理**: 运行时添加/移除管理员和商家

### 📈 数据管理
- **实时索引**: 监听区块链事件并实时存储
- **历史回填**: 支持历史区块数据回填
- **数据查询**: 高效的数据库查询和分页
- **状态跟踪**: 完整的交易状态跟踪

### 🌐 用户界面
- **现代化Dashboard**: 响应式Web界面
- **实时数据**: 自动刷新的交易数据
- **用户友好**: 直观的操作界面
- **多角色支持**: 不同角色的定制化界面
- **智能地址显示**: Solana显示Base58，EVM显示0x格式

## 🚀 快速开始

### 环境要求

- Go 1.21+
- SQLite3
- 网络连接（用于区块链RPC）

### 5分钟快速上手

#### 1. 克隆项目
```bash
git clone <repository-url>
cd cross-chain-indexer
```

#### 2. 配置环境变量
```bash
cp env.example .env
# 编辑 .env 文件，设置你的配置
```

#### 3. 启动服务

**Windows**:
```powershell
.\scripts\start.ps1
```

**Linux**:
```bash
chmod +x scripts/*.sh
./scripts/start.sh
```

#### 4. 访问Dashboard
```
http://localhost:8080/dashboard/
```

#### 5. 登录

**管理员登录**（查看所有数据）:
```
地址: 0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6
角色: admin
```

**Solana商家登录**（查看个人数据）:
```
地址: 6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp
角色: merchant
```

#### 6. 停止服务

**Windows**:
```powershell
.\scripts\stop.ps1
```

**Linux**:
```bash
./scripts/stop.sh
```

## 🌐 支持的链

| 链名称 | 网络 | EID | 地址格式 | 监听类型 |
|--------|------|-----|---------|---------|
| Base Sepolia | 测试网 | 40245 | 0x | WSS事件监听 |
| Arbitrum Sepolia | 测试网 | 40231 | 0x | 状态查询 |
| Solana Devnet | 测试网 | 40168 | Base58 | WS交易监听 |

## 🔐 支持的地址格式

### EVM地址
- **格式**: `0x` + 40个十六进制字符
- **示例**: `0x77Ed7f6455FE291728A48785090292e3D10F53Bb`
- **用途**: Base、Arbitrum等EVM链

### Solana地址
- **格式**: 32-44个Base58字符
- **示例**: `6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp`
- **用途**: Solana链商家和交易

系统会**自动识别**地址类型，无需用户指定！

## 📚 文档

### 完整文档
- **[技术文档](docs/TECHNICAL.md)** - 完整的API文档、架构设计、部署指南
- **[用户指南](docs/USER_GUIDE.md)** - 使用说明、登录指南、常见问题

### Solana集成
本项目完整支持Solana `transfer_contract` 程序监听：

**程序地址**: `GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1`

**查看Solana交易**:
- Dashboard中带有绿色"Solana Devnet"标签
- 商家地址显示为Base58格式
- 自动识别USDC/USDT代币

**在Solana Explorer验证**:
```
https://explorer.solana.com/address/GSPmsxkxd5qR5HG4fhUd5cBrVkWNJWi6pWUFQnYmTEc1?cluster=devnet
```

## 🏗️ 项目结构

```
cross-chain-indexer/
├── 📁 docs/                    # 项目文档
│   ├── TECHNICAL.md            # 技术文档
│   └── USER_GUIDE.md           # 用户指南
├── 📁 contract/                # 智能合约绑定
│   ├── myoapp.go              # Base -> Arb 合约
│   └── solana/                # Solana 合约相关
├── 📁 dashboard/               # 前端Dashboard
│   ├── index.html             # 管理员Dashboard
│   ├── login.html             # 商家登录页
│   ├── merchant-dashboard.html # 商家Dashboard
│   ├── app.js                 # 管理员JS
│   ├── merchant-app.js        # 商家JS
│   └── styles.css             # 样式
├── 📁 scripts/                 # 运维脚本
│   ├── start.sh               # Linux启动
│   ├── stop.sh                # Linux停止
│   ├── start.ps1              # Windows启动
│   └── stop.ps1               # Windows停止
├── 📄 main.go                  # 主程序入口
├── 📄 api.go                   # API服务器和路由
├── 📄 config.go                # 配置管理
├── 📄 store.go                 # 数据库存储
├── 📄 processor.go             # EVM链事件处理器
├── 📄 solana_listener.go       # Solana链监听器
├── 📄 status_updater.go        # 状态更新器
├── 📄 api_test.go              # API测试
├── 📄 security_test.go         # 安全测试
├── 📄 Dockerfile               # Docker镜像
├── 📄 docker-compose.yml       # Docker编排
└── 📄 README.md                # 本文件
```

## 🔧 配置说明

### 环境变量

创建 `.env` 文件：

```bash
# JWT密钥（生产环境必须修改）
JWT_SECRET=your-super-secret-jwt-key-32-chars-min

# 管理员地址（逗号分隔，EVM格式）
ADMIN_ADDRESSES=0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6

# 商家地址（逗号分隔，支持EVM和Solana混合）
MERCHANT_ADDRESSES=0x77Ed7f6455FE291728A48785090292e3D10F53Bb,6H7AYKpUHnMuca92gc82oArXC48igkLi14mcZh9XNLpp
```

## 🐳 Docker部署

### 简单部署
```bash
# 构建镜像
docker build -t cross-chain-indexer .

# 运行容器
docker run -p 8080:8080 \
  -e JWT_SECRET=your-secret-key \
  -e ADMIN_ADDRESSES=0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6 \
  cross-chain-indexer
```

### Docker Compose部署（推荐）
```bash
# 启动完整环境（含Nginx、监控）
docker-compose up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 🧪 测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定测试
go test -v -run TestAdminSecurity

# 测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 Dashboard功能

### 管理员Dashboard
- 查看所有商家的交易
- 按商家统计总收入
- 按付款方统计总支出
- 实时交易列表
- 搜索和筛选功能

### 商家Dashboard
- 查看个人交易记录
- 交易统计和汇总
- 按代币分类显示
- 实时更新

### 交易详情
- 点击任意交易查看完整JSON数据
- Solana交易自动显示Base58地址
- EVM交易显示0x地址
- 链接到区块浏览器

## 🔍 监控和日志

### 日志文件
- `solana_log.txt` - Solana交易处理日志
- 控制台输出 - 所有链的事件日志

### 实时监控
```bash
# Windows - 查看Solana日志
Get-Content solana_log.txt -Wait -Tail 20

# Linux - 查看Solana日志
tail -f solana_log.txt
```

### 健康检查
```bash
curl http://localhost:8080/health
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 🆘 故障排查

### 常见问题

**Q: 服务无法启动？**
```bash
# 检查端口占用
netstat -ano | findstr :8080  # Windows
lsof -i :8080                 # Linux

# 停止占用进程
.\scripts\stop.ps1  # Windows
./scripts/stop.sh   # Linux
```

**Q: Dashboard看不到数据？**
1. 确保已登录（右上角显示用户信息）
2. 清除缓存：浏览器控制台执行 `localStorage.clear(); location.reload()`
3. 检查数据库：`sqlite3 indexer.db "SELECT COUNT(*) FROM payouts;"`

**Q: Solana地址显示为0x格式？**
1. 刷新浏览器（F5）
2. 清除缓存并重新登录
3. 确认服务已重启（数据库会自动迁移）

更多问题请查看 [用户指南](docs/USER_GUIDE.md) 的故障排查部分。

## 📝 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

如果你遇到任何问题或有任何建议，请：

1. 查看 [技术文档](docs/TECHNICAL.md) 和 [用户指南](docs/USER_GUIDE.md)
2. 搜索 [Issues](../../issues)
3. 创建新的 [Issue](../../issues/new)

## 🏆 致谢

- [LayerZero](https://layerzero.network/) - 跨链协议
- [Go Ethereum](https://geth.ethereum.org/) - 以太坊Go客户端
- [Solana Go SDK](https://github.com/gagliardetto/solana-go) - Solana Go客户端
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP路由库

---

**⭐ 如果这个项目对你有帮助，请给它一个星标！**
