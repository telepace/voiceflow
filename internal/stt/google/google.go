package google

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
	"io/ioutil"
	"net/http"
)

type GoogleSTT struct {
	apiKey string
}

// NewGoogleSTT 创建并返回一个新的 GoogleSTT 实例
func NewGoogleSTT() *GoogleSTT {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("配置初始化失败: %v", err)
	}
	return &GoogleSTT{
		apiKey: cfg.Google.STTKey,
	}
}

// Recognize 调用 Google STT API 将音频数据转换为文本
func (g *GoogleSTT) Recognize(audioData []byte) (string, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"config": map[string]interface{}{
			"encoding":        "LINEAR16",
			"sampleRateHertz": 16000,
			"languageCode":    "en-US",
		},
		"audio": map[string]string{
			"content": base64.StdEncoding.EncodeToString(audioData),
		},
	})
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("https://speech.googleapis.com/v1/speech:recognize?key=%s", g.apiKey)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("google STT error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["results"].([]interface{})[0].(map[string]interface{})["alternatives"].([]interface{})[0].(map[string]interface{})["transcript"].(string), nil
}
