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

# 查找并删除二进制文件和配置目录
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
            
            # 获取安装目录
            local install_dir=$(dirname "$location")
            
            # 检查是否需要 sudo
            if [[ "$location" == "/usr/local/bin/"* ]] && [ ! -w "$(dirname "$location")" ]; then
                echo -e "${YELLOW}🔐 需要管理员权限删除系统文件${NC}"
                sudo rm -f "$location"
                
                # 删除配置目录（新的sse-configs和旧的configs）
                if [ -d "$install_dir/sse-configs" ]; then
                    echo -e "${YELLOW}📁 删除配置目录: $install_dir/sse-configs${NC}"
                    sudo rm -rf "$install_dir/sse-configs"
                fi
                if [ -d "$install_dir/configs" ]; then
                    echo -e "${YELLOW}📁 清理旧配置目录: $install_dir/configs${NC}"
                    sudo rm -rf "$install_dir/configs"
                fi
            else
                rm -f "$location"
                
                # 删除配置目录（新的sse-configs和旧的configs）
                if [ -d "$install_dir/sse-configs" ]; then
                    echo -e "${YELLOW}📁 删除配置目录: $install_dir/sse-configs${NC}"
                    rm -rf "$install_dir/sse-configs"
                fi
                if [ -d "$install_dir/configs" ]; then
                    echo -e "${YELLOW}📁 清理旧配置目录: $install_dir/configs${NC}"
                    rm -rf "$install_dir/configs"
                fi
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

# 删除配置文件（交互式）
remove_config() {
    local config_locations=(
        "$HOME/.config/sse-client"
        "$HOME/.sse-client"
    )
    
    echo -e "\n${BLUE}🗂️  配置文件清理${NC}"
    
    # 检查是否有配置文件存在
    local has_config=false
    for config_dir in "${config_locations[@]}"; do
        if [ -d "$config_dir" ]; then
            has_config=true
            break
        fi
    done
    
    if [ -f "./config.yaml" ]; then
        has_config=true
    fi
    
    if [ "$has_config" = false ]; then
        echo -e "${BLUE}💡 未找到配置文件${NC}"
        return
    fi
    
    # 只在交互模式下询问
    if [ -t 0 ]; then
        echo -e "${YELLOW}是否删除配置文件？ (y/N):${NC}"
        read -r response
        
        if [[ "$response" =~ ^[Yy]$ ]]; then
            remove_config_force
        else
            echo -e "${BLUE}💾 保留配置文件${NC}"
        fi
    else
        echo -e "${BLUE}💾 保留配置文件（非交互模式，使用 -c 参数可删除）${NC}"
    fi
}

# 主函数
main() {
    local force=false
    local remove_configs=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--force)
                force=true
                shift
                ;;
            -c|--remove-config)
                remove_configs=true
                shift
                ;;
            -h|--help)
                echo -e "${BLUE}SSE Client 卸载脚本${NC}"
                echo -e "用法: $0 [选项]"
                echo -e ""
                echo -e "选项:"
                echo -e "  -f, --force           强制卸载，不询问确认"
                echo -e "  -c, --remove-config   同时删除配置文件"
                echo -e "  -h, --help           显示此帮助信息"
                echo -e ""
                echo -e "示例:"
                echo -e "  $0                    交互式卸载"
                echo -e "  $0 -f                 强制卸载"
                echo -e "  $0 -f -c              强制卸载并删除配置"
                echo -e ""
                echo -e "通过管道使用:"
                echo -e "  curl -fsSL https://raw.githubusercontent.com/randolph555/sse-client/main/scripts/uninstall.sh | bash -s -- -f"
                exit 0
                ;;
            *)
                echo -e "${RED}❌ 未知参数: $1${NC}"
                echo -e "使用 -h 或 --help 查看帮助"
                exit 1
                ;;
        esac
    done
    
    echo -e "${BLUE}🗑️  SSE Client 卸载程序${NC}"
    echo -e "${YELLOW}⚠️  这将删除 SSE Client 及其相关文件${NC}"
    
    # 检查是否通过管道执行（stdin不是终端）
    if [ ! -t 0 ] && [ "$force" = false ]; then
        echo -e "${YELLOW}💡 检测到通过管道执行，使用 -f 参数强制卸载：${NC}"
        echo -e "   curl -fsSL https://raw.githubusercontent.com/randolph555/sse-client/main/scripts/uninstall.sh | bash -s -- -f"
        exit 1
    fi
    
    if [ "$force" = false ]; then
        echo -e "${YELLOW}是否继续？ (y/N):${NC}"
        read -r confirm
        
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            echo -e "${BLUE}❌ 取消卸载${NC}"
            exit 0
        fi
    else
        echo -e "${GREEN}🚀 强制卸载模式${NC}"
    fi
    
    detect_platform
    uninstall_binary
    
    # 处理配置文件删除
    if [ "$remove_configs" = true ]; then
        remove_config_force
    elif [ "$force" = false ]; then
        remove_config
    else
        echo -e "${BLUE}💾 保留配置文件（使用 -c 参数可删除配置）${NC}"
    fi
    
    echo -e "\n${GREEN}🎉 SSE Client 卸载完成！${NC}"
    echo -e "${BLUE}👋 感谢使用 SSE Client${NC}"
}

# 强制删除配置文件（不询问）
remove_config_force() {
    local config_locations=(
        "$HOME/.config/sse-client"
        "$HOME/.sse-client"
    )
    
    echo -e "\n${BLUE}🗂️  删除配置文件...${NC}"
    
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
}

# 执行主函数
main "$@"