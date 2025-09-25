#!/bin/bash

echo "🚀 启动 TodoList 应用..."

# 进入项目目录
cd "$(dirname "$0")"

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go"
    exit 1
fi

# 安装依赖
echo "📦 安装依赖..."
go mod tidy

# 编译应用
echo "🔨 编译应用..."
go build -o todolist-server ./cmd/main.go

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

# 启动服务器
echo "🌟 启动服务器..."
echo "📱 在浏览器中打开: http://localhost:8081"
echo "⏹️  按 Ctrl+C 停止服务器"
echo ""

PORT=8081 ./todolist-server