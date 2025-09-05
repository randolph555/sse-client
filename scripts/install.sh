#!/bin/bash

# SSE Client 一键安装脚本
# SSE Client One-Click Installation Script

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
REPO="randolph555/sse-client"
BINARY_NAME="sse"

# 检测操作系统和架构
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux)
            OS="linux"
            # 检测 Linux 发行版
            if [ -f /etc/os-release ]; then
                . /etc/os-release
                DISTRO=$ID
            else
                DISTRO="unknown"
            fi
            ;;
        darwin)
            OS="darwin"
            DISTRO="macos"
            ;;
        freebsd)
            OS="freebsd"
            DISTRO="freebsd"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            DISTRO="windows"
            ;;
        *)
            echo -e "${RED}❌ 不支持的操作系统: $os${NC}"
            echo -e "${YELLOW}💡 支持的系统: Linux, macOS, FreeBSD, Windows${NC}"
            exit 1
            ;;
    esac
    
    case $arch in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        i386|i686)
            ARCH="386"
            ;;
        armv7l)
            ARCH="arm"
            ;;
        *)
            echo -e "${RED}❌ 不支持的架构: $arch${NC}"
            echo -e "${YELLOW}💡 支持的架构: amd64, arm64, 386, arm${NC}"
            exit 1
            ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
    if [[ "$OS" == "windows" ]]; then
        BINARY_NAME="sse.exe"
        DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/sse-${PLATFORM}.zip"
    else
        DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/sse-${PLATFORM}.tar.gz"
    fi
}

# 检查依赖
check_dependencies() {
    local missing_deps=()
    
    if ! command -v curl >/dev/null 2>&1; then
        if ! command -v wget >/dev/null 2>&1; then
            missing_deps+=("curl 或 wget")
        fi
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo -e "${RED}❌ 缺少依赖: ${missing_deps[*]}${NC}"
        echo -e "${YELLOW}💡 请先安装依赖:${NC}"
        case $DISTRO in
            ubuntu|debian)
                echo -e "   sudo apt update && sudo apt install -y curl"
                ;;
            centos|rhel|fedora)
                echo -e "   sudo yum install -y curl 或 sudo dnf install -y curl"
                ;;
            arch)
                echo -e "   sudo pacman -S curl"
                ;;
            alpine)
                echo -e "   apk add curl"
                ;;
            macos)
                echo -e "   brew install curl"
                ;;
            *)
                echo -e "   请使用系统包管理器安装 curl"
                ;;
        esac
        exit 1
    fi
}

# 选择下载工具
get_download_cmd() {
    if command -v curl >/dev/null 2>&1; then
        DOWNLOAD_CMD="curl -L -f -o"
    elif command -v wget >/dev/null 2>&1; then
        DOWNLOAD_CMD="wget -O"
    else
        echo -e "${RED}❌ 未找到下载工具${NC}"
        exit 1
    fi
}

# 确定安装目录
determine_install_dir() {
    if [ "$OS" = "windows" ]; then
        # Windows: 尝试安装到用户目录
        if [ -n "$USERPROFILE" ]; then
            INSTALL_DIR="$USERPROFILE/bin"
            mkdir -p "$INSTALL_DIR"
        else
            INSTALL_DIR="."
        fi
    else
        # Unix-like: 尝试系统目录，失败则用户目录
        if [ -w "/usr/local/bin" ]; then
            INSTALL_DIR="/usr/local/bin"
        elif [ -w "$HOME/.local/bin" ]; then
            INSTALL_DIR="$HOME/.local/bin"
            mkdir -p "$INSTALL_DIR"
        else
            INSTALL_DIR="$HOME/.local/bin"
            mkdir -p "$INSTALL_DIR"
        fi
    fi
}

