package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
	"io"
	"net/http"
)

type GoogleTTS struct {
	apiKey string
	voice  string
	lang   string
}

// NewGoogleTTS 创建并返回一个新的 GoogleTTS 实例
func NewGoogleTTS() *GoogleTTS {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("配置初始化失败: %v", err)
	}
	return &GoogleTTS{
		apiKey: cfg.Google.TTSKey,
		voice:  "en-US-Wavenet-D", // 默认的 Google TTS 语音
		lang:   "en-US",
	}
}

// Synthesize 调用 Google TTS API 将文本转换为音频
func (g *GoogleTTS) Synthesize(text string) ([]byte, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"input": map[string]string{
			"text": text,
		},
		"voice": map[string]string{
			"languageCode": g.lang,
			"name":         g.voice,
		},
		"audioConfig": map[string]string{
			"audioEncoding": "LINEAR16", // 设置音频格式为 LINEAR16
		},
	})
	if err != nil {
		return nil, err
	}

	endpoint := "https://texttospeech.googleapis.com/v1/text:synthesize?key=" + g.apiKey

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google TTS error: %s", string(body))
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	audioData, err := io.ReadAll(bytes.NewReader([]byte(result["audioContent"])))
	if err != nil {
		return nil, err
	}

	return audioData, nil
}
