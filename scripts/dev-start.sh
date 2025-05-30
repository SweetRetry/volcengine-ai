#!/bin/bash

# 开发环境启动脚本
# 用于快速启动API服务器和Worker服务的开发模式

echo "🚀 Volcengine AI Server 开发环境启动脚本"
echo "============================================"

# 检查Air是否已安装
if ! command -v air &> /dev/null; then
    echo "❌ Air工具未安装，正在安装..."
    go install github.com/cosmtrek/air@latest
    if [ $? -eq 0 ]; then
        echo "✅ Air工具安装成功"
    else
        echo "❌ Air工具安装失败，请手动安装: go install github.com/cosmtrek/air@latest"
        exit 1
    fi
else
    echo "✅ Air工具已安装"
fi

# 检查环境变量文件
if [ ! -f ".env" ]; then
    echo "⚠️  .env文件不存在，正在从env.example复制..."
    cp env.example .env
    echo "✅ 已创建.env文件，请根据需要修改配置"
fi

echo ""
echo "📋 可用的开发模式："
echo "1. 只启动API服务器"
echo "2. 只启动Worker服务"
echo "3. 显示如何同时启动两个服务"
echo "4. 退出"
echo ""

read -p "请选择 (1-4): " choice

case $choice in
    1)
        echo "🔥 启动API服务器开发模式..."
        make dev
        ;;
    2)
        echo "⚡ 启动Worker服务开发模式..."
        make dev-worker
        ;;
    3)
        echo ""
        echo "🔧 同时运行两个服务的方法："
        echo "请打开两个终端窗口，分别运行以下命令："
        echo ""
        echo "终端1 (API服务器):"
        echo "  make dev"
        echo ""
        echo "终端2 (Worker服务):"
        echo "  make dev-worker"
        echo ""
        echo "或者直接运行:"
        echo "  ./scripts/dev-start.sh"
        echo ""
        ;;
    4)
        echo "👋 退出"
        exit 0
        ;;
    *)
        echo "❌ 无效选择"
        exit 1
        ;;
esac 