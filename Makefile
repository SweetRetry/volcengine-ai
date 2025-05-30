.PHONY: build run test clean docker-build docker-run dev install

# Go相关变量
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# 应用变量
BINARY_NAME=volcengine-server
BINARY_UNIX=$(BINARY_NAME)_unix
SERVER_PATH=cmd/server/main.go
WORKER_PATH=cmd/worker/main.go

# Docker变量
DOCKER_IMAGE=volcengine-go-server
DOCKER_TAG=latest

# 默认目标
all: test build

# 安装依赖
install:
	$(GOMOD) download
	$(GOMOD) tidy

# 构建应用
build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(SERVER_PATH)

# 构建所有服务
build-all:
	$(GOBUILD) -o server -v $(SERVER_PATH)
	$(GOBUILD) -o worker -v $(WORKER_PATH)

# 构建Linux版本
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v $(SERVER_PATH)

# 运行API服务器
run:
	$(GOBUILD) -o $(BINARY_NAME) -v $(SERVER_PATH)
	./$(BINARY_NAME)

# 运行任务处理中心
run-worker:
	$(GOBUILD) -o worker -v $(WORKER_PATH)
	./worker

# 开发模式运行（热重载需要安装air: go install github.com/cosmtrek/air@latest）
dev:
	air

# 测试
test:
	$(GOTEST) -v ./...

# 测试覆盖率
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out
	
# 清理
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f server worker
	rm -f coverage.out

# 清理日志文件
clean-logs:
	rm -rf logs/

# 查看日志文件
show-logs:
	@echo "=== 日志文件列表 ==="
	@ls -la logs/ 2>/dev/null || echo "logs目录不存在"
	@echo ""
	@echo "=== 最新日志内容 ==="
	@tail -20 logs/app-$(shell date +%Y-%m-%d).log 2>/dev/null || echo "今日日志文件不存在"

# 格式化代码
fmt:
	$(GOCMD) fmt ./...

# 代码检查
lint:
	golangci-lint run

# 构建Docker镜像
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# 运行Docker容器
docker-run:
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker compose启动
docker-compose-up:
	docker-compose up -d

# Docker compose停止
docker-compose-down:
	docker-compose down

# 生成API文档（需要安装swag: go install github.com/swaggo/swag/cmd/swag@latest）
docs:
	swag init -g cmd/server/main.go

# 安装开发工具
tools:
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) github.com/swaggo/swag/cmd/swag@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 部署到生产环境
deploy:
	@echo "部署到生产环境..."
	$(MAKE) build-linux
	# 这里添加具体的部署脚本

# Redis队列管理
redis-queue-status:
	@echo "查看Redis队列状态..."
	@if [ -f scripts/test_queue_status.sh ]; then \
		./scripts/test_queue_status.sh; \
	else \
		echo "scripts/test_queue_status.sh 不存在"; \
	fi

redis-queue-clear:
	@echo "清理Redis队列数据..."
	@if [ -f scripts/clear_redis_queue.sh ]; then \
		./scripts/clear_redis_queue.sh; \
	else \
		echo "scripts/clear_redis_queue.sh 不存在"; \
	fi

redis-queue-clear-force:
	@echo "强制清理Redis队列数据（无需确认）..."
	@if [ -f scripts/clear_redis_queue.sh ]; then \
		echo "y" | ./scripts/clear_redis_queue.sh; \
	else \
		echo "scripts/clear_redis_queue.sh 不存在"; \
	fi

# 帮助信息
help:
	@echo "可用的make命令:"
	@echo ""
	@echo "🏗️  构建相关:"
	@echo "  build               - 构建API服务器"
	@echo "  build-all           - 构建所有服务（API服务器 + 任务处理中心）"
	@echo "  build-linux         - 构建Linux版本"
	@echo "  install             - 安装依赖"
	@echo ""
	@echo "🚀 运行相关:"
	@echo "  run                 - 运行API服务器"
	@echo "  run-worker          - 运行任务处理中心"
	@echo "  dev                 - 开发模式运行（热重载）"
	@echo ""
	@echo "🧪 测试相关:"
	@echo "  test                - 运行测试"
	@echo "  test-coverage       - 运行测试并生成覆盖率报告"
	@echo "  test-logger         - 测试日志功能"
	@echo ""
	@echo "📝 日志相关:"
	@echo "  show-logs           - 查看日志文件"
	@echo "  clean-logs          - 清理日志文件"
	@echo ""
	@echo "🔧 开发工具:"
	@echo "  fmt                 - 格式化代码"
	@echo "  lint                - 代码检查"
	@echo "  docs                - 生成API文档"
	@echo "  tools               - 安装开发工具"
	@echo ""
	@echo "🐳 Docker相关:"
	@echo "  docker-build        - 构建Docker镜像"
	@echo "  docker-run          - 运行Docker容器"
	@echo "  docker-compose-up   - 启动Docker Compose"
	@echo "  docker-compose-down - 停止Docker Compose"
	@echo ""
	@echo "📮 Redis队列:"
	@echo "  redis-queue-status  - 查看Redis队列状态"
	@echo "  redis-queue-clear   - 清理Redis队列数据"
	@echo "  redis-queue-clear-force - 强制清理Redis队列数据（无需确认）"
	@echo ""
	@echo "🧹 清理相关:"
	@echo "  clean               - 清理构建文件"
	@echo "  clean-logs          - 清理日志文件"
	@echo ""
	@echo "ℹ️  其他:"
	@echo "  help                - 显示帮助信息"
	@echo "  deploy              - 部署到生产环境" 