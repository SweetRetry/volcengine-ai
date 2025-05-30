# 🏗️ 构建阶段
FROM golang:1.21-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建API服务器
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server

# 构建队列工作器
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/worker ./cmd/worker

# 🚀 API服务器镜像
FROM alpine:latest AS server

# 安装ca-certificates和tzdata
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/server .

# 创建日志目录
RUN mkdir -p logs && chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动命令
CMD ["./server"]

# ⚡ 队列工作器镜像
FROM alpine:latest AS worker

# 安装ca-certificates和tzdata
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/worker .

# 创建日志目录
RUN mkdir -p logs && chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# Worker服务不需要暴露端口，它通过Redis队列处理任务

# 健康检查 - 检查进程是否运行
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD pgrep -f worker || exit 1

# 启动命令
CMD ["./worker"]

# 🔧 开发环境镜像
FROM golang:1.21-alpine AS development

# 安装开发工具
RUN apk add --no-cache git make curl

# 安装air用于热重载
RUN go install github.com/cosmtrek/air@latest

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 暴露端口
EXPOSE 8080

# 启动命令（使用air进行热重载）
CMD ["air", "-c", ".air.toml"] 