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

type GoogleProvider struct{}

type GoogleRequest struct {
	Contents         []GoogleContent        `json:"contents"`
	GenerationConfig GoogleGenerationConfig `json:"generationConfig"`
}

type GoogleContent struct {
	Parts []GooglePart `json:"parts"`
}

type GooglePart struct {
	Text string `json:"text"`
}

type GoogleGenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type GoogleResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
}

func NewGoogleProvider() *GoogleProvider {
	return &GoogleProvider{}
}

func (p *GoogleProvider) SupportsModel(model string) bool {
	models := GetModelsForProvider("google")
	return ModelInList(model, models)
}

func (p *GoogleProvider) Stream(model, message string, temperature float64, maxTokens, timeout int) error {
	cfg, exists := GetProviderConfig("google")
	if !exists {
		return GetProviderNotConfiguredError("google")
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	if apiKey == "" {
		return GetAPIKeyConfigError("google")
	}

	// 修正 URL 格式 - Google Gemini API 使用不同的端点
	url := fmt.Sprintf("%s/%s:streamGenerateContent?alt=sse&key=%s", baseURL, model, apiKey)

	req := GoogleRequest{
		Contents: []GoogleContent{
			{
				Parts: []GooglePart{
					{Text: message},
				},
			},
		},
		GenerationConfig: GoogleGenerationConfig{
			Temperature:     temperature,
			MaxOutputTokens: maxTokens,
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Google API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 检查响应的 Content-Type 来判断是否为流式响应
	contentType := resp.Header.Get("Content-Type")

	if strings.Contains(contentType, "text/event-stream") || strings.Contains(contentType, "application/x-ndjson") {
		// 真正的流式响应处理
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			// 跳过空行
			if line == "" {
				continue
			}

			// 处理 SSE 格式
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// 跳过结束标记
				if data == "[DONE]" || data == "" {
					break
				}

				var response GoogleResponse
				if err := json.Unmarshal([]byte(data), &response); err == nil {
					if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
						text := response.Candidates[0].Content.Parts[0].Text
						fmt.Print(text)

						// 检查是否完成
						if response.Candidates[0].FinishReason != "" && response.Candidates[0].FinishReason != "STOP" {
							break
						}
					}
				}
			} else {
				// 尝试直接解析 JSON 行
				var response GoogleResponse
				if err := json.Unmarshal([]byte(line), &response); err == nil {
					if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
						text := response.Candidates[0].Content.Parts[0].Text
						fmt.Print(text)
					}
				}
			}
		}
		fmt.Println()
		return scanner.Err()
	} else {
		// 非流式响应，一次性读取完整响应
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var response GoogleResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return fmt.Errorf("failed to parse response: %v", err)
		}

		if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
			text := response.Candidates[0].Content.Parts[0].Text

			// 模拟流式输出效果
			for _, char := range text {
				fmt.Print(string(char))
				time.Sleep(10 * time.Millisecond) // 模拟打字机效果
			}
			fmt.Println()
		}

		return nil
	}
}
func (p *GoogleProvider) StreamWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) error {
	// Google 的图片支持暂未实现
	if imagePath != "" {
		return fmt.Errorf("image support for Google models not implemented yet")
	}
	return p.Stream(model, message, temperature, maxTokens, timeout)
}

// GetFullResponse 获取完整的AI响应（非流式）
func (p *GoogleProvider) GetFullResponse(model, message string, temperature float64, maxTokens, timeout int) (string, error) {
	cfg, exists := GetProviderConfig("google")
	if !exists {
		return "", GetProviderNotConfiguredError("google")
	}

	baseURL := cfg.BaseURL
	apiKey := cfg.APIKey

	if apiKey == "" {
		return "", GetAPIKeyConfigError("google")
	}

	req := GoogleRequest{
		Contents: []GoogleContent{
			{
				Parts: []GooglePart{
					{Text: message},
				},
			},
		},
		GenerationConfig: GoogleGenerationConfig{
			Temperature:     temperature,
			MaxOutputTokens: maxTokens,
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", baseURL, model, apiKey)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")

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
			if data == "" {
				continue
			}

			var response GoogleResponse
			if err := json.Unmarshal([]byte(data), &response); err == nil {
				if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
					fullResponse.WriteString(response.Candidates[0].Content.Parts[0].Text)
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
func (p *GoogleProvider) GetFullResponseWithImage(model, message, imagePath string, temperature float64, maxTokens, timeout int) (string, error) {
	// Google 的图片支持暂未实现
	if imagePath != "" {
		return "", fmt.Errorf("image support for Google models not implemented yet")
	}
	return p.GetFullResponse(model, message, temperature, maxTokens, timeout)
}
