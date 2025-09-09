# SSE Client

🚀 **基于 Go 语言的高性能 AI 命令行助手**

一个简洁高效的 AI 命令行工具，采用 Go 语言开发，支持多种 AI 模型，提供对话、命令生成、文件处理等功能。

## ⚡ Go 语言优势

- **🚀 高性能**: Go 原生编译，启动速度快，内存占用低
- **📦 单文件部署**: 无依赖的静态编译二进制文件，下载即用
- **🔄 并发处理**: Go 协程支持，SSE 流式响应性能优异
- **🛡️ 内存安全**: Go 垃圾回收机制，避免内存泄漏
- **🌐 跨平台**: 一次编写，多平台运行，支持8个平台架构
- **⚡ 快速构建**: Go 模块化设计，编译速度极快

## 🎯 技术特性

- **SSE 流式响应**: 实时显示 AI 回复，体验流畅
- **智能模型路由**: 自动识别模型提供商，无需手动指定
- **安全 API 管理**: 环境变量配置，密钥安全存储
- **命令行优化**: Cobra 框架，参数解析高效准确
- **文件处理**: 支持多种文件格式分析和编辑
- **管道集成**: 与系统命令无缝集成，提升工作效率

## 📚 文档导航

- 📖 **[完整使用指南](docs/SSE_CLIENT_GUIDE.md)** - 详细教程和高级功能
- ⚡ **[实战案例集锦](docs/USAGE_EXAMPLES.md)** - 工具集成与应用案例

## 🚀 快速安装

### 一键安装（推荐）

```bash
# 国内用户（推荐，使用代理加速）
curl -fsSL http://gh.cdn01.cn/https://raw.githubusercontent.com/randolph555/sse-client/main/scripts/install-zh.sh | bash

# 有科学上网,或者可以访问到github
curl -fsSL https://raw.githubusercontent.com/randolph555/sse-client/main/scripts/install.sh | bash
```

### 手动安装

**方式1：从Releases下载**
```bash
# 从GitHub Releases下载最新版本
wget https://github.com/randolph555/sse-client/releases/latest/download/sse-linux-amd64.tar.gz
tar -xzf sse-linux-amd64.tar.gz && sudo mv sse-linux-amd64 /usr/local/bin/sse
```

**方式2：直接下载预构建版本**
```bash
# 如果GitHub Actions排队，可直接下载预构建文件（国内加速）
curl -fsSL http://gh.cdn01.cn/https://raw.githubusercontent.com/randolph555/sse-client/main/dist/sse-linux-amd64 -o sse
chmod +x sse && sudo mv sse /usr/local/bin/
```

### 卸载

```bash
# 使用卸载脚本（推荐）
curl -fsSL https://raw.githubusercontent.com/randolph555/sse-client/main/scripts/uninstall.sh | bash

# 手动卸载
sudo rm -f /usr/local/bin/sse
sudo rm -rf /usr/local/bin/sse-configs
```


## ⚡ 快速开始

```bash
# 1. 设置 API 密钥（选择一个）
export OPENAI_API_KEY="your-key"
export BAILIAN_API_KEY="your-key"  
export DEEPSEEK_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
export GOOGLE_API_KEY="your-key"

# 2. 查看配置状态
sse config

# 3. 开始使用
sse "你好，请介绍一下自己"
```

## 🎯 核心功能

### 💬 AI 对话
```bash
# 基础对话
sse "解释量子计算"

# 指定模型
sse qwen-max "用中文解释机器学习"
sse gpt-4o "Write a Python function"

# 指定提供商和模型
sse bailian qwen-max "分析这个问题"
```

### 🔧 命令生成
```bash
# 生成系统命令
sse -c "查看系统状态"
sse -c "清理临时文件"

# 直接执行命令（谨慎使用）
sse -c -y "显示当前目录文件"
```

### 📁 文件处理
```bash
# 分析文件内容
sse "总结这个文档" -f README.md
sse "分析代码问题" -f main.go

# 编辑文件
sse "优化这个配置" -e config.yaml
```

### 🖼️ 图片分析
```bash
# 视觉模型分析图片
sse qwen-vl-max "描述这张图片" -i photo.jpg
sse "提取图片中的文字" -i screenshot.png
```

### 🔄 管道处理
```bash
# 分析命令输出
df -h | sse "分析磁盘使用情况"
ps aux | sse "找出占用CPU最高的进程"
docker ps | sse -c "生成容器管理命令"
```

## ⚙️ 配置说明

### 基础配置（必需）
只需设置一个 AI 提供商的 API 密钥即可开始使用：

```bash
# 选择一个设置即可
export OPENAI_API_KEY="your-openai-key"
export BAILIAN_API_KEY="your-bailian-key"  
export DEEPSEEK_API_KEY="your-deepseek-key"
export ANTHROPIC_API_KEY="your-anthropic-key"
export GOOGLE_API_KEY="your-google-key"
```

### 高级配置（可选）
如需自定义配置，可以使用配置文件：

