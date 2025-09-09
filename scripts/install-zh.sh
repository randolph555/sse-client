#!/bin/bash

# SSE Client 一键安装脚本（国内加速版）
# SSE Client One-Click Installation Script (China Accelerated)

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
# 使用国内代理加速GitHub访问
GITHUB_PROXY="http://gh.cdn01.cn"

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
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="sse.exe"
        DOWNLOAD_URL="${GITHUB_PROXY}/https://github.com/${REPO}/releases/latest/download/sse-${PLATFORM}.zip"
    else
        DOWNLOAD_URL="${GITHUB_PROXY}/https://github.com/${REPO}/releases/latest/download/sse-${PLATFORM}.tar.gz"
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
    echo -e "${BLUE}🚀 SSE Client 一键安装（国内加速版）${NC}"
    echo -e "   系统: ${OS} (${DISTRO})"
    echo -e "   架构: ${ARCH}"
    echo -e "   平台: ${PLATFORM}"
    echo -e "   安装: ${INSTALL_DIR}/${BINARY_NAME}"
    echo -e "   代理: ${GITHUB_PROXY}"
    echo ""
    
    # 检查是否有本地构建的文件（用于测试）
    local source_file=""
    local temp_dir=$(mktemp -d)
    
    if [ -f "./build/sse" ]; then
        source_file="./build/sse"
        echo -e "${YELLOW}🔧 检测到本地构建文件，使用本地版本${NC}"
    else
        # 下载压缩包
        local archive_file="$temp_dir/sse.archive"
        echo -e "   下载: ${DOWNLOAD_URL}"
        echo -e "${BLUE}📥 正在下载...${NC}"
        
        if $DOWNLOAD_CMD "$archive_file" "$DOWNLOAD_URL"; then
            echo -e "${GREEN}✅ 下载完成${NC}"
            echo -e "${BLUE}📦 正在解压...${NC}"
            
            # 解压文件
            cd "$temp_dir"
            if [ "$OS" = "windows" ]; then
                unzip -q "$archive_file"
                # Windows下解压后的文件名
                source_file="$temp_dir/sse-${OS}-${ARCH}.exe"
            else
                tar xzf "$archive_file"
                # Unix系统下解压后的文件名 - 解压后直接在当前目录
                source_file="$temp_dir/sse-${OS}-${ARCH}"
            fi
            
            # 兼容处理：若解压产物中为旧目录名 configs，则重命名为 sse-configs
            if [ -d "$temp_dir/configs" ] && [ ! -d "$temp_dir/sse-configs" ]; then
                mv "$temp_dir/configs" "$temp_dir/sse-configs"
            fi
            
            if [ ! -f "$source_file" ]; then
                echo -e "${RED}❌ 解压失败${NC}"
                rm -rf "$temp_dir"
                exit 1
            fi
            
            chmod +x "$source_file"
            echo -e "${GREEN}✅ 解压完成${NC}"
        else
            echo -e "${RED}❌ 下载失败${NC}"
            echo -e "${YELLOW}💡 可能的原因:${NC}"
            echo -e "   1. 检查网络连接"
            echo -e "   2. 代理服务器暂时不可用"
            echo -e "   3. 发布版本不存在"
            echo -e "${YELLOW}💡 备选方案:${NC}"
            echo -e "   1. 稍后重试"
            echo -e "   2. 使用原版安装脚本（需要科学上网）"
            echo -e "   3. 手动下载并安装"
            rm -rf "$temp_dir"
            exit 1
        fi
    fi
    
    # 安装到目标目录
    echo -e "${BLUE}📦 正在安装...${NC}"
    local target_path="$INSTALL_DIR/$BINARY_NAME"
    local config_dir="$INSTALL_DIR/sse-configs"
    
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] && [ ! -w "$INSTALL_DIR" ]; then
        echo -e "${YELLOW}🔐 需要管理员权限安装到系统目录${NC}"
        sudo cp "$source_file" "$target_path"
        sudo chmod +x "$target_path"
        
        # 安装配置文件
        if [ -d "$temp_dir/sse-configs" ]; then
            echo -e "${BLUE}📋 安装配置文件...${NC}"
            sudo mkdir -p "$config_dir"
            sudo cp -r "$temp_dir/sse-configs/"* "$config_dir/"
        fi
    else
        cp "$source_file" "$target_path"
        chmod +x "$target_path"
        
        # 安装配置文件
        if [ -d "$temp_dir/sse-configs" ]; then
            echo -e "${BLUE}📋 安装配置文件...${NC}"
            mkdir -p "$config_dir"
            cp -r "$temp_dir/sse-configs/"* "$config_dir/"
        fi
    fi

    # 清理临时文件
    rm -rf "$temp_dir"

    echo -e "${GREEN}✅ SSE Client 安装成功！${NC}"
    
    # 刷新命令缓存
    if command -v hash >/dev/null 2>&1; then
        hash -r 2>/dev/null || true
    fi
    
    # 检查安装是否成功
    if command -v sse >/dev/null 2>&1; then
        echo -e "${GREEN}🎉 安装完成！命令已可用${NC}"
    else
        echo -e "${GREEN}🎉 安装完成！${NC}"
        echo -e "${YELLOW}💡 如果 'sse' 命令不可用，请尝试：${NC}"
        echo -e "   # 刷新命令缓存："
        echo -e "   hash -r"
        echo -e "   # 或重新打开终端"
        echo -e "   # 或手动执行："
        echo -e "   $target_path --help"
    fi

    # 显示使用说明
    echo -e "\n${BLUE}📖 快速开始:${NC}"
    echo -e "   # 1. 设置 API 密钥（选择一个）:"
    echo -e "   export OPENAI_API_KEY=\"your-key\""
    echo -e "   export ANTHROPIC_API_KEY=\"your-key\""
    echo -e "   export BAILIAN_API_KEY=\"your-key\""
    echo -e "   export DEEPSEEK_API_KEY=\"your-key\""
    echo -e "   export GOOGLE_API_KEY=\"your-key\""
    echo -e "\n   # 2. 测试配置:"
    echo -e "   sse config"
    echo -e "\n   # 3. 开始使用:"
    echo -e "   sse \"你好，请介绍一下自己\""
    echo -e "   sse -c \"查看系统状态\""
    echo -e "   sse \"总结文档\" -f README.md"
    echo -e "\n${BLUE}📚 更多信息:${NC}"
    echo -e "   sse --help"
    echo -e "   sse list"
    echo -e "   ${GITHUB_PROXY}/https://github.com/${REPO}"
    echo -e "\n${GREEN}🚀 让 AI 成为你的终端超能力！${NC}"
    echo -e "\n${YELLOW}💡 提示: 本脚本使用 ${GITHUB_PROXY} 代理加速下载${NC}"
}

# 主函数
main() {
    detect_platform
    check_dependencies
    get_download_cmd
    determine_install_dir
    install_sse
}

# 执行主函数
main "$@"