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

type OpenAIProvider struct{}

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
	Stream      bool            `json:"stream"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{}
}

func (p *OpenAIProvider) SupportsModel(model string) bool {
	models := GetModelsForProvider("openai")
	return ModelInList(model, models)
}

func (p *OpenAIProvider) Stream(model, message string, temperature float64, maxTokens, timeout int) error {
	cfg, exists := GetProviderConfig("openai")
	if !exists {
		return GetProviderNotConfiguredError("openai")
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	if apiKey == "" {
		return GetAPIKeyConfigError("openai")
	}

	req := OpenAIRequest{
		Model: model,
		Messages: []OpenAIMessage{
			{Role: "user", Content: message},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      true,
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
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
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

			var response OpenAIResponse
			if err := json.Unmarshal([]byte(data), &response); err == nil {
				if len(response.Choices) > 0 {
					fmt.Print(response.Choices[0].Delta.Content)
					if response.Choices[0].FinishReason != nil {
						break
					}
				}
			}
		}
	}

	fmt.Println()
	return scanner.Err()
}
func (p *OpenAIProvider) StreamWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// OpenAI 的图片支持类似，但目前我们先简单实现为调用普通的 Stream 方法
	// 因为 OpenAI 的多模态 API 格式稍有不同
	if imagePath != "" {
		return fmt.Errorf("image support for OpenAI models not implemented yet")
	}
	return p.Stream(model, message, temperature, maxTokens, timeout)
}

// GetFullResponse 获取完整的AI响应（非流式）
func (p *OpenAIProvider) GetFullResponse(model, message string, temperature float64, maxTokens, timeout int) (string, error) {
	cfg, exists := GetProviderConfig("openai")
	if !exists {
		return "", GetProviderNotConfiguredError("openai")
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	if apiKey == "" {
		return "", GetAPIKeyConfigError("openai")
	}

	req := OpenAIRequest{
		Model: model,
		Messages: []OpenAIMessage{
			{Role: "user", Content: message},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      true,
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
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
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

			var response OpenAIResponse
			if err := json.Unmarshal([]byte(data), &response); err == nil {
				if len(response.Choices) > 0 {
					fullResponse.WriteString(response.Choices[0].Delta.Content)
					if response.Choices[0].FinishReason != nil {
						break
					}
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
func (p *OpenAIProvider) GetFullResponseWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) (string, error) {
	// OpenAI 的图片支持暂未实现
	if imagePath != "" {
		return "", fmt.Errorf("image support for OpenAI models not implemented yet")
	}
	return p.GetFullResponse(model, message, temperature, maxTokens, timeout)
}
