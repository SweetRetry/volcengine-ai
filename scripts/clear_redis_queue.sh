#!/bin/bash

# Redis队列清理脚本
# 用于清理asynq队列中的所有数据

set -e

echo "🧹 开始清理Redis队列数据..."
echo "================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Redis配置
REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}
REDIS_DB=${REDIS_DB:-0}

# 检查Redis连接
check_redis() {
    echo -e "${BLUE}🔍 检查Redis连接...${NC}"
    
    if ! redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB ping > /dev/null 2>&1; then
        echo -e "${RED}❌ 错误: 无法连接到Redis服务器${NC}"
        echo "   请检查Redis服务是否运行: redis-cli ping"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Redis连接正常${NC}"
}

# 显示当前队列状态
show_queue_status() {
    echo -e "${BLUE}📊 当前队列状态:${NC}"
    
    # 获取所有asynq相关的键
    local keys=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB keys "*asynq*" 2>/dev/null || echo "")
    
    if [ -z "$keys" ]; then
        echo -e "${GREEN}   队列为空${NC}"
        return
    fi
    
    # 分类显示不同类型的键
    echo -e "${BLUE}   🏗️  架构相关:${NC}"
    echo "$keys" | grep -E "(asynq:servers|asynq:workers|asynq:queues)" | while read -r key; do
        if [ -n "$key" ]; then
            local key_type=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB type "$key" 2>/dev/null || echo "unknown")
            local count=$(get_key_count "$key" "$key_type")
            local description=$(get_key_description "$key")
            echo -e "     ${YELLOW}$key${NC} (${key_type}): $count - $description"
        fi
    done
    
    echo -e "${BLUE}   📋 队列任务:${NC}"
    echo "$keys" | grep -E "(pending|active|retry|archived|completed)" | while read -r key; do
        if [ -n "$key" ]; then
            local key_type=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB type "$key" 2>/dev/null || echo "unknown")
            local count=$(get_key_count "$key" "$key_type")
            local description=$(get_key_description "$key")
            echo -e "     ${YELLOW}$key${NC} (${key_type}): $count - $description"
        fi
    done
    
    echo -e "${BLUE}   📊 统计数据:${NC}"
    echo "$keys" | grep -E "(processed|failed)" | grep -v ":t:" | while read -r key; do
        if [ -n "$key" ]; then
            local key_type=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB type "$key" 2>/dev/null || echo "unknown")
            local count=$(get_key_count "$key" "$key_type")
            local description=$(get_key_description "$key")
            echo -e "     ${YELLOW}$key${NC} (${key_type}): $count - $description"
        fi
    done
    
    echo -e "${BLUE}   🔧 任务数据:${NC}"
    echo "$keys" | grep ":t:" | while read -r key; do
        if [ -n "$key" ]; then
            local key_type=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB type "$key" 2>/dev/null || echo "unknown")
            local count=$(get_key_count "$key" "$key_type")
            echo -e "     ${YELLOW}$key${NC} (${key_type}): $count - 任务详细数据"
        fi
    done
}

# 获取键的计数
get_key_count() {
    local key="$1"
    local key_type="$2"
    
    case $key_type in
        "zset")
            redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB zcard "$key" 2>/dev/null || echo "0"
            ;;
        "list")
            redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB llen "$key" 2>/dev/null || echo "0"
            ;;
        "set")
            redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB scard "$key" 2>/dev/null || echo "0"
            ;;
        "hash")
            redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB hlen "$key" 2>/dev/null || echo "0"
            ;;
        "string")
            # 对于服务器实例详情，只显示存在标记
            if [[ "$key" == *":servers:"* ]]; then
                echo "存在"
            else
                local value=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB get "$key" 2>/dev/null || echo "0")
                echo "$value"
            fi
            ;;
        *)
            echo "N/A"
            ;;
    esac
}

# 获取键的描述
get_key_description() {
    local key="$1"
    
    case "$key" in
        "asynq:servers")
            echo "活跃服务器实例"
            ;;
        "asynq:workers")
            echo "活跃工作器"
            ;;
        "asynq:queues")
            echo "已知队列名称"
            ;;
        *":pending")
            echo "等待处理的任务"
            ;;
        *":active")
            echo "正在处理的任务"
            ;;
        *":retry")
            echo "等待重试的任务"
            ;;
        *":archived")
            echo "归档的失败任务（已跳过重试）"
            ;;
        *":completed")
            echo "已完成的任务"
            ;;
        *":processed")
            echo "已处理任务计数"
            ;;
        *":failed")
            echo "失败任务计数"
            ;;
        *":servers:"*)
            echo "服务器实例详情"
            ;;
        *)
            echo "其他数据"
            ;;
    esac
}

