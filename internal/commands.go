package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// CreateCommands 创建所有子命令
func CreateCommands() []*cobra.Command {
	return []*cobra.Command{
		createListCmd(),
		createTestCmd(),
		createConfigCmd(),
		createAddCmd(),
		createSetCmd(),
		createEnvCmd(),
	}
}

func createListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all supported models | 列出所有支持的模型",
		Long:  "List all supported models by provider | 按提供商列出所有支持的模型",
		Run:   ListModels,
	}
}

// ListModels 列出所有支持的模型
func ListModels(cmd *cobra.Command, args []string) {
	if err := loadConfig(""); err != nil {
		fmt.Printf("Error loading config | 配置加载错误: %v\n", err)
		return
	}

	fmt.Println("Available Models | 可用模型:")
	fmt.Println()

	providers := []string{"bailian", "openai", "google", "anthropic", "deepseek"}
	for _, provider := range providers {
		if cfg, exists := getProviderConfig(provider); exists && len(cfg.Models) > 0 {
			fmt.Printf("📦 %s:\n", strings.ToUpper(provider))
			for _, model := range cfg.Models {
				fmt.Printf("  • %s\n", model)
			}
			fmt.Println()
		}
	}
}

func createTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test [provider]",
		Short: "Test provider configuration | 测试提供商配置",
		Long: `Test provider configuration and API key setup | 测试提供商配置和 API key 设置

Available providers | 可用提供商:
  bailian    - Test Bailian provider configuration | 测试百炼提供商配置
  openai     - Test OpenAI provider configuration | 测试 OpenAI 提供商配置
  google     - Test Google Gemini provider configuration | 测试 Google Gemini 提供商配置  
  anthropic  - Test Anthropic Claude provider configuration | 测试 Anthropic Claude 提供商配置
  deepseek   - Test DeepSeek provider configuration | 测试 DeepSeek 提供商配置

Examples | 示例:
  sse test              # Test all configurations | 测试所有配置
  sse test openai       # Test OpenAI configuration | 测试 OpenAI 配置
  sse test google       # Test Google configuration | 测试 Google 配置`,
		Args: cobra.MaximumNArgs(1),
		Run:  TestProvider,
	}
}

// TestProvider 测试提供商配置
func TestProvider(cmd *cobra.Command, args []string) {
	// 实现测试提供商配置的逻辑
}

func createConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show current configuration | 显示当前配置",
		Long: `Show current configuration including environment variables and config file settings | 显示当前配置，包括环境变量和配置文件设置

This command shows all configuration values that would be used, with environment variables taking precedence over config file values.
此命令显示将要使用的所有配置值，环境变量优先于配置文件值。

Examples | 示例:
  sse config            # Show all configuration | 显示所有配置`,
		Run: showConfig,
	}
}

func createAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <provider> model <model_name>",
		Short: "Add custom model to provider | 为提供商添加自定义模型",
		Long: `Add a custom model to a specific provider's model list | 为特定提供商的模型列表添加自定义模型

This allows you to use custom models with the auto-detection feature.
这允许您在自动检测功能中使用自定义模型。

Examples | 示例:
  sse add openai model my-custom-gpt        # Add custom OpenAI model | 添加自定义 OpenAI 模型
  sse add bailian model qwen-custom         # Add custom Bailian model | 添加自定义百炼模型
  sse add google model gemini-custom        # Add custom Google model | 添加自定义 Google 模型
  sse add anthropic model claude-custom     # Add custom Anthropic model | 添加自定义 Anthropic 模型`,
		Args: cobra.ExactArgs(3),
		Run:  addModel,
	}
}

func createSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set default <provider> <model>",
		Short: "Set default provider and model | 设置默认提供商和模型",
		Long: `Set default provider and model for quick access | 设置默认提供商和模型以便快速访问

After setting defaults, you can use commands without specifying provider/model:
设置默认值后，您可以在不指定提供商/模型的情况下使用命令：

Examples | 示例:
  sse set default anthropic claude-3-5-sonnet-20241022    # Set default | 设置默认
  sse "Hello world"                                       # Use default | 使用默认
  sse "识别画面内容" -i image.jpg                          # Use default with image | 使用默认模型处理图片`,
		Args: cobra.ExactArgs(3),
		Run:  setDefault,
	}
}

func showConfig(cmd *cobra.Command, args []string) {
	if err := loadConfig(appConfig.CfgFile); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration:")
	fmt.Println()

	// 显示默认设置
	defaultProvider, defaultModel := getDefaultProvider()
	if defaultProvider != "" && defaultModel != "" {
		fmt.Printf("🎯 Default: %s %s\n", defaultProvider, defaultModel)
		fmt.Println()
	}

	providers := []string{"bailian", "openai", "google", "anthropic", "deepseek"}
	for _, provider := range providers {
		if cfg, exists := getProviderConfig(provider); exists {
			status := "❌"
			if cfg.APIKey != "" {
				status = "✅"
			}

			fmt.Printf("%s %s:\n", status, strings.ToUpper(provider))

			// API Key 环境变量格式
			apiKeyEnv := strings.ToUpper(provider) + "_API_KEY"
			if cfg.APIKey != "" {
				fmt.Printf("  %s=%s\n", apiKeyEnv, cfg.APIKey)
			} else {
				fmt.Printf("  %s=not_configured\n", apiKeyEnv)
			}

			// Base URL 环境变量格式
			baseUrlEnv := strings.ToUpper(provider) + "_BASE_URL"
			if cfg.BaseURL != "" {
				fmt.Printf("  %s=%s\n", baseUrlEnv, cfg.BaseURL)
			}

			fmt.Printf("  Models: %d\n", len(cfg.Models))
			fmt.Println()
		}
	}
}

