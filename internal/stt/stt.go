package stt

type Service interface {
    Recognize(audioData []byte) (string, error)
}

func NewService() Service {
    // 根据配置返回相应的实现
    cfg := config.GetConfig()
    switch cfg.STT.Provider {
    case "azure":
        return azure.NewAzureSTT()
    case "google":
        return google.NewGoogleSTT()
    case "local":
        return local.NewLocalSTT()
    default:
        return local.NewLocalSTT()
    }
}