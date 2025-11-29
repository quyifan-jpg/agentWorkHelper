#!/bin/bash

# 代码生成脚本：goctl-gin + swagger
# 使用方法: ./scripts/gen.sh

set -e  # 遇到错误立即退出

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 项目根目录（脚本所在目录的上一级）
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

# 检测 goctl-gin 命令
if ! command -v goctl-gin &> /dev/null; then
    echo -e "${RED}❌ 未找到 goctl-gin，请先安装：${NC}"
    echo -e "${YELLOW}   go install github.com/zeromicro/goctl-gin@latest${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 使用 goctl-gin${NC}"
echo -e "${GREEN}🔄 开始代码生成...${NC}\n"

# 1. 验证 .api 文件
echo -e "${YELLOW}📝 步骤 1/4: 验证 API 文件...${NC}"
if ! goctl-gin api validate --api doc/api.api 2>/dev/null; then
    echo -e "${RED}❌ API 文件验证失败，请检查语法${NC}"
    exit 1
fi
echo -e "${GREEN}✅ API 文件验证通过${NC}\n"

# 2. 格式化 .api 文件
echo -e "${YELLOW}✨ 步骤 2/4: 格式化 API 文件...${NC}"
if goctl-gin api format --dir doc 2>/dev/null; then
    echo -e "${GREEN}✅ API 文件格式化完成${NC}\n"
else
    echo -e "${YELLOW}⚠️  API 文件格式化跳过（可能已是最新格式）${NC}\n"
fi

# 3. 生成项目代码（使用 goctl-gin，生成 Gin 框架代码）
# 根据 goctl-gin 用法：goctl-gin api go -api ./doc/api.api -dir ./
echo -e "${YELLOW}🔨 步骤 3/4: 生成项目代码（goctl-gin）...${NC}"
if goctl-gin api go --api ./doc/api.api --dir ./ 2>/dev/null; then
    echo -e "${GREEN}✅ 项目代码生成完成${NC}\n"
else
    echo -e "${RED}❌ 项目代码生成失败${NC}"
    exit 1
fi

# 4. 生成 Swagger UI 文档（从代码注释）
echo -e "${YELLOW}📖 步骤 4/4: 生成 Swagger UI 文档（从代码注释）...${NC}"
if ! command -v swag &> /dev/null; then
    echo -e "${YELLOW}⚠️  swag 工具未安装，正在安装...${NC}"
    go install github.com/swaggo/swag/cmd/swag@latest
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ swag 安装失败，请手动安装: go install github.com/swaggo/swag/cmd/swag@latest${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ swag 安装完成${NC}"
fi

if swag init -g cmd/api/main.go -o docs 2>/dev/null; then
    echo -e "${GREEN}✅ Swagger UI 文档生成完成${NC}\n"
else
    echo -e "${RED}❌ Swagger UI 文档生成失败，请检查代码注释${NC}"
    exit 1
fi

# 完成
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✅ 代码生成完成！${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}📍 Swagger UI: http://localhost:8889/swagger/index.html${NC}"
echo -e "${YELLOW}💡 提示: 运行 ./start.sh 或 go run cmd/api/main.go 启动服务${NC}\n"

