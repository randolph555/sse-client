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

type BailianProvider struct{}

type BailianRequest struct {
	Model       string           `json:"model"`
	Messages    []BailianMessage `json:"messages"`
	MaxTokens   int              `json:"max_tokens"`
	Temperature float64          `json:"temperature"`
	Stream      bool             `json:"stream"`
}

type BailianMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type BailianTextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type BailianImageContent struct {
	Type     string `json:"type"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

type BailianResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func NewBailianProvider() *BailianProvider {
	return &BailianProvider{}
}

func (p *BailianProvider) SupportsModel(model string) bool {
	models := GetModelsForProvider("bailian")
	return ModelInList(model, models)
}

func (p *BailianProvider) Stream(model, message string, temperature float64, maxTokens, timeout int) error {
	cfg, exists := GetProviderConfig("bailian")
	if !exists {
		return GetProviderNotConfiguredError("bailian")
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	if apiKey == "" {
		return GetAPIKeyConfigError("bailian")
	}

	req := BailianRequest{
		Model: model,
		Messages: []BailianMessage{
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

			var response BailianResponse
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

func (p *BailianProvider) StreamWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	cfg, exists := GetProviderConfig("bailian")
	if !exists {
		return GetProviderNotConfiguredError("bailian")
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	if apiKey == "" {
		return GetAPIKeyConfigError("bailian")
	}

	// 构建消息内容
	var messages []BailianMessage

	if imagePath != "" {
		// 编码图片
		encodedImage, mimeType, err := EncodeImageToBase64(imagePath)
		if err != nil {
			return fmt.Errorf("failed to encode image: %v", err)
		}

		// 构建多模态消息
		content := []interface{}{
			BailianTextContent{
				Type: "text",
				Text: message,
			},
			BailianImageContent{
				Type: "image_url",
				ImageURL: struct {
					URL string `json:"url"`
				}{
					URL: fmt.Sprintf("data:%s;base64,%s", mimeType, encodedImage),
				},
			},
		}
		messages = []BailianMessage{
			{Role: "user", Content: content},
		}
	} else {
		// 纯文本消息
		messages = []BailianMessage{
			{Role: "user", Content: message},
		}
	}

	req := BailianRequest{
		Model:       model,
		Messages:    messages,
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

			var response BailianResponse
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

// GetFullResponse 获取完整的AI响应（非流式）
func (p *BailianProvider) GetFullResponse(model, message string, temperature float64, maxTokens, timeout int) (string, error) {
	cfg, exists := GetProviderConfig("bailian")
	if !exists {
		return "", GetProviderNotConfiguredError("bailian")
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	if apiKey == "" {
		return "", GetAPIKeyConfigError("bailian")
	}

	req := BailianRequest{
		Model: model,
		Messages: []BailianMessage{
			{Role: "user", Content: message},
		},
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      true, // 仍然使用流式，但收集所有内容
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

			var response BailianResponse
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
func (p *BailianProvider) GetFullResponseWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) (string, error) {
	cfg, exists := GetProviderConfig("bailian")
	if !exists {
		return "", GetProviderNotConfiguredError("bailian")
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	if apiKey == "" {
		return "", GetAPIKeyConfigError("bailian")
	}

	// 构建包含图片的消息内容
	var content []interface{}
	content = append(content, BailianTextContent{Type: "text", Text: message})

	if imagePath != "" {
		imageURL, err := ConvertImageToBase64URL(imagePath)
		if err != nil {
			return "", fmt.Errorf("failed to process image: %v", err)
		}
		content = append(content, BailianImageContent{
			Type: "image_url",
			ImageURL: struct {
				URL string `json:"url"`
			}{URL: imageURL},
		})
	}

	req := BailianRequest{
		Model: model,
		Messages: []BailianMessage{
			{Role: "user", Content: content},
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

			var response BailianResponse
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
