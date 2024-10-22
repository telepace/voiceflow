package openai

import (
    "github.com/telepace/voiceflow/internal/config"
	
)

type OpenAILLM struct {
    apiKey string
}

func NewOpenAILLM() *OpenAILLM {
    cfg := config.GetConfig()
    return &OpenAILLM{
        apiKey: cfg.OpenAI.APIKey,
    }
}

func (o *OpenAILLM) GetResponse(prompt string) (string, error) {
    // 调用 OpenAI API，获取模型的回复
}