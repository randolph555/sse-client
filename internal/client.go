package internal

import (
	"fmt"
	"sse-client/providers"
)

type SSEClient struct {
	providers map[string]Provider
}

type Provider interface {
	Stream(model, message string, temperature float64, maxTokens, timeout int) error
	StreamWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) error
	GetFullResponse(model, message string, temperature float64, maxTokens, timeout int) (string, error)
	GetFullResponseWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) (string, error)
	SupportsModel(model string) bool
}

func NewSSEClient() *SSEClient {
	// 将配置传递给 providers 包
	providers.SetConfig(providers.Config{
		Providers: make(map[string]providers.ProviderConfig),
	})

	// 如果有全局配置，设置它
	globalConfig := providers.GetConfig()
	if globalConfig.Providers != nil {
		providerConfigs := make(map[string]providers.ProviderConfig)
		for name, cfg := range globalConfig.Providers {
			providerConfigs[name] = providers.ProviderConfig{
				APIKey:  cfg.APIKey,
				BaseURL: cfg.BaseURL,
				Models:  cfg.Models,
			}
		}
		providers.SetConfig(providers.Config{
			Providers: providerConfigs,
		})
	}

	return &SSEClient{
		providers: map[string]Provider{
			"bailian":   providers.NewBailianProvider(),
			"openai":    providers.NewOpenAIProvider(),
			"google":    providers.NewGoogleProvider(),
			"anthropic": providers.NewAnthropicProvider(),
			"deepseek":  providers.NewDeepSeekProvider(),
		},
	}
}

// inferProviderFromModel 根据模型名称推断 provider
func (c *SSEClient) inferProviderFromModel(model string) string {
	// 首先检查自定义模型
	globalConfig := providers.GetConfig()
	if globalConfig.Providers != nil {
		for providerName, cfg := range globalConfig.Providers {
			for _, customModel := range cfg.Models {
				if customModel == model {
					return providerName
				}
			}
		}
	}

	// 然后使用默认的前缀匹配
	// Qwen 系列模型 -> bailian
	if len(model) >= 4 && model[:4] == "qwen" {
		return "bailian"
	}

	// GPT 和 O1 系列模型 -> openai
	if len(model) >= 3 && (model[:3] == "gpt" || model[:2] == "o1") {
		return "openai"
	}

	// Gemini 系列模型 -> google
	if len(model) >= 6 && model[:6] == "gemini" {
		return "google"
	}

	// Claude 系列模型 -> anthropic
	if len(model) >= 6 && model[:6] == "claude" {
		return "anthropic"
	}

	// DeepSeek 系列模型 -> deepseek
	if len(model) >= 8 && model[:8] == "deepseek" {
		return "deepseek"
	}

	// DeepSeek 模型（Hugging Face 格式）-> deepseek
	if len(model) >= 11 && model[:11] == "deepseek-ai" {
		return "deepseek"
	}

	return ""
}

// IsProviderConfigured 检查 provider 是否配置了有效的 API key
func (c *SSEClient) IsProviderConfigured(providerName string) bool {
	if cfg, exists := getProviderConfig(providerName); exists {
		return cfg.APIKey != "" && cfg.APIKey != "your-api-key-here" && cfg.APIKey != "sk-test-key"
	}
	return false
}

func (c *SSEClient) Stream(model, message string, temperature float64, maxTokens, timeout int) error {
	return c.StreamWithImage(model, message, "", temperature, maxTokens, timeout)
}

// StreamWithProvider 支持明确指定 provider 或自动推断
func (c *SSEClient) StreamWithProvider(providerName, model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// 如果明确指定了 provider
	if providerName != "" {
		return c.streamWithSpecificProvider(providerName, model, message, imagePath, temperature, maxTokens, timeout)
	}

	// 如果没有指定 provider，使用自动推断逻辑
	return c.StreamWithImage(model, message, imagePath, temperature, maxTokens, timeout)
}

// streamWithSpecificProvider 使用指定的 provider
func (c *SSEClient) streamWithSpecificProvider(providerName, model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// 检查 provider 是否存在
	provider, exists := c.providers[providerName]
	if !exists {
		return fmt.Errorf("provider not found | 提供商未找到: %s\nAvailable providers | 可用提供商: bailian, openai, google, anthropic, deepseek", providerName)
	}

	// 检查 provider 是否配置了 API key
	if !c.IsProviderConfigured(providerName) {
		return fmt.Errorf("provider '%s' is not configured. Please configure the API key first", providerName)
	}

	fmt.Printf("Using %s provider for model: %s\n", providerName, model)

	if imagePath != "" {
		return provider.StreamWithImage(model, message, imagePath, temperature, maxTokens, timeout)
	}
	return provider.Stream(model, message, temperature, maxTokens, timeout)
}

