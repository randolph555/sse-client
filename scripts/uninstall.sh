#!/bin/bash

# SSE Client 卸载脚本
# SSE Client Uninstall Script

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检测操作系统
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    
    case $os in
        linux|darwin|freebsd)
            OS="unix"
            BINARY_NAME="sse"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            BINARY_NAME="sse.exe"
            ;;
        *)
            echo -e "${RED}❌ 不支持的操作系统: $os${NC}"
            exit 1
            ;;
    esac
}

# 查找并删除二进制文件
uninstall_binary() {
    local found=false
    local locations=(
        "/usr/local/bin/$BINARY_NAME"
        "$HOME/.local/bin/$BINARY_NAME"
        "$HOME/bin/$BINARY_NAME"
    )
    
    # Windows 特殊路径
    if [ "$OS" = "windows" ]; then
        if [ -n "$USERPROFILE" ]; then
            locations+=("$USERPROFILE/bin/$BINARY_NAME")
        fi
        locations+=("./$BINARY_NAME")
    fi
    
    echo -e "${BLUE}🔍 正在查找 SSE Client...${NC}"
    
    for location in "${locations[@]}"; do
        if [ -f "$location" ]; then
            echo -e "${YELLOW}📍 找到: $location${NC}"
            
            # 检查是否需要 sudo
            if [[ "$location" == "/usr/local/bin/"* ]] && [ ! -w "$(dirname "$location")" ]; then
                echo -e "${YELLOW}🔐 需要管理员权限删除系统文件${NC}"
                sudo rm -f "$location"
            else
                rm -f "$location"
            fi
            
            if [ ! -f "$location" ]; then
                echo -e "${GREEN}✅ 已删除: $location${NC}"
                found=true
            else
                echo -e "${RED}❌ 删除失败: $location${NC}"
            fi
        fi
    done
    
    if [ "$found" = false ]; then
        echo -e "${YELLOW}⚠️  未找到 SSE Client 安装${NC}"
        echo -e "${BLUE}💡 可能的位置:${NC}"
        for location in "${locations[@]}"; do
            echo -e "   $location"
        done
    fi
}

# 删除配置文件（可选）
remove_config() {
    local config_locations=(
        "$HOME/.config/sse-client"
        "$HOME/.sse-client"
    )
    
    echo -e "\n${BLUE}🗂️  配置文件清理${NC}"
    echo -e "${YELLOW}是否删除配置文件？ (y/N):${NC}"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        for config_dir in "${config_locations[@]}"; do
            if [ -d "$config_dir" ]; then
                echo -e "${YELLOW}📁 删除配置目录: $config_dir${NC}"
                rm -rf "$config_dir"
                echo -e "${GREEN}✅ 已删除配置目录${NC}"
            fi
        done
        
        # 删除当前目录的配置文件
        if [ -f "./config.yaml" ]; then
            echo -e "${YELLOW}📄 删除当前目录配置文件: ./config.yaml${NC}"
            rm -f "./config.yaml"
            echo -e "${GREEN}✅ 已删除配置文件${NC}"
        fi
    else
        echo -e "${BLUE}💾 保留配置文件${NC}"
    fi
}

# 主函数
main() {
    echo -e "${BLUE}🗑️  SSE Client 卸载程序${NC}"
    echo -e "${YELLOW}⚠️  这将删除 SSE Client 及其相关文件${NC}"
    echo -e "${YELLOW}是否继续？ (y/N):${NC}"
    read -r confirm
    
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo -e "${BLUE}❌ 取消卸载${NC}"
        exit 0
    fi
    
    detect_platform
    uninstall_binary
    remove_config
    
    echo -e "\n${GREEN}🎉 SSE Client 卸载完成！${NC}"
    echo -e "${BLUE}👋 感谢使用 SSE Client${NC}"
}

# 执行主函数
main "$@"