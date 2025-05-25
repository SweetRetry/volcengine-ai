# 变更日志

## [2024-05-23] refactor: 删除PostgreSQL支持，仅保留MongoDB

### 移除的内容

- ✅ 删除 `internal/database/postgres.go` 文件
- ✅ 删除 `go.mod` 中的 PostgreSQL 相关依赖:
  - `github.com/lib/pq v1.10.9`
  - `gorm.io/driver/postgres v1.5.2`
  - `gorm.io/gorm v1.25.4`
- ✅ 删除 `internal/config/config.go` 中的 PostgreSQL 配置:
  - `Type` 字段
  - `PostgresURL` 字段
- ✅ 简化 `cmd/server/main.go` 中的数据库初始化逻辑
- ✅ 更新 `config.env` 删除 PostgreSQL 配置项
- ✅ 更新配置测试文件 `internal/config/config_test.go`
- ✅ 更新 `docker-compose.yml` 删除 PostgreSQL 服务
- ✅ 更新 `README.md` 和 `项目概览.md` 文档

### 保留的内容

- ✅ MongoDB 完整实现和配置
- ✅ Redis 队列系统
- ✅ 所有业务逻辑层和API接口
- ✅ 中间件和路由配置

### 环境变量变更

**之前:**
```env
DB_TYPE=postgres
POSTGRES_URL=postgres://user:password@localhost:5432/jimeng_db?sslmode=disable
MONGO_URL=mongodb://localhost:27017/jimeng_db
```

**现在:**
```env
MONGO_URL=mongodb://localhost:27017/jimeng_db
```

### 使用说明

项目现在只支持 MongoDB 作为数据库：

1. 确保 MongoDB 服务正在运行
2. 在环境变量中设置 `MONGO_URL`
3. 运行 `go run cmd/server/main.go` 启动服务

### 技术栈更新

- **数据库**: MongoDB (官方驱动)
- **队列**: Redis + Asynq
- **Web框架**: Gin
- **日志**: Logrus 