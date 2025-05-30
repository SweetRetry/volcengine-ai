# API调用日志说明

## 📋 概述

本项目已经实现了完整的API调用日志记录系统，能够详细记录所有火山引擎AI服务的API调用情况，包括请求参数、响应状态、执行时间等关键信息。

## 🚀 功能特性

### 详细的API调用记录
- **请求参数记录**: 完整记录API调用的所有参数
- **响应状态记录**: 记录HTTP状态码和响应内容
- **执行时间统计**: 精确记录API调用耗时（毫秒级）
- **错误信息记录**: 详细记录API调用失败的原因

### 结构化日志格式
- **JSON格式**: 便于日志分析和处理
- **时间戳**: 精确到秒的时间记录
- **日志级别**: INFO、WARN、ERROR等不同级别
- **结构化字段**: 使用logrus.Fields记录结构化信息

### 多层次日志记录
- **HTTP请求日志**: 记录所有HTTP请求和响应
- **业务逻辑日志**: 记录任务处理过程
- **API调用日志**: 记录第三方API调用详情
- **错误处理日志**: 记录异常和错误信息

## 📊 日志内容示例

### 火山方舟API调用日志

#### 成功调用示例
```json
{
  "api_endpoint": "GenerateImages",
  "level": "info",
  "model": "doubao-seedream-3-0-t2i-250415",
  "msg": "火山方舟API调用开始",
  "prompt": "一只可爱的小猫",
  "size": "1024x1024",
  "time": "2025-05-31 01:12:21",
  "watermark": false
}

{
  "api_endpoint": "GenerateImages",
  "duration_ms": 1250,
  "level": "info",
  "msg": "火山方舟API调用成功",
  "response_count": 1,
  "time": "2025-05-31 01:12:22"
}
```

#### 失败调用示例
```json
{
  "api_endpoint": "GenerateImages",
  "duration_ms": 252,
  "error": "Error code: 401 - Authentication Error",
  "level": "error",
  "msg": "火山方舟API调用失败",
  "time": "2025-05-31 01:12:21"
}
```

### 即梦AI API调用日志

#### CVProcess调用示例
```json
{
  "api_endpoint": "CVProcess",
  "height": "512",
  "level": "info",
  "msg": "即梦AI API调用开始",
  "prompt": "一只可爱的小猫",
  "req_key": "jimeng_high_aes_general_v21_L",
  "return_url": true,
  "time": "2025-05-31 01:12:21",
  "use_pre_llm": false,
  "use_sr": true,
  "width": "512"
}

{
  "api_endpoint": "CVProcess",
  "duration_ms": 850,
  "level": "info",
  "msg": "即梦AI API调用成功",
  "response": {
    "code": 10000,
    "data": {
      "image_urls": ["https://example.com/image.jpg"]
    }
  },
  "status_code": 200,
  "time": "2025-05-31 01:12:22"
}
```

#### CVSubmitTask调用示例（视频生成）
```json
{
  "api_endpoint": "CVSubmitTask",
  "aspect_ratio": "16:9",
  "level": "info",
  "msg": "即梦AI视频API调用开始",
  "prompt": "一只小猫在花园里玩耍",
  "req_key": "jimeng_vgfm_t2v_l20",
  "seed": 123,
  "time": "2025-05-31 01:12:21"
}

{
  "api_endpoint": "CVSubmitTask",
  "duration_ms": 450,
  "level": "info",
  "msg": "即梦AI视频API调用成功",
  "response": {
    "code": 10000,
    "data": {
      "task_id": "12345678"
    }
  },
  "status_code": 200,
  "time": "2025-05-31 01:12:22"
}
```

#### CVGetResult调用示例（结果查询）
```json
{
  "api_endpoint": "CVGetResult",
  "level": "info",
  "msg": "即梦AI视频结果查询API调用开始",
  "req_key": "jimeng_vgfm_t2v_l20",
  "task_id": "12345678",
  "time": "2025-05-31 01:12:31"
}

{
  "api_endpoint": "CVGetResult",
  "duration_ms": 120,
  "level": "info",
  "msg": "即梦AI视频结果查询API调用成功",
  "response": {
    "code": 10000,
    "data": {
      "status": "done",
      "video_url": "https://example.com/video.mp4"
    }
  },
  "status_code": 200,
  "task_id": "12345678",
  "time": "2025-05-31 01:12:31"
}
```

