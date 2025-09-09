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

# 检查系统兼容性
check_system_compatibility() {
    # 检查GLIBC版本（仅Linux系统）
    if [ "$OS" = "linux" ]; then
        if command -v ldd >/dev/null 2>&1; then
            local glibc_version=$(ldd --version 2>/dev/null | head -n1 | grep -o '[0-9]\+\.[0-9]\+' | head -n1)
            if [ -n "$glibc_version" ]; then
                echo -e "${BLUE}🔍 检测到 GLIBC 版本: $glibc_version${NC}"
                # 检查是否低于2.17（CentOS 7的版本）
                if [ "$(printf '%s\n' "2.17" "$glibc_version" | sort -V | head -n1)" = "2.17" ]; then
                    echo -e "${GREEN}✅ GLIBC版本兼容${NC}"
                else
                    echo -e "${YELLOW}⚠️  GLIBC版本较老，如果遇到兼容性问题，请联系开发者${NC}"
                fi
            fi
        fi
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

# 从dist目录下载预构建文件（fallback方案）
download_from_dist() {
    local temp_dir="$1"
    local binary_url="${GITHUB_PROXY}/https://raw.githubusercontent.com/${REPO}/main/dist/sse-${PLATFORM}"
    local config_url="${GITHUB_PROXY}/https://raw.githubusercontent.com/${REPO}/main/dist/sse-configs.tar.gz"
    
    if [ "$OS" = "windows" ]; then
        binary_url="${GITHUB_PROXY}/https://raw.githubusercontent.com/${REPO}/main/dist/sse-${PLATFORM}.exe"
    fi
    
    echo -e "${YELLOW}🔄 尝试从预构建文件下载（国内加速）...${NC}"
    echo -e "   二进制: ${binary_url}"
    echo -e "   配置: ${config_url}"
    
    # 下载二进制文件
    local binary_file="$temp_dir/sse-binary"
    if $DOWNLOAD_CMD "$binary_file" "$binary_url"; then
        echo -e "${GREEN}✅ 二进制文件下载完成${NC}"
        
        # 下载配置文件
        local config_file="$temp_dir/sse-configs.tar.gz"
        if $DOWNLOAD_CMD "$config_file" "$config_url"; then
            echo -e "${GREEN}✅ 配置文件下载完成${NC}"
            
            # 解压配置文件
            cd "$temp_dir"
            tar xzf "$config_file" 2>/dev/null || {
                echo -e "${YELLOW}⚠️  配置文件解压失败，将使用默认配置${NC}"
            }
            
            chmod +x "$binary_file"
            echo "$binary_file"
            return 0
        else
            echo -e "${YELLOW}⚠️  配置文件下载失败，将仅安装二进制文件${NC}"
            chmod +x "$binary_file"
            echo "$binary_file"
            return 0
        fi
    else
        echo -e "${RED}❌ 预构建文件下载失败${NC}"
        return 1
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
        # 首先尝试从GitHub Releases下载
        local archive_file="$temp_dir/sse.archive"
        echo -e "   下载: ${DOWNLOAD_URL}"
        echo -e "${BLUE}📥 正在从Releases下载（国内加速）...${NC}"
        
        if $DOWNLOAD_CMD "$archive_file" "$DOWNLOAD_URL"; then
            echo -e "${GREEN}✅ Releases下载完成${NC}"
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
                echo -e "${RED}❌ 解压失败，尝试fallback方案${NC}"
                source_file=$(download_from_dist "$temp_dir")
                if [ $? -ne 0 ] || [ ! -f "$source_file" ]; then
                    echo -e "${RED}❌ 所有下载方案都失败了${NC}"
                    rm -rf "$temp_dir"
                    exit 1
                fi
            else
                chmod +x "$source_file"
                echo -e "${GREEN}✅ 解压完成${NC}"
            fi
        else
            echo -e "${YELLOW}⚠️  Releases下载失败，尝试预构建文件...${NC}"
            source_file=$(download_from_dist "$temp_dir")
            if [ $? -ne 0 ] || [ ! -f "$source_file" ]; then
                echo -e "${RED}❌ 所有下载方案都失败了${NC}"
                echo -e "${YELLOW}💡 可能的原因:${NC}"
                echo -e "   1. 检查网络连接"
                echo -e "   2. 代理服务器暂时不可用"
                echo -e "   3. GitHub访问受限"
                echo -e "${YELLOW}💡 备选方案:${NC}"
                echo -e "   1. 稍后重试"
                echo -e "   2. 使用原版安装脚本（需要科学上网）"
                echo -e "   3. 手动下载并安装"
                rm -rf "$temp_dir"
                exit 1
            fi
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
    
    # 处理PATH问题
    local path_updated=false
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "${YELLOW}🔧 检测到 $INSTALL_DIR 不在PATH中，正在添加...${NC}"
        export PATH="$INSTALL_DIR:$PATH"
        
        # 尝试永久添加到shell配置文件
        local shell_config=""
        case "$SHELL" in
            */bash)
                if [ -f "$HOME/.bashrc" ]; then
                    shell_config="$HOME/.bashrc"
                elif [ -f "$HOME/.bash_profile" ]; then
                    shell_config="$HOME/.bash_profile"
                fi
                ;;
            */zsh)
                if [ -f "$HOME/.zshrc" ]; then
                    shell_config="$HOME/.zshrc"
                fi
                ;;
            */fish)
                if [ -d "$HOME/.config/fish" ]; then
                    shell_config="$HOME/.config/fish/config.fish"
                fi
                ;;
        esac
        
        if [ -n "$shell_config" ] && [ -w "$shell_config" ]; then
            if ! grep -q "export PATH.*$INSTALL_DIR" "$shell_config" 2>/dev/null; then
                echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$shell_config"
                echo -e "${GREEN}✅ 已添加到 $shell_config${NC}"
                path_updated=true
            fi
        fi
    fi
    
    # 刷新命令缓存
    if command -v hash >/dev/null 2>&1; then
        hash -r 2>/dev/null || true
    fi
    
    # 检查安装是否成功
    if command -v sse >/dev/null 2>&1; then
        echo -e "${GREEN}🎉 安装完成！命令已可用${NC}"
        # 验证功能
        echo -e "${BLUE}🔍 验证安装...${NC}"
        if sse config >/dev/null 2>&1; then
            echo -e "${GREEN}✅ 功能验证成功${NC}"
        else
            echo -e "${YELLOW}⚠️  命令可用但可能需要配置API密钥${NC}"
        fi
    else
        echo -e "${GREEN}🎉 安装完成！${NC}"
        echo -e "${YELLOW}💡 如果 'sse' 命令不可用，请尝试：${NC}"
        if [ "$path_updated" = true ]; then
            echo -e "   # 重新加载shell配置："
            echo -e "   source $shell_config"
            echo -e "   # 或重新打开终端"
        else
            echo -e "   # 刷新命令缓存："
            echo -e "   hash -r"
            echo -e "   # 或重新打开终端"
            echo -e "   # 或手动执行："
            echo -e "   $target_path --help"
        fi
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
    check_system_compatibility
    check_dependencies
    get_download_cmd
    determine_install_dir
    install_sse
}

# 执行主函数
main "$@"