package providers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AnthropicProvider struct{}

type AnthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []AnthropicMessage `json:"messages"`
	Stream    bool               `json:"stream"`
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	Type  string `json:"type"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{}
}

func (p *AnthropicProvider) SupportsModel(model string) bool {
	models := GetModelsForProvider("anthropic")
	return ModelInList(model, models)
}

func (p *AnthropicProvider) Stream(model, message string, temperature float64, maxTokens, timeout int) error {
	cfg, exists := GetProviderConfig("anthropic")
	if !exists {
		return GetProviderNotConfiguredError("anthropic")
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	if apiKey == "" {
		return GetAPIKeyConfigError("anthropic")
	}

	req := AnthropicRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages: []AnthropicMessage{
			{Role: "user", Content: message},
		},
		Stream: true,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s", string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var response AnthropicResponse
			if err := json.Unmarshal([]byte(data), &response); err == nil {
				if response.Type == "content_block_delta" && response.Delta.Type == "text_delta" {
					fmt.Print(response.Delta.Text)
				}
			}
		}
	}

	fmt.Println()
	return scanner.Err()
}
func (p *AnthropicProvider) StreamWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// Anthropic 的图片支持暂未实现
	if imagePath != "" {
		return fmt.Errorf("image support for Anthropic models not implemented yet")
	}
	return p.Stream(model, message, temperature, maxTokens, timeout)
}

// GetFullResponse 获取完整的AI响应（非流式）
func (p *AnthropicProvider) GetFullResponse(model, message string, temperature float64, maxTokens, timeout int) (string, error) {
	cfg, exists := GetProviderConfig("anthropic")
	if !exists {
		return "", GetProviderNotConfiguredError("anthropic")
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	if apiKey == "" {
		return "", GetAPIKeyConfigError("anthropic")
	}

	req := AnthropicRequest{
		Model: model,
		Messages: []AnthropicMessage{
			{Role: "user", Content: message},
		},
		MaxTokens: maxTokens,
		Stream:    true,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var fullResponse strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var response AnthropicResponse
			if err := json.Unmarshal([]byte(data), &response); err == nil {
				if response.Delta.Text != "" {
					fullResponse.WriteString(response.Delta.Text)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return fullResponse.String(), nil
}

// GetFullResponseWithImage 获取完整的AI响应（非流式，支持图片）
func (p *AnthropicProvider) GetFullResponseWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) (string, error) {
	// Anthropic 的图片支持暂未实现
	if imagePath != "" {
		return "", fmt.Errorf("image support for Anthropic models not implemented yet")
	}
	return p.GetFullResponse(model, message, temperature, maxTokens, timeout)
}
