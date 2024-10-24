package storage

import "github.com/telepace/voiceflow/internal/config"

type Service interface {
	StoreAudio(audioData []byte) (string, error) // 存储音频并返回 URL 或路径
}

// NewService 根据配置返回相应的存储服务
func NewService() Service {
	cfg := config.GetConfig()
	if cfg.MinIO.Enabled {
		return NewMinIOService() // 返回 MinIO 存储实现
	}
	return NewLocalStorageService() // 返回本地存储实现
}
