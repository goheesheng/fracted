# LayerZero Cross-Chain Payout Indexer Makefile

# 变量定义
APP_NAME = cross-chain-indexer
VERSION = 1.0.0
BUILD_TIME = $(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT = $(shell git rev-parse --short HEAD)
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Go相关变量
GO = go
GOFMT = gofmt
GOIMPORTS = goimports
GOLINT = golangci-lint
GOTEST = go test
GOCOVER = go test -coverprofile=coverage.out

# 目录变量
BIN_DIR = bin
DIST_DIR = dist
DOCS_DIR = docs
SCRIPTS_DIR = scripts

# 默认目标
.PHONY: all
all: clean build test

# 帮助信息
.PHONY: help
help:
	@echo "LayerZero Cross-Chain Payout Indexer Makefile"
	@echo ""
	@echo "可用目标:"
	@echo "  build         构建应用程序"
	@echo "  test          运行测试"
	@echo "  test-coverage 运行测试并生成覆盖率报告"
	@echo "  lint          代码检查"
	@echo "  fmt           格式化代码"
	@echo "  clean         清理构建文件"
	@echo "  run           运行应用程序"
	@echo "  docker        构建Docker镜像"
	@echo "  docker-run    运行Docker容器"
	@echo "  deploy        部署到生产环境"
	@echo "  backup        备份数据"
	@echo "  install       安装依赖"
	@echo "  help          显示此帮助信息"

# 安装依赖
.PHONY: install
install:
	@echo "安装Go依赖..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "安装开发工具..."
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "依赖安装完成"

# 构建应用程序
.PHONY: build
build:
	@echo "构建应用程序..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) .
	@echo "构建完成: $(BIN_DIR)/$(APP_NAME)"

# 构建多平台版本
.PHONY: build-all
build-all:
	@echo "构建多平台版本..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 .
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 .
	@echo "多平台构建完成"

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	$(GOTEST) -v ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	$(GOCOVER) ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成完成: coverage.html"

# 代码检查
.PHONY: lint
lint:
	@echo "运行代码检查..."
	$(GOLINT) run

# 格式化代码
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	$(GOFMT) -s -w .
	$(GOIMPORTS) -w .

# 清理构建文件
.PHONY: clean
clean:
	@echo "清理构建文件..."
	rm -rf $(BIN_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html
	$(GO) clean

# 运行应用程序
.PHONY: run
run:
	@echo "运行应用程序..."
	$(GO) run .

# 运行开发模式（热重载）
.PHONY: dev
dev:
	@echo "启动开发模式..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "请先安装air: go install github.com/cosmtrek/air@latest"; \
		$(GO) run .; \
	fi

# 构建Docker镜像
.PHONY: docker
docker:
	@echo "构建Docker镜像..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker build -t $(APP_NAME):latest .

# 运行Docker容器
.PHONY: docker-run
docker-run:
	@echo "运行Docker容器..."
	docker run -d \
		--name $(APP_NAME) \
		-p 8080:8080 \
		-e JWT_SECRET=dev-secret-key \
		-e ADMIN_ADDRESSES=0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6 \
		-e MERCHANT_ADDRESSES=0x77Ed7f6455FE291728A48785090292e3D10F53Bb \
		-v $(PWD)/data:/app/data \
		$(APP_NAME):latest

# 停止Docker容器
.PHONY: docker-stop
docker-stop:
	@echo "停止Docker容器..."
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

# 使用Docker Compose运行
.PHONY: compose-up
compose-up:
	@echo "启动Docker Compose服务..."
	docker-compose up -d

# 停止Docker Compose服务
.PHONY: compose-down
compose-down:
	@echo "停止Docker Compose服务..."
	docker-compose down

# 查看Docker Compose日志
.PHONY: compose-logs
compose-logs:
	@echo "查看Docker Compose日志..."
	docker-compose logs -f

# 部署到生产环境
.PHONY: deploy
deploy:
	@echo "部署到生产环境..."
	@if [ -f "$(SCRIPTS_DIR)/deploy.sh" ]; then \
		chmod +x $(SCRIPTS_DIR)/deploy.sh; \
		$(SCRIPTS_DIR)/deploy.sh update; \
	else \
		echo "部署脚本不存在: $(SCRIPTS_DIR)/deploy.sh"; \
	fi

# 备份数据
.PHONY: backup
backup:
	@echo "备份数据..."
	@if [ -f "$(SCRIPTS_DIR)/backup.sh" ]; then \
		chmod +x $(SCRIPTS_DIR)/backup.sh; \
		$(SCRIPTS_DIR)/backup.sh; \
	else \
		echo "备份脚本不存在: $(SCRIPTS_DIR)/backup.sh"; \
	fi

# 安装系统服务
.PHONY: install-service
install-service:
	@echo "安装系统服务..."
	@if [ -f "$(SCRIPTS_DIR)/$(APP_NAME).service" ]; then \
		sudo cp $(SCRIPTS_DIR)/$(APP_NAME).service /etc/systemd/system/; \
		sudo systemctl daemon-reload; \
		sudo systemctl enable $(APP_NAME); \
		echo "系统服务安装完成"; \
	else \
		echo "服务文件不存在: $(SCRIPTS_DIR)/$(APP_NAME).service"; \
	fi

# 启动系统服务
.PHONY: start-service
start-service:
	@echo "启动系统服务..."
	sudo systemctl start $(APP_NAME)
	sudo systemctl status $(APP_NAME)

# 停止系统服务
.PHONY: stop-service
stop-service:
	@echo "停止系统服务..."
	sudo systemctl stop $(APP_NAME)

# 重启系统服务
.PHONY: restart-service
restart-service:
	@echo "重启系统服务..."
	sudo systemctl restart $(APP_NAME)

# 查看服务日志
.PHONY: logs
logs:
	@echo "查看服务日志..."
	journalctl -u $(APP_NAME) -f

# 生成API文档
.PHONY: docs
docs:
	@echo "生成API文档..."
	@if command -v swag > /dev/null; then \
		swag init -g main.go -o $(DOCS_DIR)/swagger; \
	else \
		echo "请先安装swag: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# 安全检查
.PHONY: security
security:
	@echo "运行安全检查..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "请先安装gosec: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 性能测试
.PHONY: bench
bench:
	@echo "运行性能测试..."
	$(GOTEST) -bench=. -benchmem ./...

# 内存分析
.PHONY: profile
profile:
	@echo "运行内存分析..."
	$(GOTEST) -memprofile=mem.prof -cpuprofile=cpu.prof ./...
	$(GO) tool pprof mem.prof

# 检查依赖更新
.PHONY: deps-update
deps-update:
	@echo "检查依赖更新..."
	$(GO) list -u -m all

# 更新依赖
.PHONY: deps-upgrade
deps-upgrade:
	@echo "更新依赖..."
	$(GO) get -u ./...
	$(GO) mod tidy

# 发布版本
.PHONY: release
release: clean build-all test lint
	@echo "发布版本 $(VERSION)..."
	@mkdir -p $(DIST_DIR)
	cd $(DIST_DIR) && \
	for file in *; do \
		tar -czf $${file}.tar.gz $$file; \
	done
	@echo "发布文件已生成在 $(DIST_DIR) 目录"

# 检查代码质量
.PHONY: quality
quality: fmt lint test
	@echo "代码质量检查完成"

# 完整构建流程
.PHONY: ci
ci: clean install fmt lint test build
	@echo "CI构建流程完成"
