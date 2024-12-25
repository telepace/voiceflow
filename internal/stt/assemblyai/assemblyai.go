package assemblyai

import (
	"bytes"
	"context"
	"fmt"
	"time"

	aai "github.com/AssemblyAI/assemblyai-go-sdk"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
)

type STT struct {
	client *aai.Client
}

// NewAssemblyAI 创建并返回一个新的 AssemblyAI STT 实例
func NewAssemblyAI() *STT {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("配置初始化失败: %v", err)
	}

	client := aai.NewClient(cfg.AssemblyAI.APIKey)
	return &STT{
		client: client,
	}
}

// Recognize 实现了 stt.Service 接口，使用 AssemblyAI 进行语音识别
func (s *STT) Recognize(audioData []byte, audioURL string) (string, error) {
	if audioURL != "" {
		// 使用提供的 audioURL 调用 AssemblyAI 的转录服务
		return s.transcribeFromURL(audioURL)
	}

	// 原有的处理流程，直接使用 audioData
	return s.transcribeFromData(audioData)
}

func (s *STT) transcribeFromURL(audioURL string) (string, error) {
	ctx := context.Background()

	// 首次尝试：启用语言检测
	params := &aai.TranscriptOptionalParams{
		LanguageDetection:           aai.Bool(true),
		LanguageConfidenceThreshold: aai.Float64(0.5),
		FormatText:                  aai.Bool(true),
		Punctuate:                   aai.Bool(true),
	}

	transcript, err := s.client.Transcripts.TranscribeFromURL(ctx, audioURL, params)
	if err != nil {
		// 检查是否是语言置信度错误
		if transcript.Status == "error" && transcript.Error != nil &&
			*transcript.Error == "detected language 'en', confidence 0.6602, is below the requested confidence threshold value of '0.8'" {

			// 第二次尝试：使用英语作为默认语言
			params.LanguageDetection = aai.Bool(false)
			params.LanguageCode = aai.TranscriptLanguageCode("en")

			transcript, err = s.client.Transcripts.TranscribeFromURL(ctx, audioURL, params)
			if err != nil {
				return "", fmt.Errorf("使用默认语言重新转录失败: %v", err)
			}
		} else {
			return "", fmt.Errorf("转录失败: %v", err)
		}
	}

	// 使用指数退避策略
	backoff := 100 * time.Millisecond
	maxBackoff := 2 * time.Second

	for transcript.Status != "completed" {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("转录超时")
		default:
			time.Sleep(backoff)

			transcript, err = s.client.Transcripts.Get(ctx, *transcript.ID)
			if err != nil {
				return "", fmt.Errorf("获取转录结果失败: %v", err)
			}

			if transcript.Status == "error" {
				return "", fmt.Errorf("转录出错: %s", *transcript.Error)
			}

			// 增加等待时间，但不超过最大值
			backoff = time.Duration(float64(backoff) * 1.5)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	if transcript.Text == nil {
		return "", fmt.Errorf("转录结果为空")
	}

	return *transcript.Text, nil
}

func (s *STT) transcribeFromData(audioData []byte) (string, error) {
	ctx := context.Background()

	// 首先上传音频数据
	upload, err := s.client.Upload(ctx, bytes.NewReader(audioData))
	if err != nil {
		return "", fmt.Errorf("上传音频数据失败: %v", err)
	}

	// 使用上传后的 URL 进行转录
	return s.transcribeFromURL(upload)
}

// StreamRecognize 实现实时转录接口
func (s *STT) StreamRecognize(ctx context.Context, audioDataChan <-chan []byte, transcriptChan chan<- string) error {
	// AssemblyAI 目前不支持直接的流式处理
	// 这里我们可以实现一个简单的缓冲处理
	return fmt.Errorf("AssemblyAI 不支持流式处理")
}
