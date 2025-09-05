package providers

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

// ProviderConfig 结构体定义
type ProviderConfig struct {
	APIKey  string   `yaml:"api_key"`
	BaseURL string   `yaml:"base_url"`
	Models  []string `yaml:"models"`
}

// Config 结构体定义
type Config struct {
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// 全局配置变量
var config Config

// SetConfig 设置配置
func SetConfig(cfg Config) {
	config = cfg
}

// GetConfig 获取配置
func GetConfig() Config {
	return config
}

func GetModelsForProvider(providerName string) []string {
	if cfg, exists := GetProviderConfig(providerName); exists {
		return cfg.Models
	}
	return []string{}
}

func ModelInList(model string, models []string) bool {
	for _, m := range models {
		if strings.EqualFold(m, model) {
			return true
		}
	}
	return false
}

func GetProviderConfig(provider string) (ProviderConfig, bool) {
	cfg, exists := config.Providers[provider]
	return cfg, exists
}

// GetAPIKeyConfigError 返回带有配置指导的 API key 错误信息
func GetAPIKeyConfigError(providerName string) error {
	providerUpper := strings.ToUpper(providerName)
	return fmt.Errorf(`%s API key not configured

Please configure your API key using one of these methods:

Method 1: Environment variable (recommended)
  export %s_API_KEY="your-api-key-here"

Method 2: Config file
  Edit config.yaml or ~/.config/sse-client/config.yaml:
  providers:
    %s:
      api_key: "your-api-key-here"

Then try your command again.`, providerName, providerUpper, providerName)
}

// GetProviderNotConfiguredError 返回带有配置指导的 provider 未配置错误信息
func GetProviderNotConfiguredError(providerName string) error {
	providerUpper := strings.ToUpper(providerName)
	return fmt.Errorf(`%s provider not configured

Please configure the provider using one of these methods:

Method 1: Environment variables (recommended)
  export %s_API_KEY="your-api-key-here"
  export %s_BASE_URL="provider-base-url"  # optional

Method 2: Config file
  Edit config.yaml or ~/.config/sse-client/config.yaml:
  providers:
    %s:
      api_key: "your-api-key-here"
      base_url: "provider-base-url"  # optional

Then try your command again.`, providerName, providerUpper, providerUpper, providerName)
}

func EncodeImageToBase64(imagePath string) (string, string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read image file: %v", err)
	}

	// 获取 MIME 类型
	ext := filepath.Ext(imagePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// 默认类型
		switch ext {
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".webp":
			mimeType = "image/webp"
		default:
			mimeType = "image/jpeg"
		}
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, mimeType, nil
}

// ConvertImageToBase64URL 将图片转换为 base64 URL 格式
func ConvertImageToBase64URL(imagePath string) (string, error) {
	encoded, mimeType, err := EncodeImageToBase64(imagePath)
	if err != nil {
		return "", err
	}

	// 返回 data URL 格式
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}
