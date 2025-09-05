# SSE Client

一个基于 Go 语言开发的简洁高效 AI 命令行助手，支持多种 AI 模型，提供对话、命令生成、文件处理等功能。

## 🏗️ 项目架构

- **cmd**: 命令行入口和参数处理
- **internal**: 核心业务逻辑实现
  - 配置管理
  - 命令处理
  - 安全控制
  - SSE 客户端
- **providers**: 多模型适配层
  - OpenAI
  - Anthropic
  - Bailian
  - Deepseek
  - Google
- **configs**: 配置文件模板
- **scripts**: 安装脚本

## ✨ 特色优势

- 🎯 **多模型支持**: 支持 OpenAI、Anthropic、百炼、Deepseek、Google 等多个 AI 模型
- 🔄 **SSE 实时响应**: 采用 Server-Sent Events 技术，实现流式对话体验
- 🛡️ **安全性设计**: 内置 API 密钥管理和安全控制机制
- 🎨 **模块化架构**: 清晰的分层设计，易于扩展新的模型支持
- 🚀 **跨平台支持**: 支持 macOS、Linux、Windows、FreeBSD 等多个平台

## 📚 文档导航

- 📖 **[完整使用指南](SSE_CLIENT_GUIDE.md)** - 详细教程和高级功能
- ⚡ **[实战案例集锦](USAGE_EXAMPLES.md)** - 工具集成与应用案例

## 🚀 快速安装

### 方法1：一键安装（推荐）

```bash
# macOS / Linux / Windows
curl -fsSL https://raw.githubusercontent.com/randolph555/sse-client/main/install.sh | bash



# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/randolph555/sse-client/releases/latest/download/sse-windows-amd64.exe" -OutFile "sse.exe"
```
# SSE Client

一个简洁高效的 AI 命令行助手，支持命令生成、文件处理。

## 📚 文档导航

- 📖 **[完整使用指南](SSE_CLIENT_GUIDE.md)** - 详细教程和高级功能
- ⚡ **[实战案例集锦](USAGE_EXAMPLES.md)** - 工具集成与应用案例



## ⚡ 立即使用

```bash
# 设置 API 密钥（选择一个）
export OPENAI_API_KEY="your-key"
export BAILIAN_API_KEY="your-key"  
export DEEPSEEK_API_KEY="your-key"

# AI 对话
sse "你好，请介绍一下自己"

# 生成命令
sse -c "查看系统状态"

# 分析文件
sse "总结这个文档" -f README.md
```

## 🎯 主要功能

| 功能 | 命令示例 | 说明 |
|------|----------|------|
| 🤖 **AI 对话** | `sse "解释量子计算"` | 与 AI 自然对话 |
| 🔧 **命令生成** | `sse -c "清理日志"` | 生成系统命令 |
| 📁 **文件处理** | `sse "优化代码" -e main.go` | 分析和编辑文件 |
| 🖼️ **图片分析** | `sse "描述图片" -i photo.jpg` | 视觉模型分析 |
| 🔄 **管道处理** | `ps aux \| sse "找出高CPU进程"` | 分析命令输出 |

## 🛠️ 技术栈

- **语言**: Go 1.24+ (基于 go.mod)
- **依赖管理**: Go Modules
- **命令行框架**: Cobra CLI
- **配置管理**: YAML 配置 + 环境变量
- **构建工具**: Make + 跨平台构建
- **SSE通信**: 原生HTTP客户端

## 📦 预编译版本

已构建完成的二进制文件（位于 `dist/` 目录）：

- **macOS**: `sse-darwin-amd64`, `sse-darwin-arm64`
- **Linux**: `sse-linux-amd64`, `sse-linux-arm64`
- **Windows**: `sse-windows-amd64.exe`, `sse-windows-arm64.exe`
- **FreeBSD**: `sse-freebsd-amd64`

对应压缩包：
- macOS: `.tar.gz` 格式
- Linux: `.tar.gz` 格式  
- Windows: `.zip` 格式
- FreeBSD: `.tar.gz` 格式

## 🔧 构建说明

### 本地构建
```bash
make build        # 本地构建到 build/ 目录
make build-all    # 跨平台构建到 dist/ 目录
make release      # 生成发布压缩包
```

### 文件结构
```
sse-client/
├── cmd/main.go              # 程序入口
├── internal/                # 核心逻辑
│   ├── client.go           # SSE客户端
│   ├── commands.go       # 命令处理
│   ├── config.go         # 配置管理
│   ├── handlers.go       # 请求处理器
│   └── safety.go         # 安全控制
├── providers/             # AI模型适配器
│   ├── openai.go         # OpenAI API
│   ├── anthropic.go      # Anthropic API
│   ├── bailian.go        # 百炼API
│   ├── deepseek.go       # DeepSeek API
│   ├── google.go         # Google API
│   └── utils.go          # 工具函数
├── configs/               # 配置模板
│   ├── config.example.yaml
│   └── .env.example
├── scripts/install.sh     # 安装脚本
├── dist/                  # 预编译文件
└── Makefile              # 构建配置
```

**让 AI 成为你的终端超能力！** 🚀