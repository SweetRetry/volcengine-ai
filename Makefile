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
MAIN_PATH=cmd/server/main.go

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
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)

# 构建Linux版本
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v $(MAIN_PATH)

# 运行应用
run:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)
	./$(BINARY_NAME)

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
	rm -f coverage.out

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
	docker run -p 8080:8080 --env-file config.env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker compose启动
docker-compose-up:
	docker-compose up -d

# Docker compose停止
docker-compose-down:
	docker-compose down

# 数据库迁移（需要安装migrate工具）
migrate-up:
	migrate -path migrations -database "$(POSTGRES_URL)" up

migrate-down:
	migrate -path migrations -database "$(POSTGRES_URL)" down

# 创建新的迁移文件
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

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
	./scripts/clear_redis_queue.sh --show-only

redis-queue-clear:
	@echo "清理Redis队列数据..."
	./scripts/clear_redis_queue.sh

redis-queue-clear-force:
	@echo "强制清理Redis队列数据（无需确认）..."
	echo "y" | ./scripts/clear_redis_queue.sh

# 测试日志功能
test-logger:
	$(GOBUILD) -o test_logger scripts/test_logger.go
	./test_logger
	rm -f test_logger

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

# 帮助信息
help:
	@echo "可用的make命令:"
	@echo "  build               - 构建应用"
	@echo "  run                 - 运行应用"
	@echo "  dev                 - 开发模式运行（热重载）"
	@echo "  test                - 运行测试"
	@echo "  test-coverage       - 运行测试并生成覆盖率报告"
	@echo "  test-logger         - 测试日志功能"
	@echo "  clean               - 清理构建文件"
	@echo "  clean-logs          - 清理日志文件"
	@echo "  show-logs           - 查看日志文件"
	@echo "  fmt                 - 格式化代码"
	@echo "  lint                - 代码检查"
	@echo "  docker-build        - 构建Docker镜像"
	@echo "  docker-run          - 运行Docker容器"
	@echo "  docker-compose-up   - 启动Docker Compose"
	@echo "  docker-compose-down - 停止Docker Compose"
	@echo "  install             - 安装依赖"
	@echo "  docs                - 生成API文档"
	@echo "  tools               - 安装开发工具"
	@echo "  redis-queue-status  - 查看Redis队列状态"
	@echo "  redis-queue-clear   - 清理Redis队列数据"
	@echo "  redis-queue-clear-force - 强制清理Redis队列数据（无需确认）"
	@echo "  help                - 显示帮助信息" 