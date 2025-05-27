#!/bin/bash

# 测试队列状态处理脚本
# 用于验证失败任务的正确处理

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

API_BASE="http://localhost:8080/api/v1"

echo -e "${BLUE}🧪 测试队列状态处理${NC}"
echo "================================"

# 检查服务器是否运行
check_server() {
    echo -e "${BLUE}🔍 检查服务器状态...${NC}"
    
    if ! curl -s "$API_BASE/../health" > /dev/null 2>&1; then
        echo -e "${RED}❌ 服务器未运行，请先启动服务器${NC}"
        echo "   运行: make dev"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 服务器运行正常${NC}"
}

# 创建测试用户
create_test_user() {
    echo -e "${BLUE}👤 创建测试用户...${NC}"
    
    USER_RESPONSE=$(curl -s -X POST "$API_BASE/users" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "test@queue-status.com",
            "name": "Queue Test User"
        }')
    
    USER_ID=$(echo "$USER_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    
    if [ -z "$USER_ID" ]; then
        echo -e "${RED}❌ 创建用户失败${NC}"
        echo "响应: $USER_RESPONSE"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 用户创建成功: $USER_ID${NC}"
}

# 测试正常任务
test_normal_task() {
    echo -e "${BLUE}✅ 测试正常任务处理...${NC}"
    
    TASK_RESPONSE=$(curl -s -X POST "$API_BASE/ai/image/task" \
        -H "Content-Type: application/json" \
        -d '{
            "prompt": "一只可爱的小猫咪",
            "user_id": "'$USER_ID'",
            "model": "doubao-seedream-3.0-t2i",
            "size": "1024x1024"
        }')
    
    TASK_ID=$(echo "$TASK_RESPONSE" | grep -o '"task_id":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$TASK_ID" ]; then
        echo -e "${GREEN}✅ 正常任务创建成功: $TASK_ID${NC}"
        NORMAL_TASK_ID="$TASK_ID"
    else
        echo -e "${RED}❌ 正常任务创建失败${NC}"
        echo "响应: $TASK_RESPONSE"
    fi
}

# 测试无效提供商任务（应该被归档）
test_invalid_provider_task() {
    echo -e "${BLUE}❌ 测试无效提供商任务（应该被归档）...${NC}"
    
    # 创建一个使用无效提供商的任务
    TASK_RESPONSE=$(curl -s -X POST "$API_BASE/ai/tasks" \
        -H "Content-Type: application/json" \
        -d '{
            "prompt": "测试无效提供商",
            "user_id": "'$USER_ID'",
            "type": "image",
            "provider": "invalid_provider",
            "model": "test-model"
        }')
    
    TASK_ID=$(echo "$TASK_RESPONSE" | grep -o '"task_id":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$TASK_ID" ]; then
        echo -e "${YELLOW}⚠️  无效提供商任务创建: $TASK_ID${NC}"
        INVALID_TASK_ID="$TASK_ID"
    else
        echo -e "${RED}❌ 无效提供商任务创建失败${NC}"
        echo "响应: $TASK_RESPONSE"
    fi
}

# 等待任务处理
wait_for_processing() {
    echo -e "${BLUE}⏳ 等待任务处理（10秒）...${NC}"
    sleep 10
}

# 检查队列状态
check_queue_status() {
    echo -e "${BLUE}📊 检查队列状态...${NC}"
    
    ./scripts/clear_redis_queue.sh --show-only
}

# 清理测试数据
cleanup() {
    echo -e "${BLUE}🧹 清理测试数据...${NC}"
    
    if [ -n "$USER_ID" ]; then
        curl -s -X DELETE "$API_BASE/users/$USER_ID" > /dev/null || true
        echo -e "${GREEN}✅ 测试用户已删除${NC}"
    fi
}

# 主函数
main() {
    echo -e "${BLUE}开始队列状态测试${NC}"
    echo ""
    
    # 设置清理陷阱
    trap cleanup EXIT
    
    # 执行测试步骤
    check_server
    echo ""
    
    create_test_user
    echo ""
    
    test_normal_task
    echo ""
    
    test_invalid_provider_task
    echo ""
    
    wait_for_processing
    echo ""
    
    check_queue_status
    echo ""
    
    echo -e "${GREEN}🎉 队列状态测试完成！${NC}"
    echo ""
    echo -e "${YELLOW}📋 预期结果:${NC}"
    echo "  - 正常任务应该在 processed 或 retry 队列中"
    echo "  - 无效提供商任务应该在 archived 队列中（不是 processed）"
    echo "  - archived 队列表示失败但跳过重试的任务"
}

# 运行主函数
main "$@" 