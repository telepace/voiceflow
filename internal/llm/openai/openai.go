package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/telepace/voiceflow/internal/config"
)

// OpenAILLM 结构体存储 OpenAI 交互所需的信息
type OpenAILLM struct {
	apiKey   string
	endpoint string
}

// NewOpenAILLM 创建一个新的 OpenAILLM 实例
func NewOpenAILLM() *OpenAILLM {
	cfg := config.GetConfig()
	return &OpenAILLM{
		apiKey:   cfg.OpenAI.APIKey,
		endpoint: "https://api.openai.com/v1/completions", // 具体API路径
	}
}

// GetResponse 调用 OpenAI API，获取对话模型的回复
func (o *OpenAILLM) GetResponse(prompt string) (string, error) {
	// 定义请求体结构
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":      "text-davinci-003", // 或者其他模型
		"prompt":     prompt,
		"max_tokens": 150,
	})
	if err != nil {
		return "", err
	}

	// 创建请求
	req, err := http.NewRequest("POST", o.endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 处理响应
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get response from OpenAI")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// 返回模型生成的文本
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			return choice["text"].(string), nil
		}
	}

	return "", errors.New("invalid response format from OpenAI")
}
