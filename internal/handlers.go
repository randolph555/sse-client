package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AppConfig 保存应用程序配置参数
type AppConfig struct {
	CfgFile     string
	Temperature float64
	MaxTokens   int
	Timeout     int
	ImagePath   string
	FilePath    string
	EditPath    string
	ExecuteMode bool
	CommandMode bool
}

// 全局配置实例
var appConfig AppConfig

// SetAppConfig 设置应用程序配置
func SetAppConfig(cfg AppConfig) {
	appConfig = cfg
}

// 处理主要的 SSE 逻辑
func HandleSSE(args []string) {
	var provider, model, message string

	if err := loadConfig(appConfig.CfgFile); err != nil {
		fmt.Printf("Error loading config | 配置加载错误: %v\n", err)
		os.Exit(1)
	}

	// 检查是否有 stdin 输入（管道输入）
	stdinData := readStdinIfAvailable()

	// 解析参数
	provider, model, message = parseArgs(args, stdinData)

	client := NewSSEClient()

	// 处理文件输入
	if appConfig.FilePath != "" {
		fileContent, err := readFileContent(appConfig.FilePath)
		if err != nil {
			fmt.Printf("Error reading file | 文件读取错误: %v\n", err)
			os.Exit(1)
		}
		// 将文件内容添加到消息中
		message = message + "\n\n文件内容:\n" + fileContent
	}

	// 处理文件编辑
	if appConfig.EditPath != "" {
		err := handleFileEdit(client, provider, model, appConfig.EditPath, message, appConfig.ImagePath, appConfig.Temperature, appConfig.MaxTokens, appConfig.Timeout)
		if err != nil {
			fmt.Printf("Error editing file | 文件编辑错误: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 根据模式处理
	var err error
	if appConfig.CommandMode {
		// 命令模式：生成或执行命令
		if appConfig.ExecuteMode {
			// -c -y: 命令模式 + 直接执行
			err = handleCommandExecution(client, provider, model, message, appConfig.ImagePath, appConfig.Temperature, appConfig.MaxTokens, appConfig.Timeout)
		} else {
			// -c: 命令模式，只输出命令
			err = handleCommandOutput(client, provider, model, message, appConfig.ImagePath, appConfig.Temperature, appConfig.MaxTokens, appConfig.Timeout)
		}
	} else {
		// 普通对话模式（默认）
		err = handleNormalConversation(client, provider, model, message, appConfig.ImagePath, appConfig.Temperature, appConfig.MaxTokens, appConfig.Timeout)
	}

	if err != nil {
		fmt.Printf("Error | 错误: %v\n", err)
		os.Exit(1)
	}
}

// 解析命令行参数
func parseArgs(args []string, stdinData string) (provider, model, message string) {
	if len(args) == 0 {
		// Format: command | sse - use stdin as message with default provider/model
		if stdinData == "" {
			fmt.Printf("No message provided. Use: sse \"your message\" or pipe data: command | sse\n")
			fmt.Printf("未提供消息。请使用: sse \"您的消息\" 或管道输入: 命令 | sse\n")
			os.Exit(1)
		}
		message = stdinData
		defaultProvider, defaultModel := getDefaultProvider()
		if defaultProvider == "" || defaultModel == "" {
			fmt.Printf("No default provider/model set. Use: sse set default <provider> <model>\n")
			fmt.Printf("未设置默认提供商/模型。请使用: sse set default <provider> <model>\n")
			os.Exit(1)
		}
		provider = defaultProvider
		model = defaultModel
	} else if len(args) == 1 {
		// Format: sse [message] or command | sse [additional_message]
		if stdinData != "" {
			// 组合 stdin 数据和用户消息
			message = stdinData + "\n\n" + args[0]
		} else {
			message = args[0]
		}
		defaultProvider, defaultModel := getDefaultProvider()
		if defaultProvider == "" || defaultModel == "" {
			fmt.Printf("No default provider/model set. Use: sse set default <provider> <model>\n")
			fmt.Printf("未设置默认提供商/模型。请使用: sse set default <provider> <model>\n")
			os.Exit(1)
		}
		provider = defaultProvider
		model = defaultModel
	} else if len(args) == 2 {
		// Format: sse [model] [message] or command | sse [model] [additional_message]
		model = args[0]
		if stdinData != "" {
			message = stdinData + "\n\n" + args[1]
		} else {
			message = args[1]
		}
	} else if len(args) == 3 {
		// Format: sse [provider] [model] [message] or command | sse [provider] [model] [additional_message]
		provider = args[0]
		model = args[1]
		if stdinData != "" {
			message = stdinData + "\n\n" + args[2]
		} else {
			message = args[2]
		}
	}

	return provider, model, message
}

// readStdinIfAvailable 检查并读取 stdin 数据（如果有的话）
func readStdinIfAvailable() string {
	// 检查 stdin 是否有数据可读
	stat, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}

	// 检查是否是管道输入或重定向输入
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// 有管道输入，读取所有数据
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return ""
		}
		return strings.Join(lines, "\n")
	}

	return ""
}

