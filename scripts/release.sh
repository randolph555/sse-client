#!/bin/bash

# SSE Client 发布脚本
# 自动创建 tag 并触发 GitHub Actions 构建发布

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 获取当前版本
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo -e "${BLUE}📋 当前版本: ${CURRENT_VERSION}${NC}"

# 提示输入新版本
echo -e "${YELLOW}请输入新版本号 (格式: v1.0.0):${NC}"
read -r NEW_VERSION

# 验证版本格式
if [[ ! $NEW_VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}❌ 版本格式错误！请使用 v1.0.0 格式${NC}"
    exit 1
fi

# 检查是否有未提交的更改
if [[ -n $(git status --porcelain) ]]; then
    echo -e "${YELLOW}⚠️  检测到未提交的更改：${NC}"
    git status --short
    echo -e "${YELLOW}是否继续？(y/N):${NC}"
    read -r CONTINUE
    if [[ $CONTINUE != "y" && $CONTINUE != "Y" ]]; then
        echo -e "${RED}❌ 发布已取消${NC}"
        exit 1
    fi
fi

# 运行测试
echo -e "${BLUE}🧪 运行测试...${NC}"
if ! go test ./...; then
    echo -e "${RED}❌ 测试失败！请修复后重试${NC}"
    exit 1
fi

# 构建检查
echo -e "${BLUE}🔨 构建检查...${NC}"
if ! make build; then
    echo -e "${RED}❌ 构建失败！请修复后重试${NC}"
    exit 1
fi

# 提交所有更改（如果有）
if [[ -n $(git status --porcelain) ]]; then
    echo -e "${BLUE}📝 提交更改...${NC}"
    git add .
    git commit -m "chore: prepare for release ${NEW_VERSION}"
fi

# 创建并推送 tag
echo -e "${BLUE}🏷️  创建 tag: ${NEW_VERSION}${NC}"
git tag -a "${NEW_VERSION}" -m "Release ${NEW_VERSION}"

echo -e "${BLUE}📤 推送到远程仓库...${NC}"
git push origin main
git push origin "${NEW_VERSION}"

echo -e "${GREEN}✅ 发布流程已启动！${NC}"
echo -e "${GREEN}🚀 GitHub Actions 将自动构建并发布 ${NEW_VERSION}${NC}"
echo -e "${BLUE}📋 查看构建状态: https://github.com/randolph555/sse-client/actions${NC}"
echo -e "${BLUE}📦 发布页面: https://github.com/randolph555/sse-client/releases${NC}"

echo -e "\n${YELLOW}📝 发布后续步骤：${NC}"
echo -e "1. 等待 GitHub Actions 构建完成（约 2-5 分钟）"
echo -e "2. 检查 Releases 页面确认文件已上传"
echo -e "3. 测试安装脚本是否正常工作"
echo -e "4. 更新文档（如需要）"