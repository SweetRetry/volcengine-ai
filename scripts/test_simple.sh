#!/bin/bash

# 简单的API测试脚本
echo "🔍 简单API测试"
echo "==============="

SERVER_URL="http://localhost:8080"

# 1. 测试健康检查
echo "1. 测试健康检查..."
curl -s "$SERVER_URL/health" | jq '.' || echo "健康检查失败"

echo ""

# 2. 创建用户
echo "2. 创建测试用户..."
USER_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "simple_test@example.com",
    "name": "简单测试用户"
  }')

echo "用户创建响应:"
echo "$USER_RESPONSE" | jq '.' || echo "$USER_RESPONSE"

# 提取用户ID
USER_ID=$(echo "$USER_RESPONSE" | jq -r '.data.id' 2>/dev/null)
echo "用户ID: $USER_ID"

echo ""

# 3. 创建图像任务
echo "3. 创建图像生成任务..."
TASK_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/ai/image/task" \
  -H "Content-Type: application/json" \
  -d "{
    \"prompt\": \"一只可爱的小猫\",
    \"user_id\": \"$USER_ID\",
    \"provider\": \"volcengine\",
    \"model\": \"doubao-seedream-3.0-t2i\",
    \"size\": \"512x512\"
  }")

echo "任务创建响应:"
echo "$TASK_RESPONSE" | jq '.' || echo "$TASK_RESPONSE"

# 提取任务ID
TASK_ID=$(echo "$TASK_RESPONSE" | jq -r '.data.task_id' 2>/dev/null)
echo "任务ID: $TASK_ID"

echo ""

# 4. 查询任务状态
echo "4. 查询任务状态..."
RESULT_RESPONSE=$(curl -s -X GET "$SERVER_URL/api/v1/ai/image/result/$TASK_ID")

echo "任务状态响应:"
echo "$RESULT_RESPONSE" | jq '.' || echo "$RESULT_RESPONSE"

echo ""

# # 5. 清理
# echo "5. 清理测试数据..."
# if [ -n "$TASK_ID" ] && [ "$TASK_ID" != "null" ]; then
#     curl -s -X DELETE "$SERVER_URL/api/v1/ai/image/task/$TASK_ID" > /dev/null
#     echo "已删除任务: $TASK_ID"
# fi

# if [ -n "$USER_ID" ] && [ "$USER_ID" != "null" ]; then
#     curl -s -X DELETE "$SERVER_URL/api/v1/users/$USER_ID" > /dev/null
#     echo "已删除用户: $USER_ID"
# fi

echo "测试完成" 