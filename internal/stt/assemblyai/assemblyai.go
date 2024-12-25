package assemblyai

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	aai "github.com/AssemblyAI/assemblyai-go-sdk"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
)

type STT struct {
	client *aai.Client
	cfg    *config.Config
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
		cfg:    cfg,
	}
}

// Recognize 实现了 stt.Service 接口，使用 AssemblyAI 进行语音识别
func (s *STT) Recognize(audioData []byte, audioURL string) (string, error) {
	if audioURL != "" {
		// 使用提供的 audioURL 调用 AssemblyAI 的转录服务
		return s.transcribeFromURL(audioURL)
	}

	// 原有的处理流程，直接使用音频数据
	return s.transcribeFromData(audioData)
}

func (s *STT) transcribeFromURL(audioURL string) (string, error) {
	ctx := context.Background()

	// 第一次尝试：启用语言检测
	params := &aai.TranscriptOptionalParams{
		LanguageDetection:           aai.Bool(true),
		LanguageConfidenceThreshold: aai.Float64(0.1), // 设置较低的初始阈值
		Punctuate:                   aai.Bool(s.cfg.AssemblyAI.Punctuate),
		FormatText:                  aai.Bool(s.cfg.AssemblyAI.FormatText),
		SpeechThreshold:             aai.Float64(s.cfg.AssemblyAI.SpeechThreshold),
		Multichannel:                aai.Bool(s.cfg.AssemblyAI.Multichannel),
	}

	transcript, err := s.client.Transcripts.TranscribeFromURL(ctx, audioURL, params)
	if err != nil {
		// 检查是否是语言置信度错误
		if s.isLanguageConfidenceError(err) && s.cfg.AssemblyAI.DefaultLanguageCode != "" {
			logger.Infof("第一次尝试失败(语言置信度低), 使用默认语言 %s 重试",
				s.cfg.AssemblyAI.DefaultLanguageCode)

			// 第二次尝试：禁用语言检测，使用固定语言
			retryParams := &aai.TranscriptOptionalParams{
				LanguageDetection: aai.Bool(false), // 明确禁用语言检测
				LanguageCode:      aai.TranscriptLanguageCode(s.cfg.AssemblyAI.DefaultLanguageCode),
				// 基础参数
				Punctuate:       aai.Bool(s.cfg.AssemblyAI.Punctuate),
				FormatText:      aai.Bool(s.cfg.AssemblyAI.FormatText),
				SpeechThreshold: aai.Float64(s.cfg.AssemblyAI.SpeechThreshold),
				Multichannel:    aai.Bool(s.cfg.AssemblyAI.Multichannel),
				// 不再设置 LanguageConfidenceThreshold
			}

			// 记录重试请求参数
			logger.Debugf("重试请求参数: %+v", retryParams)

			transcript, err = s.client.Transcripts.TranscribeFromURL(ctx, audioURL, retryParams)
			if err != nil {
				return "", fmt.Errorf("使用默认语言 %s 重试失败: %v",
					s.cfg.AssemblyAI.DefaultLanguageCode, err)
			}
		} else {
			return "", fmt.Errorf("转录请求失败: %v", err)
		}
	}

	// 使用指数退避策略，轮询转录状态
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
				if transcript.Error != nil {
					return "", fmt.Errorf("转录出错: %s", *transcript.Error)
				}
				return "", fmt.Errorf("转录出错, 未返回具体错误信息")
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

	// 先上传音频数据
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
	// 这里你可以实现一个间歇式合并音频然后再调用 TranscribeFromData 的逻辑
	// 或者直接返回错误，以示暂不支持流式
	return fmt.Errorf("AssemblyAI 不支持流式处理")
}

// buildParams 将 config.yaml 中的字段映射到 AssemblyAI 的 TranscriptOptionalParams
func (s *STT) buildParams() *aai.TranscriptOptionalParams {
	aaiCfg := s.cfg.AssemblyAI

	params := &aai.TranscriptOptionalParams{
		SpeechModel:     aai.SpeechModel(aaiCfg.Model),
		Punctuate:       aai.Bool(aaiCfg.Punctuate),
		FormatText:      aai.Bool(aaiCfg.FormatText),
		SpeechThreshold: aai.Float64(aaiCfg.SpeechThreshold),
		Multichannel:    aai.Bool(aaiCfg.Multichannel),
	}

	// 词汇增强设置
	if len(aaiCfg.WordBoost) > 0 {
		params.WordBoost = aaiCfg.WordBoost
		params.BoostParam = aai.TranscriptBoostParam(aaiCfg.BoostParam)
	}

	// 自定义拼写设置
	if len(aaiCfg.CustomSpelling) > 0 {
		var customSpellings []aai.TranscriptCustomSpelling
		for _, cs := range aaiCfg.CustomSpelling {
			customSpellings = append(customSpellings, aai.TranscriptCustomSpelling{
				From: cs.From,
				To:   aai.String(cs.To),
			})
		}
		params.CustomSpelling = customSpellings
	}

	return params
}

// isLanguageConfidenceError 优化错误检测逻辑
func (s *STT) isLanguageConfidenceError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "below the requested confidence threshold") ||
		strings.Contains(errMsg, "confidence threshold value")
}
