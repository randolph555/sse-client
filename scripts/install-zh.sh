#!/bin/bash

# SSE Client 一键安装脚本（国内加速版）
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
REPO="randolph555/sse-client"
BINARY_NAME="sse"
GITHUB_PROXY="http://gh.cdn01.cn"

# 检测操作系统和架构
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux) OS="linux" ;;
        darwin) OS="darwin" ;;
        freebsd) OS="freebsd" ;;
        *)
            echo -e "${RED}❌ 不支持的操作系统: $os${NC}"
            echo -e "${YELLOW}💡 支持的系统: Linux, macOS, FreeBSD${NC}"
            exit 1
            ;;
    esac
    
    case $arch in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *)
            echo -e "${RED}❌ 不支持的架构: $arch${NC}"
            echo -e "${YELLOW}💡 支持的架构: amd64, arm64${NC}"
            exit 1
            ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
}

# 检查依赖
check_dependencies() {
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        echo -e "${RED}❌ 需要 curl 或 wget${NC}"
        echo -e "${YELLOW}💡 请安装: sudo apt install curl 或 sudo yum install curl${NC}"
        exit 1
    fi
    
    # 选择下载工具
    if command -v curl >/dev/null 2>&1; then
        DOWNLOAD_TOOL="curl"
    else
        DOWNLOAD_TOOL="wget"
    fi
}

# 确定安装目录
determine_install_dir() {
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
        USE_SUDO=false
    else
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
        USE_SUDO=false
    fi
}

# 下载文件
download_files() {
    local temp_dir="$1"
    local binary_url="${GITHUB_PROXY}/https://raw.githubusercontent.com/${REPO}/main/dist/sse-${PLATFORM}"
    local config_url="${GITHUB_PROXY}/https://raw.githubusercontent.com/${REPO}/main/configs/config.yaml"
    
    echo -e "${BLUE}📥 下载二进制文件...${NC}"
    echo -e "${BLUE}   URL: $binary_url${NC}"
    local binary_file="$temp_dir/sse-binary"
    if [ "$DOWNLOAD_TOOL" = "curl" ]; then
        if ! curl -L -f -o "$binary_file" "$binary_url"; then
            echo -e "${RED}❌ 二进制文件下载失败${NC}"
            echo -e "${YELLOW}💡 尝试直接从GitHub下载...${NC}"
            local direct_url="https://raw.githubusercontent.com/${REPO}/main/dist/sse-${PLATFORM}"
            if ! curl -L -f -o "$binary_file" "$direct_url"; then
                echo -e "${RED}❌ 直接下载也失败${NC}"
                return 1
            fi
        fi
    else
        if ! wget -O "$binary_file" "$binary_url"; then
            echo -e "${RED}❌ 二进制文件下载失败${NC}"
            echo -e "${YELLOW}💡 尝试直接从GitHub下载...${NC}"
            local direct_url="https://raw.githubusercontent.com/${REPO}/main/dist/sse-${PLATFORM}"
            if ! wget -O "$binary_file" "$direct_url"; then
                echo -e "${RED}❌ 直接下载也失败${NC}"
                return 1
            fi
        fi
    fi
    
    echo -e "${BLUE}📥 下载配置文件...${NC}"
    if [ "$DOWNLOAD_TOOL" = "curl" ]; then
        if ! curl -L -f -o "$temp_dir/config.yaml" "$config_url"; then
            echo -e "${YELLOW}⚠️  配置文件下载失败，将使用默认配置${NC}"
        fi
    else
        if ! wget -O "$temp_dir/config.yaml" "$config_url"; then
            echo -e "${YELLOW}⚠️  配置文件下载失败，将使用默认配置${NC}"
        fi
    fi
    
    chmod +x "$binary_file"
    echo "$binary_file"
    return 0
}

# 安装文件
install_files() {
    local source_file="$1"
    local temp_dir="$2"
    local target_path="$INSTALL_DIR/$BINARY_NAME"
    local config_dir="$INSTALL_DIR/sse-configs"
    
    echo -e "${BLUE}📦 安装到 $target_path${NC}"
    
    # 安装二进制文件（强制覆盖）
    if [ "$USE_SUDO" = true ]; then
        sudo cp "$source_file" "$target_path"
        sudo chmod +x "$target_path"
    else
        cp "$source_file" "$target_path"
        chmod +x "$target_path"
    fi
    
    # 安装配置文件
    if [ -f "$temp_dir/config.yaml" ]; then
        echo -e "${BLUE}📋 安装配置文件到 $config_dir${NC}"
        if [ "$USE_SUDO" = true ]; then
            sudo mkdir -p "$config_dir"
            sudo cp "$temp_dir/config.yaml" "$config_dir/"
        else
            mkdir -p "$config_dir"
            cp "$temp_dir/config.yaml" "$config_dir/"
        fi
    fi
}

# 更新PATH
update_path() {
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        export PATH="$INSTALL_DIR:$PATH"
        
        # 添加到shell配置文件
        local shell_config=""
        case "$SHELL" in
            */bash) shell_config="$HOME/.bashrc" ;;
            */zsh) shell_config="$HOME/.zshrc" ;;
        esac
        
        if [ -n "$shell_config" ] && [ -w "$shell_config" ]; then
            if ! grep -q "export PATH.*$INSTALL_DIR" "$shell_config" 2>/dev/null; then
                echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$shell_config"
                echo -e "${GREEN}✅ 已添加到 $shell_config${NC}"
            fi
        fi
    fi
}

# 验证安装
verify_installation() {
    if command -v sse >/dev/null 2>&1; then
        echo -e "${GREEN}🎉 安装成功！${NC}"
        if sse config >/dev/null 2>&1; then
            echo -e "${GREEN}✅ 功能验证成功${NC}"
        else
            echo -e "${YELLOW}⚠️  需要配置API密钥${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  请重新打开终端或运行: source ~/.bashrc${NC}"
    fi
}

# 显示使用说明
show_usage() {
    echo -e "\n${BLUE}📖 快速开始:${NC}"
    echo -e "   # 1. 设置API密钥:"
    echo -e "   export OPENAI_API_KEY=\"your-key\""
    echo -e "   export DEEPSEEK_API_KEY=\"your-key\""
    echo -e "\n   # 2. 开始使用:"
    echo -e "   sse \"你好，请介绍一下自己\""
    echo -e "   sse -c \"查看系统状态\""
    echo -e "\n${BLUE}📚 更多信息:${NC}"
    echo -e "   sse --help"
    echo -e "   sse list"
}

# 主函数
main() {
    echo -e "${BLUE}🚀 SSE Client 一键安装（国内加速版）${NC}"
    
    detect_platform
    check_dependencies
    determine_install_dir
    
    echo -e "   系统: $OS"
    echo -e "   架构: $ARCH"
    echo -e "   安装: $INSTALL_DIR/$BINARY_NAME"
    echo ""
    
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # 检查本地构建文件
    local source_file=""
    if [ -f "./build/sse" ]; then
        source_file="./build/sse"
        echo -e "${YELLOW}🔧 使用本地构建文件${NC}"
    else
        source_file=$(download_files "$temp_dir")
        if [ $? -ne 0 ] || [ ! -f "$source_file" ]; then
            echo -e "${RED}❌ 下载失败${NC}"
            exit 1
        fi
        echo -e "${GREEN}✅ 下载完成${NC}"
    fi
    
    install_files "$source_file" "$temp_dir"
    update_path
    verify_installation
    show_usage
    
    echo -e "\n${GREEN}🚀 让 AI 成为你的终端超能力！${NC}"
}

main "$@"