#!/bin/bash

# API测试脚本
BASE_URL="http://localhost:8080"

echo "=== 即梦AI服务API测试 ==="

# 1. 健康检查
echo "1. 测试健康检查..."
curl -X GET "$BASE_URL/health" | jq '.'
echo -e "\n"

# 2. 创建用户
echo "2. 创建用户..."
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "name": "测试用户"
  }')
echo $USER_RESPONSE | jq '.'

# 提取用户ID
USER_ID=$(echo $USER_RESPONSE | jq -r '.data.id')
echo "用户ID: $USER_ID"
echo -e "\n"

# 3. 获取用户信息
echo "3. 获取用户信息..."
curl -s -X GET "$BASE_URL/api/v1/users/$USER_ID" | jq '.'
echo -e "\n"

# 4. 创建AI任务
echo "4. 创建文本生成任务..."
TASK_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/ai/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "text_generation",
    "model": "gpt-3.5-turbo",
    "input": {
      "prompt": "写一首关于春天的诗"
    },
    "user_id": "'$USER_ID'",
    "provider": "jimeng"
  }')
echo $TASK_RESPONSE | jq '.'

# 提取任务ID
TASK_ID=$(echo $TASK_RESPONSE | jq -r '.data.id')
echo "任务ID: $TASK_ID"
echo -e "\n"

# 5. 获取任务状态
echo "5. 获取任务状态..."
curl -s -X GET "$BASE_URL/api/v1/tasks/$TASK_ID" | jq '.'
echo -e "\n"

# 6. 获取用户任务列表
echo "6. 获取用户任务列表..."
curl -s -X GET "$BASE_URL/api/v1/tasks/user/$USER_ID?limit=10&offset=0" | jq '.'
echo -e "\n"

# 7. 获取队列统计
echo "7. 获取队列统计..."
curl -s -X GET "$BASE_URL/api/v1/queue/stats" | jq '.'
echo -e "\n"

# 8. 创建图像生成任务
echo "8. 创建图像生成任务..."
curl -s -X POST "$BASE_URL/api/v1/ai/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "image_generation",
    "model": "dall-e-3",
    "input": {
      "prompt": "一只可爱的小猫在花园里玩耍",
      "size": "1024x1024"
    },
    "user_id": "'$USER_ID'",
    "provider": "jimeng"
  }' | jq '.'
echo -e "\n"

# 9. 创建延迟任务
echo "9. 创建延迟任务（5秒后执行）..."
curl -s -X POST "$BASE_URL/api/v1/ai/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "translation",
    "model": "gpt-3.5-turbo",
    "input": {
      "text": "Hello, world!",
      "target_lang": "中文"
    },
    "user_id": "'$USER_ID'",
    "provider": "jimeng",
    "delay": 5
  }' | jq '.'
echo -e "\n"

echo "=== API测试完成 ===" 