// internal/tts/tts.go

package tts

import (
	"github.com/telepace/voiceflow/internal/tts/azure"
	"github.com/telepace/voiceflow/internal/tts/google"
	"github.com/telepace/voiceflow/internal/tts/local"
	"github.com/telepace/voiceflow/pkg/logger"
)

// Service 定义了 TTS 服务的通用接口
type Service interface {
	Synthesize(text string) ([]byte, error) // 将文本合成为音频数据
}

// NewService 根据配置返回相应的 TTS 服务实现
func NewService(provider string) Service {
	logger.Debugf("Using TTS provider: %s", provider)
	switch provider {
	case "azure":
		return azure.NewAzureTTS() // 调用 Azure TTS 实现
	case "google":
		return google.NewGoogleTTS() // 调用 Google TTS 实现
	case "local":
		return local.NewLocalTTS() // 调用本地 TTS 实现
	default:
		return local.NewLocalTTS() // 默认使用本地 TTS
	}
}
