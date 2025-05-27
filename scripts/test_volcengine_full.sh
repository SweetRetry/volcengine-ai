#!/bin/bash

# 火山引擎即梦AI全链路测试脚本
# 使用方法: ./scripts/test_volcengine_full.sh

set -e  # 遇到错误立即退出

echo "🚀 火山引擎即梦AI全链路测试"
echo "================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 服务器配置
SERVER_URL="http://localhost:8080"
API_BASE="$SERVER_URL/api/v1"

# 测试用户信息
TEST_USER_EMAIL="test_volcengine_$(date +%s)@example.com"
TEST_USER_NAME="火山引擎测试用户"

# 检查依赖
check_dependencies() {
    echo -e "${BLUE}🔍 检查依赖...${NC}"
    
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}❌ 错误: curl 未安装${NC}"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        echo -e "${YELLOW}⚠️  警告: jq 未安装，将使用grep解析JSON${NC}"
        USE_JQ=false
    else
        USE_JQ=true
    fi
}

# JSON解析函数
parse_json() {
    local json="$1"
    local key="$2"
    
    if [ "$USE_JQ" = true ]; then
        echo "$json" | jq -r ".data.$key // .$key // empty"
    else
        # 使用grep和sed作为fallback
        echo "$json" | grep -o "\"$key\":\"[^\"]*\"" | cut -d'"' -f4
    fi
}

# 检查服务器状态
check_server() {
    echo -e "${BLUE}🔍 检查服务器状态...${NC}"
    
    if ! curl -s "$SERVER_URL/health" > /dev/null; then
        echo -e "${RED}❌ 错误: 服务器未运行，请先启动服务器${NC}"
        echo "   使用命令: air 或 go run cmd/server/main.go"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 服务器运行正常${NC}"
}

# 创建测试用户
create_test_user() {
    echo -e "${BLUE}👤 创建测试用户...${NC}"
    
    USER_RESPONSE=$(curl -s -X POST "$API_BASE/users" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_USER_EMAIL\",
            \"name\": \"$TEST_USER_NAME\"
        }")
    
    echo "用户创建响应: $USER_RESPONSE"
    
    if echo "$USER_RESPONSE" | grep -q '"success":true'; then
        USER_ID=$(parse_json "$USER_RESPONSE" "id")
        if [ -z "$USER_ID" ]; then
            # 尝试其他可能的字段名
            USER_ID=$(echo "$USER_RESPONSE" | grep -o '"_id":"[^"]*"' | cut -d'"' -f4)
        fi
        echo -e "${GREEN}✅ 用户创建成功，用户ID: $USER_ID${NC}"
    else
        echo -e "${RED}❌ 用户创建失败${NC}"
        exit 1
    fi
}

