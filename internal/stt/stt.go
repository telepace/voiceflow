// stt.go
package stt

import (
	assemblyai "github.com/telepace/voiceflow/internal/stt/assemblyai"
	aaiws "github.com/telepace/voiceflow/internal/stt/assemblyai-ws"
	"github.com/telepace/voiceflow/internal/stt/azure"
	"github.com/telepace/voiceflow/internal/stt/google"
	"github.com/telepace/voiceflow/internal/stt/local"
	"github.com/telepace/voiceflow/internal/stt/volcengine"
	"github.com/telepace/voiceflow/pkg/logger"
)

// Service 定义了 STT 服务的接口
type Service interface {
	Recognize(audioData []byte, audioURL string) (string, error) // 接收音频数据，返回文本
}

// NewService 根据配置返回相应的 STT 服务实现
func NewService(provider string) Service {
	logger.Debugf("Using STT provider: %s", provider)
	switch provider {
	case "azure":
		return azure.NewAzureSTT()
	case "google":
		return google.NewGoogleSTT()
	case "assemblyai-ws":
		return aaiws.NewAssemblyAI()
	case "volcengine":
		return volcengine.NewVolcengineSTT()
	case "local":
		return local.NewLocalSTT()
	case "assemblyai":
		return assemblyai.NewAssemblyAI()
	default:
		return local.NewLocalSTT()
	}
}