func addModel(cmd *cobra.Command, args []string) {
	if len(args) != 3 || args[1] != "model" {
		fmt.Println("Usage: sse add <provider> model <model_name>")
		fmt.Println("Example: sse add openai model my-custom-gpt")
		os.Exit(1)
	}

	provider := args[0]
	modelName := args[2]

	// 验证 provider 是否有效
	validProviders := []string{"bailian", "openai", "google", "anthropic", "deepseek"}
	isValidProvider := false
	for _, p := range validProviders {
		if p == provider {
			isValidProvider = true
			break
		}
	}

	if !isValidProvider {
		fmt.Printf("❌ Invalid provider: %s\n", provider)
		fmt.Printf("Available providers | 可用提供商: %s\n", strings.Join(validProviders, ", "))
		os.Exit(1)
	}

	if err := loadConfig(appConfig.CfgFile); err != nil {
		fmt.Printf("Error loading config | 配置加载错误: %v\n", err)
		os.Exit(1)
	}

	// 添加模型到配置
	if err := addModelToConfig(provider, modelName); err != nil {
		fmt.Printf("Error adding model | 添加模型错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully added model '%s' to %s provider\n", modelName, provider)
	fmt.Printf("✅ 成功将模型 '%s' 添加到 %s 提供商\n", modelName, provider)
	fmt.Println()
	fmt.Printf("Now you can use: sse %s \"your message\"\n", modelName)
	fmt.Printf("现在您可以使用: sse %s \"您的消息\"\n", modelName)
}

func setDefault(cmd *cobra.Command, args []string) {
	if err := loadConfig(appConfig.CfgFile); err != nil {
		fmt.Printf("Error loading config | 配置加载错误: %v\n", err)
		os.Exit(1)
	}

	if args[0] != "default" {
		fmt.Printf("Usage: sse set default <provider> <model>\n")
		os.Exit(1)
	}

	provider := args[1]
	model := args[2]

	// 设置默认提供商和模型
	if err := setDefaultProvider(provider, model); err != nil {
		fmt.Printf("Error setting default | 设置默认值错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully set default provider to '%s' with model '%s'\n", provider, model)
	fmt.Printf("✅ 成功设置默认提供商为 '%s'，模型为 '%s'\n", provider, model)
	fmt.Println()
	fmt.Printf("Now you can use: sse \"your message\"\n")
	fmt.Printf("现在您可以使用: sse \"您的消息\"\n")
}

func createEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "Show environment variables | 显示环境变量",
		Long: `Show all supported environment variables for API configuration | 显示所有支持的 API 配置环境变量

This command shows the environment variable names you can use to configure API keys and base URLs for all providers.
此命令显示您可以用来为所有提供商配置 API 密钥和基础 URL 的环境变量名称。

Examples | 示例:
  sse env                    # Show all environment variables | 显示所有环境变量
  sse env --copy             # Show with copy commands | 显示带复制命令的格式`,
		Run: showEnvVars,
	}
}

func showEnvVars(cmd *cobra.Command, args []string) {
	fmt.Println("🌍 Supported Environment Variables | 支持的环境变量")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	providers := []struct {
		name    string
		display string
		baseURL string
	}{
		{"bailian", "阿里云百炼 (Bailian)", "https://dashscope.aliyuncs.com/compatible-mode/v1"},
		{"openai", "OpenAI", "https://api.openai.com/v1"},
		{"google", "Google Gemini", "https://generativelanguage.googleapis.com/v1beta"},
		{"anthropic", "Anthropic Claude", "https://api.anthropic.com"},
		{"deepseek", "DeepSeek", "https://api.deepseek.com/v1"},
	}

	for i, provider := range providers {
		fmt.Printf("📌 %s\n", provider.display)
		fmt.Printf("   API Key:  %s_API_KEY\n", strings.ToUpper(provider.name))
		fmt.Printf("   Base URL: %s_BASE_URL (optional, default: %s)\n", strings.ToUpper(provider.name), provider.baseURL)

		if i < len(providers)-1 {
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Println("💡 Usage Examples | 使用示例:")
	fmt.Println("   # Set API key | 设置 API 密钥")
	fmt.Println("   export OPENAI_API_KEY=\"your-api-key-here\"")
	fmt.Println("   export DEEPSEEK_API_KEY=\"your-deepseek-key\"")
	fmt.Println()
	fmt.Println("   # Set custom base URL (optional) | 设置自定义基础 URL（可选）")
	fmt.Println("   export OPENAI_BASE_URL=\"https://your-proxy.com/v1\"")
	fmt.Println()
	fmt.Println("📄 Configuration Files | 配置文件:")
	fmt.Println("   • Copy .env.example to .env and edit | 复制 .env.example 到 .env 并编辑")
	fmt.Println("   • Or edit config.yaml directly | 或直接编辑 config.yaml")
	fmt.Println()
	fmt.Println("🔍 Check current values | 检查当前值:")
	fmt.Println("   sse config                    # Show current configuration | 显示当前配置")
	fmt.Println("   sse test <provider>           # Test provider setup | 测试提供商设置")
}
