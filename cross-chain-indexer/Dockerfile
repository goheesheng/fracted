# 多阶段构建 Dockerfile for LayerZero Cross-Chain Payout Indexer

# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cross-chain-indexer .

# 运行阶段
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata sqlite

# 创建非root用户
RUN addgroup -g 1001 -S indexer && \
    adduser -u 1001 -S indexer -G indexer

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/cross-chain-indexer .

# 复制Dashboard文件
COPY --from=builder /app/dashboard ./dashboard/

# 复制配置文件
COPY --from=builder /app/env.example .env.example

# 创建数据目录
RUN mkdir -p /app/data && \
    chown -R indexer:indexer /app

# 切换到非root用户
USER indexer

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 设置环境变量
ENV DB_PATH=/app/data/indexer.db
ENV JWT_SECRET=dev-local-secret-change-me
ENV ADMIN_ADDRESSES=0x27f9b6a7c1fd66ac4d0e76a2d43b35e8590165f6
ENV MERCHANT_ADDRESSES=0x77ed7f6455fe291728a48785090292e3d10f53bb

# 启动命令
CMD ["./cross-chain-indexer"]
