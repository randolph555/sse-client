#!/bin/bash

# SSE Client 一键安装脚本
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

# 检测操作系统和架构
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux) OS="linux" ;;
        darwin) OS="darwin" ;;
        freebsd) OS="freebsd" ;;
        *)
            echo -e "${RED}❌ Unsupported OS: $os${NC}"
            echo -e "${YELLOW}💡 Supported: Linux, macOS, FreeBSD${NC}"
            exit 1
            ;;
    esac
    
    case $arch in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *)
            echo -e "${RED}❌ Unsupported architecture: $arch${NC}"
            echo -e "${YELLOW}💡 Supported: amd64, arm64${NC}"
            exit 1
            ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
}

# 检查依赖
check_dependencies() {
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        echo -e "${RED}❌ curl or wget required${NC}"
        echo -e "${YELLOW}💡 Install: sudo apt install curl or sudo yum install curl${NC}"
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
    local binary_url="https://raw.githubusercontent.com/${REPO}/main/dist/sse-${PLATFORM}"
    local config_url="https://raw.githubusercontent.com/${REPO}/main/configs/config.yaml"
    
    echo -e "${BLUE}📥 Downloading binary...${NC}"
    echo -e "${BLUE}   URL: $binary_url${NC}"
    local binary_file="$temp_dir/sse-binary"
    if [ "$DOWNLOAD_TOOL" = "curl" ]; then
        if ! curl -L -f -o "$binary_file" "$binary_url"; then
            echo -e "${RED}❌ Binary download failed${NC}"
            return 1
        fi
    else
        if ! wget -O "$binary_file" "$binary_url"; then
            echo -e "${RED}❌ Binary download failed${NC}"
            return 1
        fi
    fi
    
    echo -e "${BLUE}📥 Downloading config file...${NC}"
    if [ "$DOWNLOAD_TOOL" = "curl" ]; then
        if ! curl -L -f -o "$temp_dir/config.yaml" "$config_url"; then
            echo -e "${YELLOW}⚠️  Config file download failed, using defaults${NC}"
        fi
    else
        if ! wget -O "$temp_dir/config.yaml" "$config_url"; then
            echo -e "${YELLOW}⚠️  Config file download failed, using defaults${NC}"
        fi
    fi
    
    chmod +x "$binary_file"
    return 0
}

# 安装文件
install_files() {
    local source_file="$1"
    local temp_dir="$2"
    local target_path="$INSTALL_DIR/$BINARY_NAME"
    local config_dir="$INSTALL_DIR/sse-configs"
    
    echo -e "${BLUE}📦 Installing to $target_path${NC}"
    
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
        echo -e "${BLUE}📋 Installing config file to $config_dir${NC}"
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
                echo -e "${GREEN}✅ Added to $shell_config${NC}"
            fi
        fi
    fi
}

# 验证安装
verify_installation() {
    if command -v sse >/dev/null 2>&1; then
        echo -e "${GREEN}🎉 Installation successful!${NC}"
        if sse config >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Functionality verified${NC}"
        else
            echo -e "${YELLOW}⚠️  API key configuration needed${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Please restart terminal or run: source ~/.bashrc${NC}"
    fi
}

# 显示使用说明
show_usage() {
    echo -e "\n${BLUE}📖 Quick Start:${NC}"
    echo -e "   # 1. Set API keys:"
    echo -e "   export OPENAI_API_KEY=\"your-key\""
    echo -e "   export DEEPSEEK_API_KEY=\"your-key\""
    echo -e "\n   # 2. Start using:"
    echo -e "   sse \"Hello, introduce yourself\""
    echo -e "   sse -c \"check system status\""
    echo -e "\n${BLUE}📚 More info:${NC}"
    echo -e "   sse --help"
    echo -e "   sse list"
}

# 主函数
main() {
    echo -e "${BLUE}🚀 SSE Client One-Click Installer${NC}"
    
    detect_platform
    check_dependencies
    determine_install_dir
    
    echo -e "   OS: $OS"
    echo -e "   Arch: $ARCH"
    echo -e "   Install: $INSTALL_DIR/$BINARY_NAME"
    echo ""
    
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # 检查本地构建文件
    local source_file=""
    if [ -f "./build/sse" ]; then
        source_file="./build/sse"
        echo -e "${YELLOW}🔧 Using local build${NC}"
    else
        if download_files "$temp_dir"; then
            source_file="$temp_dir/sse-binary"
            echo -e "${GREEN}✅ Download complete${NC}"
        else
            echo -e "${RED}❌ Download failed${NC}"
            exit 1
        fi
    fi
    
    install_files "$source_file" "$temp_dir"
    update_path
    verify_installation
    show_usage
    
    echo -e "\n${GREEN}🚀 Make AI your terminal superpower!${NC}"
}

main "$@"