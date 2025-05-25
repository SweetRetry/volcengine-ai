#!/bin/bash

echo "=== 火山引擎即梦AI接口测试 ==="

# 测试健康检查
echo "1. 测试健康检查..."
curl -s http://localhost:8080/health | jq .

echo -e "\n2. 测试图像生成接口..."
# 测试图像生成
curl -X POST http://localhost:8080/api/v1/ai/image \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪在花园里玩耍",
    "model": "doubao-seedream-3.0-t2i",
    "size": "1024x1024",
    "quality": "standard"
  }' | jq .

echo -e "\n3. 测试参数验证..."
# 测试缺少必填参数
curl -X POST http://localhost:8080/api/v1/ai/image \
  -H "Content-Type: application/json" \
  -d '{}' | jq .

echo -e "\n=== 测试完成 ===" 