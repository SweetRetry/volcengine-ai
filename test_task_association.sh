#!/bin/bash

# 测试任务ID关联功能
# 验证数据库任务ID和火山引擎外部任务ID是否正确关联

BASE_URL="http://localhost:8080/api/v1"

echo "=== 测试火山引擎即梦AI任务ID关联功能 ==="
echo

# 1. 创建图像生成任务
echo "1. 创建图像生成任务..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/ai/image/task" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪，坐在花园里，动漫风格",
    "user_id": "test-user-123",
    "model": "doubao-seedream-3.0-t2i",
    "size": "1024x1024",
    "quality": "hd",
    "style": "anime"
  }')

echo "创建任务响应:"
echo "$CREATE_RESPONSE" | jq '.'
echo

# 提取任务ID和外部任务ID
TASK_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.task_id')
EXTERNAL_TASK_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.external_task_id')

if [ "$TASK_ID" = "null" ] || [ "$EXTERNAL_TASK_ID" = "null" ]; then
    echo "❌ 创建任务失败，无法获取任务ID"
    exit 1
fi

echo "✅ 任务创建成功"
echo "   数据库任务ID: $TASK_ID"
echo "   外部任务ID: $EXTERNAL_TASK_ID"
echo

# 2. 查询任务状态（第一次）
echo "2. 查询任务状态（第一次）..."
RESULT_RESPONSE_1=$(curl -s "$BASE_URL/ai/image/result/$TASK_ID")
echo "任务状态响应:"
echo "$RESULT_RESPONSE_1" | jq '.'
echo

STATUS_1=$(echo "$RESULT_RESPONSE_1" | jq -r '.data.status')
echo "当前状态: $STATUS_1"
echo

# 3. 等待一段时间后再次查询
echo "3. 等待3秒后再次查询任务状态..."
sleep 3

RESULT_RESPONSE_2=$(curl -s "$BASE_URL/ai/image/result/$TASK_ID")
echo "任务状态响应:"
echo "$RESULT_RESPONSE_2" | jq '.'
echo

STATUS_2=$(echo "$RESULT_RESPONSE_2" | jq -r '.data.status')
echo "当前状态: $STATUS_2"

# 检查是否有图像URL
IMAGE_URL=$(echo "$RESULT_RESPONSE_2" | jq -r '.data.image_url // empty')
if [ -n "$IMAGE_URL" ] && [ "$IMAGE_URL" != "null" ]; then
    echo "✅ 任务完成，图像URL: $IMAGE_URL"
else
    echo "⏳ 任务仍在处理中或失败"
fi
echo

# 4. 获取用户任务列表
echo "4. 获取用户任务列表..."
TASKS_RESPONSE=$(curl -s "$BASE_URL/ai/image/tasks?user_id=test-user-123&limit=5")
echo "用户任务列表:"
echo "$TASKS_RESPONSE" | jq '.'
echo

# 5. 验证任务ID关联
echo "5. 验证任务ID关联..."
FOUND_TASK=$(echo "$TASKS_RESPONSE" | jq -r ".data.tasks[] | select(.task_id == \"$TASK_ID\")")
if [ -n "$FOUND_TASK" ]; then
    echo "✅ 在用户任务列表中找到了创建的任务"
    echo "任务详情:"
    echo "$FOUND_TASK" | jq '.'
else
    echo "❌ 在用户任务列表中未找到创建的任务"
fi
echo

# 6. 测试多次查询的一致性
echo "6. 测试多次查询的一致性..."
for i in {1..3}; do
    echo "第 $i 次查询:"
    CONSISTENCY_RESPONSE=$(curl -s "$BASE_URL/ai/image/result/$TASK_ID")
    CONSISTENCY_STATUS=$(echo "$CONSISTENCY_RESPONSE" | jq -r '.data.status')
    echo "  状态: $CONSISTENCY_STATUS"
    
    if [ "$i" -gt 1 ] && [ "$CONSISTENCY_STATUS" != "$PREV_STATUS" ]; then
        echo "  ⚠️  状态发生变化: $PREV_STATUS -> $CONSISTENCY_STATUS"
    fi
    PREV_STATUS="$CONSISTENCY_STATUS"
    sleep 1
done
echo

echo "=== 测试完成 ==="
echo "总结:"
echo "- 数据库任务ID: $TASK_ID"
echo "- 外部任务ID: $EXTERNAL_TASK_ID"
echo "- 最终状态: $CONSISTENCY_STATUS"

if [ "$CONSISTENCY_STATUS" = "completed" ]; then
    echo "✅ 任务成功完成"
elif [ "$CONSISTENCY_STATUS" = "processing" ] || [ "$CONSISTENCY_STATUS" = "pending" ]; then
    echo "⏳ 任务仍在处理中"
else
    echo "❌ 任务失败或状态异常"
fi 