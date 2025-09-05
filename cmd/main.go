package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sse-client/internal"
)

var (
	cfgFile     string
	temperature float64
	maxTokens   int
	timeout     int
	imagePath   string
	filePath    string // -f 参数：文件路径
	editPath    string // -e 参数：编辑文件路径
	executeMode bool   // -y 参数：是否直接执行命令
	commandMode bool   // -c 参数：命令模式
)

var rootCmd = &cobra.Command{
	Use:   "sse [message] OR sse [model] [message] OR sse [provider] [model] [message] OR command | sse",
	Short: "A simple SSE client for AI models | 简洁的 AI 模型 SSE 客户端",
	Long: `A command-line SSE client that supports multiple AI providers with streaming responses.
支持多个 AI 提供商流式响应的命令行 SSE 客户端。

Examples | 使用示例:
  # Normal conversation mode (default) | 普通对话模式（默认）
  sse "你好，请介绍一下自己"
  sse "Hello, explain quantum computing"
  sse qwen-max "你好，请介绍一下自己"
  sse bailian qwen-max "Hello, explain quantum computing"
  
  # File processing | 文件处理
  sse "总结这个文件内容" -f document.pdf
  sse "分析这个代码文件" -f main.go
  sse qwen-max "翻译这个文档" -f readme.txt
  
  # Vision models | 视觉模型
  sse qwen-vl-max "请描述这张图片" --image /path/to/image.jpg
  sse bailian qwen-vl-max "请描述这张图片" --image /path/to/image.jpg
  
  # Command mode | 命令模式
  sse -c "帮我清理系统垃圾文件"              # Generate commands (safe) | 生成命令（安全）
  sse -c qwen-max "列出所有进程"             # Command mode with model | 指定模型的命令模式
  sse -c bailian qwen-max "检查磁盘使用"     # Command mode with provider/model | 指定提供商和模型的命令模式
  sse -c -y "清理系统"                      # Execute commands directly | 直接执行命令
  
  # Pipe input | 管道输入
  df -h | sse "分析磁盘使用情况"             # Normal analysis | 普通分析
  docker ps | sse -c "检查容器状态"         # Generate commands | 生成命令
  kubectl get pods | sse -c "分析 Pod 状态" # Generate kubectl commands | 生成 kubectl 命令`,
	Args: cobra.RangeArgs(0, 3),
	Run:  runSSE,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default search: ./config.yaml, ~/.config/sse-client/config.yaml) | 配置文件 (默认搜索: ./config.yaml, ~/.config/sse-client/config.yaml)")
	rootCmd.PersistentFlags().Float64VarP(&temperature, "temperature", "t", 0.7, "sampling temperature | 采样温度")
	rootCmd.PersistentFlags().IntVarP(&maxTokens, "max-tokens", "m", 4096, "maximum tokens | 最大 token 数")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 30, "request timeout in seconds | 请求超时时间（秒）")
	rootCmd.PersistentFlags().StringVarP(&imagePath, "image", "i", "", "path to image file (for vision models) | 图片文件路径（用于视觉模型）")
	rootCmd.PersistentFlags().StringVarP(&filePath, "file", "f", "", "path to file for content analysis | 文件路径（用于内容分析）")
	rootCmd.PersistentFlags().StringVarP(&editPath, "edit", "e", "", "path to file for editing | 文件路径（用于编辑修改）")
	rootCmd.PersistentFlags().BoolVarP(&executeMode, "yes", "y", false, "execute commands directly | 直接执行命令")
	rootCmd.PersistentFlags().BoolVarP(&commandMode, "command", "c", false, "command mode for generating/executing commands | 命令模式，用于生成/执行命令")

	// 添加所有子命令
	for _, cmd := range internal.CreateCommands() {
		rootCmd.AddCommand(cmd)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runSSE(cmd *cobra.Command, args []string) {
	// 设置应用程序配置
	internal.SetAppConfig(internal.AppConfig{
		CfgFile:     cfgFile,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Timeout:     timeout,
		ImagePath:   imagePath,
		FilePath:    filePath,
		EditPath:    editPath,
		ExecuteMode: executeMode,
		CommandMode: commandMode,
	})

	// 调用处理函数
	internal.HandleSSE(args)
}