# 创建图像生成任务
create_image_task() {
    echo -e "${BLUE}🎨 创建图像生成任务...${NC}"
    
    # 测试提示词列表
    local prompts=(
        "一只可爱的小猫咪在花园里玩耍，阳光明媚，花朵盛开，高质量，4K"
        "未来科技城市，霓虹灯闪烁，赛博朋克风格，夜景"
        "中国古典山水画，水墨风格，远山如黛，云雾缭绕"
        "宇宙中的星云，绚烂色彩，深空摄影风格"
    )
    
    # 随机选择一个提示词
    local random_index=$((RANDOM % ${#prompts[@]}))
    local selected_prompt="${prompts[$random_index]}"
    
    echo -e "${YELLOW}📝 使用提示词: $selected_prompt${NC}"
    
    TASK_RESPONSE=$(curl -s -X POST "$API_BASE/ai/image/task" \
        -H "Content-Type: application/json" \
        -d "{
            \"prompt\": \"$selected_prompt\",
            \"user_id\": \"$USER_ID\",
            \"provider\": \"volcengine_jimeng\",
            \"model\": \"doubao-seedream-3.0-t2i\",
            \"size\": \"512x512\",
            \"quality\": \"standard\"
        }")
    
    echo "任务创建响应: $TASK_RESPONSE"
    
    if echo "$TASK_RESPONSE" | grep -q '"success":true'; then
        TASK_ID=$(parse_json "$TASK_RESPONSE" "task_id")
        if [ -z "$TASK_ID" ]; then
            # 尝试从data字段中获取
            TASK_ID=$(echo "$TASK_RESPONSE" | grep -o '"task_id":"[^"]*"' | cut -d'"' -f4)
        fi
        echo -e "${GREEN}✅ 任务创建成功，任务ID: $TASK_ID${NC}"
    else
        echo -e "${RED}❌ 任务创建失败${NC}"
        echo "响应内容: $TASK_RESPONSE"
        exit 1
    fi
}

# 轮询查询任务结果
poll_task_result() {
    echo -e "${BLUE}🔄 开始轮询查询任务结果...${NC}"
    
    local max_attempts=60  # 最多轮询60次
    local attempt=0
    local sleep_interval=5  # 每次间隔5秒
    
    while [ $attempt -lt $max_attempts ]; do
        attempt=$((attempt + 1))
        
        echo -e "${YELLOW}⏳ 第 $attempt 次查询 (最多 $max_attempts 次)...${NC}"
        
        RESULT=$(curl -s -X GET "$API_BASE/ai/image/result/$TASK_ID")
        STATUS=$(parse_json "$RESULT" "status")
        
        if [ -z "$STATUS" ]; then
            # 尝试从data字段中获取状态
            STATUS=$(echo "$RESULT" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
        fi
        
        echo "当前状态: $STATUS"
        
        case "$STATUS" in
            "completed")
                IMAGE_URL=$(parse_json "$RESULT" "image_url")
                if [ -z "$IMAGE_URL" ]; then
                    # 尝试从不同路径获取图片URL
                    IMAGE_URL=$(echo "$RESULT" | grep -o '"image_url":"[^"]*"' | cut -d'"' -f4)
                fi
                
                echo ""
                echo -e "${GREEN}🎉 图像生成完成!${NC}"
                echo -e "${GREEN}🖼️  图像URL: $IMAGE_URL${NC}"
                echo ""
                echo "完整响应:"
                if [ "$USE_JQ" = true ]; then
                    echo "$RESULT" | jq '.'
                else
                    echo "$RESULT"
                fi
                return 0
                ;;
            "failed")
                local error_msg=$(parse_json "$RESULT" "message")
                if [ -z "$error_msg" ]; then
                    error_msg=$(echo "$RESULT" | grep -o '"message":"[^"]*"' | cut -d'"' -f4)
                fi
                
                echo ""
                echo -e "${RED}❌ 任务失败: $error_msg${NC}"
                echo ""
                echo "完整响应:"
                if [ "$USE_JQ" = true ]; then
                    echo "$RESULT" | jq '.'
                else
                    echo "$RESULT"
                fi
                return 1
                ;;
            "processing"|"pending"|*)
                echo -e "${YELLOW}⏳ 任务处理中，等待 $sleep_interval 秒后重试...${NC}"
                sleep $sleep_interval
                ;;
        esac
    done
    
    echo ""
    echo -e "${RED}⏰ 任务查询超时（超过 $((max_attempts * sleep_interval)) 秒）${NC}"
    return 1
}

# 获取用户任务列表
get_user_tasks() {
    echo -e "${BLUE}📋 获取用户任务列表...${NC}"
    
    TASKS_RESPONSE=$(curl -s -X GET "$API_BASE/ai/image/tasks?user_id=$USER_ID&limit=10")
    
    echo "用户任务列表:"
    if [ "$USE_JQ" = true ]; then
        echo "$TASKS_RESPONSE" | jq '.'
    else
        echo "$TASKS_RESPONSE"
    fi
}

# 清理测试数据
cleanup() {
    echo -e "${BLUE}🧹 清理测试数据...${NC}"
    
    if [ -n "$TASK_ID" ]; then
        echo "删除测试任务: $TASK_ID"
        curl -s -X DELETE "$API_BASE/ai/image/task/$TASK_ID" > /dev/null || true
    fi
    
    if [ -n "$USER_ID" ]; then
        echo "删除测试用户: $USER_ID"
        curl -s -X DELETE "$API_BASE/users/$USER_ID" > /dev/null || true
    fi
    
    echo -e "${GREEN}✅ 清理完成${NC}"
}

# 主函数
main() {
    echo -e "${BLUE}开始时间: $(date)${NC}"
    echo ""
    
    # 设置错误处理
    trap cleanup EXIT
    
    # 执行测试步骤
    check_dependencies
    check_server
    create_test_user
    create_image_task
    
    # 记录开始时间
    start_time=$(date +%s)
    
    # 轮询任务结果
    if poll_task_result; then
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        echo -e "${GREEN}✅ 全链路测试成功完成！${NC}"
        echo -e "${GREEN}⏱️  总耗时: ${duration} 秒${NC}"
        
        # 获取用户任务列表
        get_user_tasks
    else
        echo -e "${RED}❌ 全链路测试失败${NC}"
        exit 1
    fi
    
    echo ""
    echo -e "${BLUE}结束时间: $(date)${NC}"
    echo -e "${GREEN}🏁 测试完成${NC}"
}

# 运行主函数
main "$@" 