# 下载并安装
install_sse() {
    echo -e "${BLUE}🚀 SSE Client 一键安装${NC}"
    echo -e "   系统: ${OS} (${DISTRO})"
    echo -e "   架构: ${ARCH}"
    echo -e "   平台: ${PLATFORM}"
    echo -e "   安装: ${INSTALL_DIR}/${BINARY_NAME}"
    echo ""
    
    # 检查是否有本地构建的文件（用于测试）
    local source_file=""
    if [ -f "./sse" ]; then
        source_file="./sse"
        echo -e "${YELLOW}🔧 检测到本地构建文件，使用本地版本${NC}"
    elif [ -f "./build/sse" ]; then
        source_file="./build/sse"
        echo -e "${YELLOW}🔧 检测到本地构建文件，使用本地版本${NC}"
    else
        # 创建临时文件用于下载
        source_file=$(mktemp)
        echo -e "   下载: ${DOWNLOAD_URL}"
        echo -e "${BLUE}📥 正在下载...${NC}"
        if $DOWNLOAD_CMD "$source_file" "$DOWNLOAD_URL"; then
            echo -e "${GREEN}✅ 下载完成${NC}"
        else
            echo -e "${RED}❌ 下载失败${NC}"
            echo -e "${YELLOW}💡 可能的原因:${NC}"
            echo -e "   1. 网络连接问题"
            echo -e "   2. GitHub 访问受限"
            echo -e "   3. 发布版本不存在"
            echo -e "${YELLOW}💡 解决方案:${NC}"
            echo -e "   1. 检查网络连接"
            echo -e "   2. 使用代理或 VPN"
            echo -e "   3. 手动下载: ${DOWNLOAD_URL}"
            rm -f "$source_file"
            exit 1
        fi
    fi
    
    # 安装到目标目录
    echo -e "${BLUE}📦 正在安装...${NC}"
    local target_path="$INSTALL_DIR/$BINARY_NAME"
    
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] && [ ! -w "$INSTALL_DIR" ]; then
        echo -e "${YELLOW}🔐 需要管理员权限安装到系统目录${NC}"
        sudo cp "$source_file" "$target_path"
        sudo chmod +x "$target_path"
    else
        cp "$source_file" "$target_path"
        chmod +x "$target_path"
    fi
    
    # 如果是临时下载的文件，清理它
    if [[ "$source_file" == /tmp/* ]]; then
        rm -f "$source_file"
    fi
    
    echo -e "${GREEN}✅ SSE Client 安装成功！${NC}"
    
    # 检查 PATH
    check_path
}

# 检查 PATH 设置
check_path() {
    if [ "$OS" != "windows" ] && [ "$INSTALL_DIR" != "/usr/local/bin" ]; then
        if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
            echo -e "${YELLOW}⚠️  需要将 $INSTALL_DIR 添加到 PATH${NC}"
            echo -e "${YELLOW}💡 运行以下命令:${NC}"
            case $SHELL in
                */zsh)
                    echo -e "   echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.zshrc"
                    echo -e "   source ~/.zshrc"
                    ;;
                */bash)
                    echo -e "   echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.bashrc"
                    echo -e "   source ~/.bashrc"
                    ;;
                *)
                    echo -e "   export PATH=\"$INSTALL_DIR:\$PATH\""
                    ;;
            esac
            echo ""
        fi
    fi
}

# 显示使用说明
show_usage() {
    echo -e "${GREEN}🎉 安装完成！${NC}"
    echo -e "\n${BLUE}📖 快速开始:${NC}"
    echo -e "   ${YELLOW}# 1. 设置 API 密钥（选择一个）:${NC}"
    echo -e "   export OPENAI_API_KEY=\"your-key\""
    echo -e "   export BAILIAN_API_KEY=\"your-key\""
    echo -e "   export DEEPSEEK_API_KEY=\"your-key\""
    echo -e "   export GOOGLE_API_KEY=\"your-key\""
    echo -e ""
    echo -e "   ${YELLOW}# 2. 测试安装:${NC}"
    echo -e "   sse test"
    echo -e ""
    echo -e "   ${YELLOW}# 3. 开始使用:${NC}"
    echo -e "   sse \"你好，请介绍一下自己\""
    echo -e "   sse -c \"查看系统状态\""
    echo -e "   sse \"总结文档\" -f README.md"
    echo -e ""
    echo -e "${BLUE}📚 更多信息:${NC}"
    echo -e "   sse --help"
    echo -e "   sse list"
    echo -e "   https://github.com/${REPO}"
    echo -e ""
    echo -e "${GREEN}🚀 让 AI 成为你的终端超能力！${NC}"
}

# 主函数
main() {
    detect_platform
    check_dependencies
    get_download_cmd
    determine_install_dir
    install_sse
    show_usage
}

main "$@"
