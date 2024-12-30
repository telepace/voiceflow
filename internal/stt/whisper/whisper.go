package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
)

type WhisperSTT struct {
	apiKey      string
	endpoint    string
	model       string
	temperature float64
	vadModel    string
}

type WhisperResponse struct {
	Text     string  `json:"text"`
	Language string  `json:"language,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

func NewWhisperSTT() *WhisperSTT {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("配置初始化失败: %v", err)
	}

	return &WhisperSTT{
		apiKey:      cfg.Whisper.APIKey,
		endpoint:    cfg.Whisper.Endpoint,
		model:       cfg.Whisper.Model,
		temperature: cfg.Whisper.Temperature,
		vadModel:    cfg.Whisper.VADModel,
	}
}

func (w *WhisperSTT) Recognize(audioData []byte, audioURL string) (string, error) {
	// 创建一个 buffer 来写入 multipart 数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 写入音频文件
	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return "", fmt.Errorf("创建表单文件失败: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(audioData)); err != nil {
		return "", fmt.Errorf("写入音频数据失败: %v", err)
	}

	// 添加其他参数
	writer.WriteField("model", w.model)
	writer.WriteField("temperature", fmt.Sprintf("%f", w.temperature))
	writer.WriteField("vad_model", w.vadModel)

	// 关闭 multipart writer
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("关闭 writer 失败: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", w.endpoint, body)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.apiKey))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API 请求失败，状态码: %d，响应: %s",
			resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var result WhisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	logger.Infof("语音识别完成，语言: %s, 时长: %.2f秒",
		result.Language, result.Duration)

	return result.Text, nil
}

func (w *WhisperSTT) StreamRecognize(ctx context.Context, audioDataChan <-chan []byte,
	transcriptChan chan<- string) error {
	// Whisper V3 Turbo 目前不支持流式处理
	return fmt.Errorf("Whisper V3 Turbo 不支持流式处理")
}
