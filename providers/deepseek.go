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

type DeepSeekProvider struct{}

type DeepSeekRequest struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	Temperature float64           `json:"temperature"`
	MaxTokens   int               `json:"max_tokens"`
	Stream      bool              `json:"stream"`
}

type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func NewDeepSeekProvider() *DeepSeekProvider {
	return &DeepSeekProvider{}
}

func (p *DeepSeekProvider) SupportsModel(model string) bool {
	models := GetModelsForProvider("deepseek")
	for _, m := range models {
		if m == model {
			return true
		}
	}
	return false
}

func (p *DeepSeekProvider) Stream(model, message string, temperature float64, maxTokens, timeout int) error {
	return p.StreamWithImage(model, message, "", temperature, maxTokens, timeout)
}

func (p *DeepSeekProvider) StreamWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	config := GetConfig()
	providerConfig, exists := config.Providers["deepseek"]
	if !exists {
		return fmt.Errorf("deepseek provider not configured")
	}

	if imagePath != "" {
		return fmt.Errorf("deepseek provider does not support image input yet")
	}

	reqBody := DeepSeekRequest{
		Model: model,
		Messages: []DeepSeekMessage{
			{
				Role:    "user",
				Content: message,
			},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", providerConfig.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+providerConfig.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var response DeepSeekResponse
			if err := json.Unmarshal([]byte(data), &response); err != nil {
				continue
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				if content != "" {
					fmt.Print(content)
				}
			}
		}
	}

	fmt.Println() // 添加换行
	return scanner.Err()
}

func (p *DeepSeekProvider) GetFullResponse(model, message string, temperature float64, maxTokens, timeout int) (string, error) {
	return p.GetFullResponseWithImage(model, message, "", temperature, maxTokens, timeout)
}

func (p *DeepSeekProvider) GetFullResponseWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) (string, error) {
	config := GetConfig()
	providerConfig, exists := config.Providers["deepseek"]
	if !exists {
		return "", fmt.Errorf("deepseek provider not configured")
	}

	if imagePath != "" {
		return "", fmt.Errorf("deepseek provider does not support image input yet")
	}

	reqBody := DeepSeekRequest{
		Model: model,
		Messages: []DeepSeekMessage{
			{
				Role:    "user",
				Content: message,
			},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", providerConfig.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+providerConfig.APIKey)

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response content")
}