### HTTP请求日志

```json
{
  "body_size": 156,
  "client_ip": "127.0.0.1",
  "latency_ms": 1250,
  "level": "info",
  "method": "POST",
  "msg": "HTTP请求",
  "path": "/api/v1/ai/image/task",
  "query_params": "",
  "status_code": 200,
  "timestamp": "2025-05-31T01:12:21+08:00",
  "user_agent": "curl/7.68.0"
}
```

## 🔧 配置说明

### 环境变量配置

```bash
# 日志级别 (debug, info, warn, error)
LOG_LEVEL=info

# 日志保留天数
LOG_KEEP_DAYS=7
```

### 日志级别说明

- **DEBUG**: 详细的调试信息（开发环境）
- **INFO**: 一般信息记录（生产环境推荐）
- **WARN**: 警告信息
- **ERROR**: 错误信息

## 📁 日志文件结构

```
logs/
├── app-2025-05-31.log    # 今天的日志
├── app-2025-05-30.log    # 昨天的日志
└── ...                   # 历史日志文件
```

## 🔍 日志分析方法

### 查看实时日志
```bash
# 查看今天的日志
tail -f logs/app-$(date +%Y-%m-%d).log

# 查看最新100行日志
tail -100 logs/app-$(date +%Y-%m-%d).log
```

### 搜索特定API调用
```bash
# 搜索火山方舟API调用
grep "GenerateImages" logs/app-*.log

# 搜索即梦AI API调用
grep "CVProcess\|CVSubmitTask\|CVGetResult" logs/app-*.log

# 搜索API调用失败
grep "API调用失败" logs/app-*.log
```

### 分析API性能
```bash
# 查看API调用耗时
grep "duration_ms" logs/app-*.log | jq '.duration_ms'

# 查看慢查询（超过1秒）
grep "duration_ms" logs/app-*.log | jq 'select(.duration_ms > 1000)'
```

### 统计API调用次数
```bash
# 统计各API端点调用次数
grep "api_endpoint" logs/app-*.log | jq -r '.api_endpoint' | sort | uniq -c

# 统计成功和失败的API调用
grep "API调用" logs/app-*.log | grep -c "成功"
grep "API调用" logs/app-*.log | grep -c "失败"
```

## 📊 监控指标

### 关键指标
- **API调用成功率**: 成功调用数 / 总调用数
- **平均响应时间**: 所有API调用的平均耗时
- **错误率**: 失败调用数 / 总调用数
- **QPS**: 每秒查询数

### 告警阈值建议
- **响应时间**: > 5秒
- **错误率**: > 5%
- **连续失败**: > 3次

## 🛠️ 故障排查

### 常见问题

1. **API调用失败**
   ```bash
   # 查看具体错误信息
   grep "API调用失败" logs/app-*.log | tail -10
   ```

2. **响应时间过长**
   ```bash
   # 查看慢查询
   grep "duration_ms" logs/app-*.log | jq 'select(.duration_ms > 5000)'
   ```

3. **认证错误**
   ```bash
   # 查看认证相关错误
   grep "AuthenticationError\|401" logs/app-*.log
   ```

## 🔒 安全考虑

### 敏感信息处理
- **API密钥**: 不记录完整的API密钥
- **用户数据**: 避免记录敏感的用户信息
- **响应内容**: 大型响应内容会被截断

### 日志访问控制
```bash
# 设置适当的文件权限
chmod 640 logs/*.log
chown app:app logs/*.log
```

## 📈 性能优化

### 日志性能优化
- **异步写入**: 使用缓冲写入减少I/O开销
- **日志轮转**: 定期轮转日志文件避免单文件过大
- **压缩存储**: 历史日志文件可以压缩存储

### 存储优化
```bash
# 压缩历史日志
gzip logs/app-2025-05-*.log

# 清理过期日志
find logs/ -name "*.log" -mtime +7 -delete
```

## 🎯 最佳实践

1. **结构化日志**: 使用JSON格式便于分析
2. **适当级别**: 生产环境使用INFO级别
3. **关键信息**: 记录请求ID、用户ID等关键标识
4. **性能监控**: 定期分析API调用性能
5. **错误追踪**: 建立错误告警机制
6. **日志轮转**: 定期清理历史日志文件 