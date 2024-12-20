package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
)

// OpenAILLM 结构体存储 OpenAI 交互所需的信息
type OpenAILLM struct {
	apiKey   string
	endpoint string
}

// NewOpenAILLM 创建一个新的 OpenAILLM 实例
func NewOpenAILLM() *OpenAILLM {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("配置初始化失败: %v", err)
	}
	return &OpenAILLM{
		apiKey:   cfg.OpenAI.APIKey,
		endpoint: "https://api.openai.com/v1/completions", // 具体API路径
	}
}

// GetResponse 调用 OpenAI API，获取对话模型的回复
func (o *OpenAILLM) GetResponse(prompt string) (string, error) {
	if o.apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	url := o.endpoint

	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return result.Choices[0].Message.Content, nil
}
