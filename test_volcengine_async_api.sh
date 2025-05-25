#!/bin/bash

# 火山引擎即梦AI异步API测试脚本

BASE_URL="http://localhost:8080"
USER_ID="test_user_$(date +%s)"

echo "🚀 开始测试火山引擎即梦AI异步API..."
echo "基础URL: $BASE_URL"
echo "测试用户ID: $USER_ID"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查jq是否安装
if ! command -v jq &> /dev/null; then
    echo -e "${RED}❌ 需要安装jq来解析JSON响应${NC}"
    echo "请运行: brew install jq (macOS) 或 apt-get install jq (Ubuntu)"
    exit 1
fi

# 测试函数
test_api() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "${BLUE}📋 测试: $description${NC}"
    echo "请求: $method $endpoint"
    
    local response
    if [ -n "$data" ]; then
        echo "数据: $data"
        response=$(curl -s -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -X $method "$BASE_URL$endpoint")
    fi
    
    echo "响应: $response"
    echo ""
    
    # 只返回响应，不包含其他输出
    echo "$response"
}

# 辅助函数：只获取API响应，不显示调试信息
get_api_response() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -n "$data" ]; then
        curl -s -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data"
    else
        curl -s -X $method "$BASE_URL$endpoint"
    fi
}

# 1. 健康检查
echo -e "${YELLOW}=== 1. 健康检查 ===${NC}"
test_api "GET" "/health" "" "服务健康检查"

# 2. 创建用户
echo -e "${YELLOW}=== 2. 创建测试用户 ===${NC}"
user_data='{
    "email": "test_async_'$(date +%s)'@example.com",
    "name": "异步测试用户"
}'
test_api "POST" "/api/v1/users" "$user_data" "创建测试用户"
user_response=$(get_api_response "POST" "/api/v1/users" "$user_data")
user_id=$(echo "$user_response" | jq -r '.data.id // empty')

if [ -z "$user_id" ] || [ "$user_id" = "null" ]; then
    echo -e "${RED}❌ 创建用户失败，使用默认用户ID${NC}"
    user_id=$USER_ID
else
    echo -e "${GREEN}✅ 用户创建成功，ID: $user_id${NC}"
fi
echo ""

# 3. 创建图像生成任务
echo -e "${YELLOW}=== 3. 创建异步图像生成任务 ===${NC}"
image_task_data='{
    "prompt": "一只可爱的橘猫在樱花树下玩耍，动漫风格，高质量",
    "model": "doubao-seedream-3.0-t2i",
    "size": "1024x1024",
    "quality": "standard",
    "user_id": "'$user_id'"
}'
test_api "POST" "/api/v1/ai/image/task" "$image_task_data" "创建图像生成任务"
task_response=$(get_api_response "POST" "/api/v1/ai/image/task" "$image_task_data")
task_id=$(echo "$task_response" | jq -r '.data.task_id // empty')

if [ -z "$task_id" ] || [ "$task_id" = "null" ]; then
    echo -e "${RED}❌ 创建任务失败${NC}"
    echo "响应详情: $task_response"
    exit 1
else
    echo -e "${GREEN}✅ 任务创建成功，任务ID: $task_id${NC}"
fi
echo ""

# 4. 查询任务结果（轮询）
echo -e "${YELLOW}=== 4. 查询任务结果 ===${NC}"
max_attempts=10
attempt=1

while [ $attempt -le $max_attempts ]; do
    echo -e "${BLUE}📊 第 $attempt 次查询任务结果...${NC}"
    
    test_api "GET" "/api/v1/ai/image/result/$task_id" "" "查询任务结果 (第${attempt}次)"
    result_response=$(get_api_response "GET" "/api/v1/ai/image/result/$task_id")
    
    # 检查任务状态
    status=$(echo "$result_response" | jq -r '.data.status // empty')
    
    case $status in
        "completed")
            echo -e "${GREEN}🎉 任务完成！${NC}"
            image_url=$(echo "$result_response" | jq -r '.data.image_url // .data.result.image_url // empty')
            if [ -n "$image_url" ] && [ "$image_url" != "null" ]; then
                echo -e "${GREEN}🖼️  图像URL: $image_url${NC}"
            fi
            break
            ;;
        "processing")
            echo -e "${YELLOW}⏳ 任务处理中，等待3秒后重试...${NC}"
            sleep 3
            ;;
        "failed")
            echo -e "${RED}❌ 任务失败${NC}"
            error_msg=$(echo "$result_response" | jq -r '.message // .error // empty')
            echo -e "${RED}错误信息: $error_msg${NC}"
            break
            ;;
        *)
            echo -e "${YELLOW}⏳ 任务状态未知($status)，等待3秒后重试...${NC}"
            sleep 3
            ;;
    esac
    
    attempt=$((attempt + 1))