# 清理队列数据
clear_queue_data() {
    echo -e "${BLUE}🗑️  开始清理队列数据...${NC}"
    
    # 获取所有asynq相关的键
    local keys=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB keys "*asynq*" 2>/dev/null || echo "")
    
    if [ -z "$keys" ]; then
        echo -e "${GREEN}   没有需要清理的数据${NC}"
        return
    fi
    
    local deleted_count=0
    
    echo "$keys" | while read -r key; do
        if [ -n "$key" ]; then
            if redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "$key" > /dev/null 2>&1; then
                echo -e "   ${GREEN}✅ 已删除: $key${NC}"
                deleted_count=$((deleted_count + 1))
            else
                echo -e "   ${RED}❌ 删除失败: $key${NC}"
            fi
        fi
    done
    
    echo -e "${GREEN}🎉 队列数据清理完成${NC}"
}

# 清理特定类型的队列
clear_specific_queues() {
    echo -e "${BLUE}🎯 清理特定队列类型...${NC}"
    
    # 清理重试队列
    echo -e "${YELLOW}   清理重试队列...${NC}"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:{default}:retry" > /dev/null 2>&1 || true
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:{critical}:retry" > /dev/null 2>&1 || true
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:{low}:retry" > /dev/null 2>&1 || true
    
    # 清理失败队列
    echo -e "${YELLOW}   清理失败队列...${NC}"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:{default}:failed" > /dev/null 2>&1 || true
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:{critical}:failed" > /dev/null 2>&1 || true
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:{low}:failed" > /dev/null 2>&1 || true
    
    # 清理已处理队列
    echo -e "${YELLOW}   清理已处理队列...${NC}"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:{default}:processed" > /dev/null 2>&1 || true
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:{critical}:processed" > /dev/null 2>&1 || true
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:{low}:processed" > /dev/null 2>&1 || true
    
    # 清理统计数据
    echo -e "${YELLOW}   清理统计数据...${NC}"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB keys "asynq:{*}:processed:*" | xargs -r redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del > /dev/null 2>&1 || true
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB keys "asynq:{*}:failed:*" | xargs -r redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del > /dev/null 2>&1 || true
    
    # 清理任务数据
    echo -e "${YELLOW}   清理任务数据...${NC}"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB keys "asynq:{*}:t:*" | xargs -r redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del > /dev/null 2>&1 || true
    
    # 清理服务器和工作器信息
    echo -e "${YELLOW}   清理服务器信息...${NC}"
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:servers" > /dev/null 2>&1 || true
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:workers" > /dev/null 2>&1 || true
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $REDIS_DB del "asynq:queues" > /dev/null 2>&1 || true
}

# 主函数
main() {
    echo -e "${BLUE}Redis队列清理工具${NC}"
    echo "使用方法: $0 [选项]"
    echo "选项:"
    echo "  --show-only    只显示队列状态，不清理"
    echo "  --help         显示帮助信息"
    echo ""
    
    # 解析参数
    case "${1:-}" in
        "--show-only")
            check_redis
            show_queue_status
            exit 0
            ;;
        "--help")
            echo "Redis队列清理工具"
            echo ""
            echo "环境变量:"
            echo "  REDIS_HOST     Redis主机地址 (默认: localhost)"
            echo "  REDIS_PORT     Redis端口 (默认: 6379)"
            echo "  REDIS_DB       Redis数据库编号 (默认: 0)"
            echo ""
            echo "示例:"
            echo "  $0                    # 清理所有队列数据"
            echo "  $0 --show-only       # 只显示队列状态"
            echo "  REDIS_HOST=redis.example.com $0  # 使用自定义Redis主机"
            exit 0
            ;;
    esac
    
    # 执行清理
    check_redis
    
    echo -e "${YELLOW}⚠️  警告: 即将清理所有asynq队列数据!${NC}"
    echo "清理前状态:"
    show_queue_status
    echo ""
    
    read -p "确认要清理所有队列数据吗? (y/N): " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        clear_specific_queues
        clear_queue_data
        echo ""
        echo "清理后状态:"
        show_queue_status
        echo ""
        echo -e "${GREEN}🎉 队列清理完成!${NC}"
    else
        echo -e "${YELLOW}❌ 操作已取消${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@" 