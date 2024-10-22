package tts

type Service interface {
    Synthesize(text string) ([]byte, error)
}

func NewService(provider string) Service {
    switch provider {
    case "azure":
        return NewAzureTTS()
    case "google":
        return NewGoogleTTS()
    case "local":
        return NewLocalTTS()
    default:
        return NewLocalTTS()
    }
}