func (c *SSEClient) StreamWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// 首先尝试根据模型名称推断 provider
	if providerName := c.inferProviderFromModel(model); providerName != "" {
		if provider, exists := c.providers[providerName]; exists {
			// 检查该 provider 是否配置了 API key
			if c.IsProviderConfigured(providerName) {
				fmt.Printf("Using %s provider for model: %s\n", providerName, model)
				if imagePath != "" {
					return provider.StreamWithImage(model, message, imagePath, temperature, maxTokens, timeout)
				}
				return provider.Stream(model, message, temperature, maxTokens, timeout)
			} else {
				return fmt.Errorf("provider '%s' is not configured for model '%s'. Please configure the API key first", providerName, model)
			}
		}
	}

	// 如果推断不出来，返回错误并提示用户明确指定 provider
	return fmt.Errorf("cannot determine provider for model '%s'. Please specify provider explicitly:\n"+
		"  sse bailian %s \"your message\"     # for Qwen models\n"+
		"  sse openai %s \"your message\"      # for GPT models\n"+
		"  sse google %s \"your message\"      # for Gemini models\n"+
		"  sse anthropic %s \"your message\"   # for Claude models\n"+
		"Or use a recognizable model name like: qwen-max, gpt-4o, gemini-2.5-pro, claude-3-5-sonnet-20241022",
		model, model, model, model, model)
}

// GetFullResponseWithProvider 获取完整的AI响应（非流式）
func (c *SSEClient) GetFullResponseWithProvider(providerName, model, message, imagePath string, temperature float64, maxTokens, timeout int) (string, error) {
	// 如果明确指定了 provider
	if providerName != "" {
		return c.getFullResponseWithSpecificProvider(providerName, model, message, imagePath, temperature, maxTokens, timeout)
	}

	// 如果没有指定 provider，使用自动推断逻辑
	return c.GetFullResponseAuto(model, message, imagePath, temperature, maxTokens, timeout)
}

// getFullResponseWithSpecificProvider 使用指定的 provider 获取完整响应
func (c *SSEClient) getFullResponseWithSpecificProvider(providerName, model, message, imagePath string, temperature float64, maxTokens, timeout int) (string, error) {
	// 检查 provider 是否存在
	provider, exists := c.providers[providerName]
	if !exists {
		return "", fmt.Errorf("provider not found | 提供商未找到: %s\nAvailable providers | 可用提供商: bailian, openai, google, anthropic, deepseek", providerName)
	}

	// 检查 provider 是否配置了 API key
	if !c.IsProviderConfigured(providerName) {
		return "", fmt.Errorf("provider '%s' is not configured. Please configure the API key first", providerName)
	}

	if imagePath != "" {
		return provider.GetFullResponseWithImage(model, message, imagePath, temperature, maxTokens, timeout)
	}
	return provider.GetFullResponse(model, message, temperature, maxTokens, timeout)
}

// GetFullResponseAuto 自动推断 provider 并获取完整响应
func (c *SSEClient) GetFullResponseAuto(model, message, imagePath string, temperature float64, maxTokens, timeout int) (string, error) {
	// 首先尝试根据模型名称推断 provider
	if providerName := c.inferProviderFromModel(model); providerName != "" {
		if provider, exists := c.providers[providerName]; exists {
			// 检查该 provider 是否配置了 API key
			if c.IsProviderConfigured(providerName) {
				if imagePath != "" {
					return provider.GetFullResponseWithImage(model, message, imagePath, temperature, maxTokens, timeout)
				}
				return provider.GetFullResponse(model, message, temperature, maxTokens, timeout)
			} else {
				return "", fmt.Errorf("provider '%s' is not configured for model '%s'. Please configure the API key first", providerName, model)
			}
		}
	}

	// 如果推断不出来，返回错误并提示用户明确指定 provider
	return "", fmt.Errorf("cannot determine provider for model '%s'. Please specify provider explicitly:\n"+
		"  sse bailian %s \"your message\"     # for Qwen models\n"+
		"  sse openai %s \"your message\"      # for GPT models\n"+
		"  sse google %s \"your message\"      # for Gemini models\n"+
		"  sse anthropic %s \"your message\"   # for Claude models\n"+
		"Or use a recognizable model name like: qwen-max, gpt-4o, gemini-2.5-pro, claude-3-5-sonnet-20241022",
		model, model, model, model, model)
}
