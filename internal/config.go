package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Providers       map[string]ProviderConfig `yaml:"providers"`
	Timeout         int                       `yaml:"timeout"`
	MaxTokens       int                       `yaml:"max_tokens"`
	Temperature     float64                   `yaml:"temperature"`
	DefaultProvider string                    `yaml:"default_provider"`
	DefaultModel    string                    `yaml:"default_model"`
}

type ProviderConfig struct {
	BaseURL string   `yaml:"base_url"`
	APIKey  string   `yaml:"api_key"`
	Models  []string `yaml:"models"`
}

var config *Config

func loadConfig(configFile string) error {
	// 初始化默认配置
	config = &Config{
		Providers:   make(map[string]ProviderConfig),
		Timeout:     30,
		MaxTokens:   4096,
		Temperature: 0.7,
	}

	// 确定配置文件路径
	configPath := findConfigFile(configFile)

	// 尝试读取配置文件
	if configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			if err := yaml.Unmarshal(data, config); err != nil {
				return fmt.Errorf("failed to parse config file %s: %v", configPath, err)
			}
		}
	}

	// 从环境变量加载配置，覆盖文件配置
	loadFromEnvironment()

	return nil
}

// 查找配置文件
func findConfigFile(specifiedFile string) string {
	// 如果用户指定了配置文件，直接使用
	if specifiedFile != "" {
		if _, err := os.Stat(specifiedFile); err == nil {
			return specifiedFile
		}
		return ""
	}

	// 按优先级搜索配置文件
	searchPaths := []string{
		"./config.yaml", // 当前目录
		os.ExpandEnv("$HOME/.config/sse-client/config.yaml"), // 用户配置目录
		"/etc/sse-client/config.yaml",                        // 系统配置目录
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "" // 没有找到配置文件，将使用环境变量和默认值
}

// 从环境变量加载配置
func loadFromEnvironment() {
	providers := []string{"bailian", "openai", "google", "anthropic", "deepseek"}

	for _, provider := range providers {
		providerUpper := strings.ToUpper(provider)

		// 获取环境变量
		apiKey := os.Getenv(providerUpper + "_API_KEY")
		baseURL := os.Getenv(providerUpper + "_BASE_URL")

		// 如果环境变量存在，更新配置
		if apiKey != "" || baseURL != "" {
			if config.Providers == nil {
				config.Providers = make(map[string]ProviderConfig)
			}

			// 获取现有配置或创建新的
			existingConfig, exists := config.Providers[provider]
			if !exists {
				existingConfig = ProviderConfig{
					Models: getDefaultModels(provider),
				}
			}

			// 环境变量优先级更高，但保留现有的模型配置
			if apiKey != "" {
				existingConfig.APIKey = apiKey
			}
			if baseURL != "" {
				existingConfig.BaseURL = baseURL
			}

			// 只有在完全没有模型配置时才使用默认模型
			if len(existingConfig.Models) == 0 {
				existingConfig.Models = getDefaultModels(provider)
			}

			config.Providers[provider] = existingConfig
		}
	}
}

// 获取默认模型列表
func getDefaultModels(provider string) []string {
	switch provider {
	case "bailian":
		return []string{"qwen-max", "qwen-plus", "qwen-turbo", "qwen-vl-max", "qwen-vl-plus"}
	case "openai":
		return []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-3.5-turbo"}
	case "google":
		return []string{"gemini-2.5-flash", "gemini-1.5-flash", "gemini-2.5-pro", "gemini-1.5-pro"}
	case "anthropic":
		return []string{"claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022", "claude-3-opus-20240229"}
	case "deepseek":
		return []string{"deepseek-chat", "deepseek-coder", "deepseek-reasoner"}
	default:
		return []string{}
	}
}

func getProviderConfig(provider string) (ProviderConfig, bool) {
	cfg, exists := config.Providers[provider]
	return cfg, exists
}

func addModelToConfig(providerName, modelName string) error {
	// 确保配置已加载
	if config.Providers == nil {
		config.Providers = make(map[string]ProviderConfig)
	}

	// 获取或创建 provider 配置
	providerCfg, exists := config.Providers[providerName]
	if !exists {
		providerCfg = ProviderConfig{
			Models: []string{},
		}
	}

	// 检查模型是否已存在
	for _, model := range providerCfg.Models {
		if model == modelName {
			return fmt.Errorf("model '%s' already exists for provider '%s'", modelName, providerName)
		}
	}

	// 添加模型
	providerCfg.Models = append(providerCfg.Models, modelName)
	config.Providers[providerName] = providerCfg

	// 保存配置到文件
	return saveConfig()
}

func saveConfig() error {
	// 确定配置文件路径 - 使用与 loadConfig 相同的逻辑
	configPath := findConfigFile(appConfig.CfgFile)
	if configPath == "" {
		// 如果没有找到现有配置文件，使用默认路径
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot get home directory: %v", err)
		}
		configDir := filepath.Join(homeDir, ".config", "sse-client")
		configPath = filepath.Join(configDir, "config.yaml")

		// 确保目录存在
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("cannot create config directory: %v", err)
		}
	}

	// 将配置写入 YAML 文件
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %v", err)
	}

	return os.WriteFile(configPath, data, 0644)
}
func setDefaultProvider(provider, model string) error {
	// 确保配置已加载
	if config == nil {
		return fmt.Errorf("config not loaded")
	}

	// 验证 provider 是否存在且已配置
	if _, exists := config.Providers[provider]; !exists {
		return fmt.Errorf("provider '%s' is not configured", provider)
	}

	// 验证 model 是否在该 provider 的模型列表中
	providerCfg := config.Providers[provider]
	modelExists := false
	for _, m := range providerCfg.Models {
		if m == model {
			modelExists = true
			break
		}
	}

	if !modelExists {
		return fmt.Errorf("model '%s' is not available for provider '%s'", model, provider)
	}

	// 设置默认值
	config.DefaultProvider = provider
	config.DefaultModel = model

	// 保存配置
	return saveConfig()
}

func getDefaultProvider() (string, string) {
	if config == nil {
		return "", ""
	}
	return config.DefaultProvider, config.DefaultModel
}
