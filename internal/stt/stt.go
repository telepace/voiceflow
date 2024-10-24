package stt

import (
	"github.com/telepace/voiceflow/internal/stt/azure"
	"github.com/telepace/voiceflow/internal/stt/google"
	"github.com/telepace/voiceflow/internal/stt/local"
)

// Service 定义了 STT 服务的接口
type Service interface {
	Recognize(audioData []byte) (string, error) // 接收音频数据，返回文本
}

// NewService 根据配置返回相应的 STT 服务实现
func NewService(provider string) Service {
	switch provider {
	case "azure":
		return azure.NewAzureSTT() // Azure STT 实现
	case "google":
		return google.NewGoogleSTT() // Google STT 实现
	case "local":
		return local.NewLocalSTT() // 本地 STT 实现
	default:
		return local.NewLocalSTT() // 如果未设置，默认为本地实现
	}
}
