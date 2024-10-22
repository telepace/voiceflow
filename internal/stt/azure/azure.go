package azure

type AzureSTT struct {
	// Azure STT 所需的字段
}

func NewAzureSTT() *AzureSTT {
	return &AzureSTT{
		// 初始化
	}
}

func (a *AzureSTT) Recognize(audioData []byte) (string, error) {
	// 调用 Azure 的 API 进行语音识别
}
