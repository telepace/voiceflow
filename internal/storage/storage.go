package storage

type Service interface {
    StoreAudio(audioData []byte) (string, error)
}

func NewService() Service {
    cfg := config.GetConfig()
    if cfg.MinIO.Enabled {
        return NewMinIOService()
    }
    // 可添加其他存储实现
    return NewLocalStorageService()
}