// handleCommandOutput 处理命令输出模式（默认行为：只输出命令，不执行）
func handleCommandOutput(client *SSEClient, provider, model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// 修改提示词以获得纯命令输出
	modifiedMessage := message + "\n\n重要：请只返回可以直接执行的命令行命令，不要包含任何解释文字、描述或说明。每个命令单独一行。不要使用代码块格式。请确保命令在 macOS 和 Linux 系统上都能正常工作。\n\nIMPORTANT: Only return executable command line commands without any explanations, descriptions, or commentary. One command per line. Do not use code block formatting. Ensure commands work on both macOS and Linux systems."

	// 获取完整的AI响应（非流式）
	response, err := getFullResponse(client, provider, model, modifiedMessage, imagePath, temperature, maxTokens, timeout)
	if err != nil {
		return err
	}

	// 提取并输出纯命令
	commands := extractCommands(response)
	if len(commands) > 0 {
		fmt.Println(strings.Join(commands, "\n"))
	} else {
		// 如果没有找到命令，输出原始响应
		fmt.Println(response)
	}

	return nil
}

// handleCommandExecution 处理命令执行模式（-y 参数：获取命令并直接执行）
func handleCommandExecution(client *SSEClient, provider, model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// 修改提示词以获得纯命令输出
	modifiedMessage := message + "\n\n重要：请只返回可以直接执行的命令行命令，不要包含任何解释文字、描述或说明。每个命令单独一行。不要使用代码块格式。请确保命令在 macOS 和 Linux 系统上都能正常工作。\n\nIMPORTANT: Only return executable command line commands without any explanations, descriptions, or commentary. One command per line. Do not use code block formatting. Ensure commands work on both macOS and Linux systems."

	// 获取完整的AI响应（非流式）
	response, err := getFullResponse(client, provider, model, modifiedMessage, imagePath, temperature, maxTokens, timeout)
	if err != nil {
		return err
	}

	// 提取命令
	commands := extractCommands(response)
	if len(commands) == 0 {
		fmt.Printf("No executable commands found in AI response\n")
		fmt.Printf("AI响应中未找到可执行命令\n")
		return nil
	}

	// 执行每个命令
	for _, cmd := range commands {
		fmt.Printf("🚀 Executing: %s\n", cmd)

		// 使用 bash 执行命令
		execCmd := exec.Command("bash", "-c", cmd)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		if err := execCmd.Run(); err != nil {
			fmt.Printf("❌ Command failed: %v\n", err)
		} else {
			fmt.Printf("✅ Command completed successfully\n")
		}
		fmt.Println()
	}

	return nil
}

// handleFileAnalysis 处理文件分析模式（直接分析文件内容，不生成命令）
func handleFileAnalysis(client *SSEClient, provider, model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// 直接使用流式响应进行文件内容分析
	return client.StreamWithProvider(provider, model, message, imagePath, temperature, maxTokens, timeout)
}

// handlePipeAnalysis 处理管道输入分析模式（用户自己执行了命令，分析命令输出结果）
func handlePipeAnalysis(client *SSEClient, provider, model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// 直接使用流式响应进行数据分析，不生成命令
	// 这里的 message 已经包含了管道输入的数据和用户的分析请求
	return client.StreamWithProvider(provider, model, message, imagePath, temperature, maxTokens, timeout)
}

// handleNormalConversation 处理普通对话模式（默认模式：纯对话，不生成命令）
func handleNormalConversation(client *SSEClient, provider, model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// 直接使用流式响应进行对话，不修改消息内容
	return client.StreamWithProvider(provider, model, message, imagePath, temperature, maxTokens, timeout)
}

// handleFileEdit 处理文件编辑模式（读取文件，根据指令修改，写回文件）
func handleFileEdit(client *SSEClient, provider, model, filePath, instruction, imagePath string, temperature float64, maxTokens, timeout int) error {
	// 读取文件内容，如果文件不存在则创建空文件
	var originalContent string
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("📝 File does not exist, creating new file: %s\n", filePath)
		originalContent = ""
	} else {
		content, err := readFileContent(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
		originalContent = content
	}

	// 构造编辑提示词
	var editPrompt string
	if originalContent == "" {
		editPrompt = fmt.Sprintf(`请创建一个新文件，要求：%s

请直接返回完整的文件内容，不要添加任何解释、代码块标记或其他格式。`, instruction)
	} else {
		editPrompt = fmt.Sprintf(`请修改以下文件内容，要求：%s

原文件内容：
%s

请直接返回修改后的完整文件内容，不要添加任何解释、代码块标记或其他格式。`, instruction, originalContent)
	}

	// 获取AI的完整响应
	newContent, err := getFullResponse(client, provider, model, editPrompt, imagePath, temperature, maxTokens, timeout)
	if err != nil {
		return fmt.Errorf("failed to get AI response: %v", err)
	}

	// 写入文件
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("✅ File edited successfully: %s\n", filePath)
	return nil
}

// getFullResponse 获取完整的AI响应（非流式）
func getFullResponse(client *SSEClient, provider, model, message, imagePath string, temperature float64, maxTokens, timeout int) (string, error) {
	// 使用新的 GetFullResponseWithProvider 方法获取完整响应
	return client.GetFullResponseWithProvider(provider, model, message, imagePath, temperature, maxTokens, timeout)
}

// readFileContent 读取文件内容
func readFileContent(filePath string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return string(content), nil
}
