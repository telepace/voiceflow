package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
)

type STT struct {
	apiKey   string
	region   string
	endpoint string
}

// NewAzureSTT 创建并返回一个新的 AzureSTT 实例
func NewAzureSTT() *STT {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("配置初始化失败: %v", err)
	}
	return &STT{
		apiKey:   cfg.Azure.STTKey,
		region:   cfg.Azure.Region,
		endpoint: fmt.Sprintf("https://%s.stt.speech.microsoft.com/speech/recognition/conversation/cognitiveservices/v1", cfg.Azure.Region),
	}
}

// Recognize 调用 Azure 的 STT API 将音频数据转换为文本
// 新增 audioURL 参数，但 Azure 不使用该参数
func (a *STT) Recognize(audioData []byte, audioURL string) (string, error) {
	if audioURL != "" {
		logger.Infof("Azure STT 不支持使用 audioURL，忽略该参数")
	}

	req, err := http.NewRequest("POST", a.endpoint, bytes.NewReader(audioData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", a.apiKey)
	req.Header.Set("Content-Type", "audio/wav; codec=\"audio/pcm\"; samplerate=16000")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("Azure STT 错误: %s", string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 假设返回的结果中有识别文本（JSON 格式，需解析）
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	text, ok := result["DisplayText"].(string)
	if !ok {
		return "", fmt.Errorf("无法解析 Azure STT 的响应")
	}

	return text, nil
}
