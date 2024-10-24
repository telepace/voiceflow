package llm
import (
    "github.com/telepace/voiceflow/internal/llm/local"
    "github.com/telepace/voiceflow/internal/llm/openai"
)

type Service interface {
    GetResponse(prompt string) (string, error)
}

func NewService(provider string) Service {
    switch provider {
    case "openai":
        return openai.NewOpenAILLM()
    case "local":
        return local.NewLocalLLM()
    default:
        return local.NewLocalLLM()
    }
}