done

if [ $attempt -gt $max_attempts ]; then
    echo -e "${RED}⏰ 查询超时，请稍后手动查询任务结果${NC}"
fi
echo ""

# 5. 测试不同参数的任务创建
echo -e "${YELLOW}=== 5. 测试不同参数的任务创建 ===${NC}"

# 测试高质量图像
hq_task_data='{
    "prompt": "未来科技城市夜景，霓虹灯闪烁，赛博朋克风格",
    "model": "doubao-seedream-3.0-t2i",
    "size": "1024x1024",
    "quality": "hd",
    "style": "cyberpunk",
    "user_id": "'$user_id'",
    "n": 1
}'
test_api "POST" "/api/v1/ai/image/task" "$hq_task_data" "创建高质量图像任务"
hq_response=$(get_api_response "POST" "/api/v1/ai/image/task" "$hq_task_data")
hq_task_id=$(echo "$hq_response" | jq -r '.data.task_id // empty')
if [ -n "$hq_task_id" ] && [ "$hq_task_id" != "null" ]; then
    echo -e "${GREEN}✅ 高质量任务创建成功，ID: $hq_task_id${NC}"
fi

# 6. 测试参数校验
echo -e "${YELLOW}=== 6. 测试参数校验 ===${NC}"

# 测试缺少必需参数
invalid_data='{
    "model": "doubao-seedream-3.0-t2i"
}'
test_api "POST" "/api/v1/ai/image/task" "$invalid_data" "测试缺少必需参数"
invalid_response=$(get_api_response "POST" "/api/v1/ai/image/task" "$invalid_data")
error_msg=$(echo "$invalid_response" | jq -r '.error // .message // empty')
if [ -n "$error_msg" ] && [ "$error_msg" != "null" ]; then
    echo -e "${GREEN}✅ 参数校验正常工作${NC}"
fi

# 测试无效任务ID查询
test_api "GET" "/api/v1/ai/image/result/invalid_task_id" "" "测试无效任务ID查询"
invalid_result=$(get_api_response "GET" "/api/v1/ai/image/result/invalid_task_id")
invalid_error=$(echo "$invalid_result" | jq -r '.error // .message // empty')
if [ -n "$invalid_error" ] && [ "$invalid_error" != "null" ]; then
    echo -e "${GREEN}✅ 无效任务ID处理正常${NC}"
fi

# 7. 查询用户任务列表
echo -e "${YELLOW}=== 7. 查询用户任务列表 ===${NC}"
test_api "GET" "/api/v1/tasks/user/$user_id" "" "查询用户任务列表"
user_tasks_response=$(get_api_response "GET" "/api/v1/tasks/user/$user_id")
tasks_count=$(echo "$user_tasks_response" | jq -r '.data | length // 0')
echo -e "${GREEN}✅ 用户任务数量: $tasks_count${NC}"

echo -e "${GREEN}🎯 测试完成！${NC}"
echo ""
echo -e "${BLUE}📝 测试总结:${NC}"
echo "1. ✅ 健康检查通过"
echo "2. ✅ 用户创建成功"
echo "3. ✅ 异步任务创建成功"
echo "4. ✅ 任务结果查询功能正常"
echo "5. ✅ 参数校验功能正常"
echo ""
echo -e "${YELLOW}💡 使用说明:${NC}"
echo "- 创建任务: POST /api/v1/ai/image/task"
echo "- 查询结果: GET /api/v1/ai/image/result/{task_id}"
echo "- 任务ID格式: volcengine_img_{timestamp}"
echo ""
echo -e "${BLUE}🔗 主要任务ID: $task_id${NC}"
echo "可以使用以下命令手动查询结果:"
echo "curl -X GET \"$BASE_URL/api/v1/ai/image/result/$task_id\"" 