#!/bin/bash

# AIWorkHelper Backend 启动脚本

echo "🚀 启动 AIWorkHelper Backend..."

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "❌ 错误: 请在 BackEnd 目录下运行此脚本"
    exit 1
fi

# 检查配置文件
if [ ! -f "etc/backend.yaml" ]; then
    echo "❌ 错误: 配置文件 etc/backend.yaml 不存在"
    exit 1
fi

# 检查是否需要生成 Swagger 文档
if [ ! -f "docs/docs.go" ]; then
    echo "📝 生成 Swagger 文档..."
    if ! command -v swag &> /dev/null; then
        echo "⚠️  swag 工具未安装，正在安装..."
        go install github.com/swaggo/swag/cmd/swag@latest
    fi
    swag init -g cmd/api/main.go
fi

# 启动服务
echo "✅ 启动服务..."
echo "📍 Swagger UI: http://localhost:8889/swagger/index.html"
echo "📍 API 地址: http://localhost:8889"
echo ""
go run cmd/api/main.go -f etc/backend.yaml

