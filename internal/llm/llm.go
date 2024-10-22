package llm

type Service interface {
    GetResponse(prompt string) (string, error)
}

func NewService(provider string) Service {
    switch provider {
    case "openai":
        return NewOpenAILLM()
    case "local":
        return NewLocalLLM()
    default:
        return NewLocalLLM()
    }
}