```bash
# 方法1：环境变量文件（推荐）
cp configs/.env.example .env
# 编辑 .env 文件，然后: source .env

# 方法2：YAML配置文件
cp configs/config.example.yaml config.yaml
# 编辑 config.yaml 文件
```

配置文件位置：
- `./config.yaml` (项目目录)
- `~/.config/sse-client/config.yaml` (用户目录)

### 配置管理命令
```bash
sse config              # 查看当前配置状态
sse env                 # 查看支持的环境变量
sse list                # 列出所有支持的模型
sse set default openai gpt-4o    # 设置默认模型
sse test openai         # 测试提供商配置
```

## 🎨 高级用法

### 参数调整
```bash
# 调整创造性
sse "写一首诗" --temperature 0.8

# 限制输出长度
sse "简单解释" --max-tokens 200

# 设置超时时间
sse "复杂问题" --timeout 60
```

### 工作流示例
```bash
# 1. 系统诊断
df -h | sse "分析磁盘使用" > disk_analysis.txt

# 2. 基于分析生成命令
sse -c "根据分析结果生成清理命令" -f disk_analysis.txt

# 3. 代码审查
sse "检查代码质量和安全问题" -f main.go
```

## 🔧 构建和开发

### 使用 Make 构建
```bash
# 安装依赖
make deps

# 本地构建（当前平台）
make build

# 跨平台构建
make build-all

# 生成发布包（包含压缩文件）
make release

# 本地安装
make install-local

# 清理构建文件
make clean
```

### 使用 Go 直接构建
```bash
# 安装依赖
go mod tidy

# 本地构建
go build -o sse ./cmd/

# 跨平台构建示例
GOOS=linux GOARCH=amd64 go build -o sse-linux-amd64 ./cmd/
GOOS=windows GOARCH=amd64 go build -o sse-windows-amd64.exe ./cmd/
GOOS=darwin GOARCH=arm64 go build -o sse-darwin-arm64 ./cmd/

# 运行测试
go test ./...

# 直接运行（开发模式）
go run ./cmd/ "你的问题"
```

### 技术特点
- **📦 单文件部署**: Go 静态编译，无依赖
- **⚡ 并发处理**: 基于 Go 协程的 SSE 连接
- **🌐 跨平台**: 支持 8 个主流平台架构
- **🔧 简单配置**: 环境变量即可开始使用

### 实际文件大小
- **二进制文件**: ~7.2MB (未压缩)
- **压缩包**: ~2.7-2.9MB (.tar.gz)
- **启动时间**: ~30ms (实测)




## 🏗️ 项目架构

```
sse-client/
├── cmd/                   # 命令行入口
├── internal/              # 核心业务逻辑
│   ├── client.go         # SSE 客户端
│   ├── commands.go       # 子命令实现
│   ├── config.go         # 配置管理
│   ├── handlers.go       # 请求处理
│   └── safety.go         # 安全控制
├── providers/             # AI 提供商适配
├── configs/               # 配置模板
├── scripts/               # 安装脚本
└── docs/                  # 文档
```

## 🤝 支持的 AI 提供商

- **OpenAI**: GPT-4o, GPT-4, GPT-3.5 等
- **Anthropic**: Claude-3.5, Claude-3 等  
- **阿里云百炼**: Qwen 系列模型
- **DeepSeek**: DeepSeek Chat, Coder 等
- **Google**: Gemini 系列模型

## 🚀 开发和发布

### 脚本说明

**`scripts/install.sh`** - 用户安装脚本
- 用途：普通用户下载和安装 SSE Client
- 功能：自动检测系统架构，下载对应版本，安装到系统路径
- 使用：`curl -fsSL https://raw.githubusercontent.com/.../install.sh | bash`

**`scripts/release.sh`** - 开发者发布脚本  
- 用途：项目维护者发布新版本
- 功能：运行测试、创建 Git Tag、触发自动构建发布
- 使用：`./scripts/release.sh` (需要 git 仓库写权限)

### GitHub Actions 自动化

**代码提交时**：
- 自动运行测试
- 构建检查

**创建 Tag 时** (如 `git tag v1.0.1`):
- 自动构建所有平台版本
- 创建 GitHub Release
- 上传二进制文件
- 生成更新日志

### 发布流程
```bash
# 开发者发布新版本
./scripts/release.sh
# 输入版本号 → 自动测试 → 创建 Tag → 推送 → 触发构建发布

# 用户安装使用
curl -fsSL https://raw.githubusercontent.com/randolph555/sse-client/main/scripts/install.sh | bash
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

**让 AI 成为你的终端超能力！** 🚀

## 🔧 CI/CD 优化说明

现在CI构建已优化，只在以下文件变更时触发：
- Go源代码文件 (`**.go`)
- 依赖文件 (`go.mod`, `go.sum`)
- 构建文件 (`Makefile`)
- 配置文件 (`configs/**`)
- 核心目录 (`cmd/**`, `internal/**`, `providers/**`)
- CI配置 (`.github/workflows/**`)

文档更新（如README.md）不再触发不必要的构建，节省CI资源。

