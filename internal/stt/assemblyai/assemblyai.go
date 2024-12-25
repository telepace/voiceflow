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

	// 第一次尝试：使用语言检测
	params := s.buildParams()
	transcript, err := s.client.Transcripts.TranscribeFromURL(ctx, audioURL, params)
	if err != nil {
		// 检查是否是语言置信度错误
		if s.isLanguageConfidenceError(err) && s.cfg.AssemblyAI.DefaultLanguageCode != "" {
			// 使用默认语言重试
			logger.Infof("语言置信度低于阈值 %.2f，使用默认语言 %s 重试",
				s.cfg.AssemblyAI.LanguageConfidenceThreshold,
				s.cfg.AssemblyAI.DefaultLanguageCode)

			// 构建新的参数，使用默认语言（禁用自动检测、去掉 threshold）
			params = s.buildParamsWithDefaultLanguage()
			transcript, err = s.client.Transcripts.TranscribeFromURL(ctx, audioURL, params)
			if err != nil {
				return "", fmt.Errorf("使用默认语言重试失败: %v", err)
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

// buildParams 将 config.yaml 中的字段映射到 AssemblyAI 的 TranscriptOptionalParams（第一次请求用）
func (s *STT) buildParams() *aai.TranscriptOptionalParams {
	aaiCfg := s.cfg.AssemblyAI

	params := &aai.TranscriptOptionalParams{
		// 将 string 转换为 SpeechModel 类型
		SpeechModel:                 aai.SpeechModel(aaiCfg.Model),
		LanguageDetection:           aai.Bool(aaiCfg.LanguageDetection),
		LanguageConfidenceThreshold: aai.Float64(aaiCfg.LanguageConfidenceThreshold),
		Punctuate:                   aai.Bool(aaiCfg.Punctuate),
		FormatText:                  aai.Bool(aaiCfg.FormatText),
		Disfluencies:                aai.Bool(aaiCfg.Disfluencies),
		FilterProfanity:             aai.Bool(aaiCfg.FilterProfanity),
		AudioStartFrom:              aai.Int64(aaiCfg.AudioStartFrom),
		AudioEndAt:                  aai.Int64(aaiCfg.AudioEndAt),
		SpeechThreshold:             aai.Float64(aaiCfg.SpeechThreshold),
		Multichannel:                aai.Bool(aaiCfg.Multichannel),
	}

	// 如果设置了固定的 language_code，则禁用语言检测并指定语言
	if aaiCfg.LanguageCode != "" {
		params.LanguageDetection = aai.Bool(false)
		params.LanguageCode = aai.TranscriptLanguageCode(aaiCfg.LanguageCode)
	}

	// 如果配置了词汇增强
	if len(aaiCfg.WordBoost) > 0 {
		params.WordBoost = aaiCfg.WordBoost
		// 将 string 转换为 TranscriptBoostParam 类型
		params.BoostParam = aai.TranscriptBoostParam(aaiCfg.BoostParam)
	}

	// 如果配置了自定义拼写
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

// 新增：检查是否是语言置信度错误
func (s *STT) isLanguageConfidenceError(err error) bool {
	return strings.Contains(err.Error(), "below the requested confidence threshold value")
}

// **优化后的关键点**：使用默认语言构建参数（禁用自动检测，不再带 threshold）
func (s *STT) buildParamsWithDefaultLanguage() *aai.TranscriptOptionalParams {
	// 直接手动指定，不再从 buildParams() 继承
	return &aai.TranscriptOptionalParams{
		LanguageDetection: aai.Bool(false),
		// 在这里写死你要使用的语言
		LanguageCode: aai.TranscriptLanguageCode(s.cfg.AssemblyAI.DefaultLanguageCode),

		// 以下可按需打开/关闭
		Punctuate:       aai.Bool(s.cfg.AssemblyAI.Punctuate),
		FormatText:      aai.Bool(s.cfg.AssemblyAI.FormatText),
		Disfluencies:    aai.Bool(s.cfg.AssemblyAI.Disfluencies),
		FilterProfanity: aai.Bool(s.cfg.AssemblyAI.FilterProfanity),

		// 如果想让二次请求也支持别的功能（词汇增强、自定义拼写等），
		// 也可以自行在这里加上。但注意不要再设 LanguageConfidenceThreshold。
	}
}
