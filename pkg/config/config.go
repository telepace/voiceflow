package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

// 新增 AssemblyAIConfig，用于在 config.yaml 配置更多可调参数
type AssemblyAIConfig struct {
	APIKey                      string   `mapstructure:"api_key"`
	Model                       string   `mapstructure:"model"` // 可选 "best" 或 "nano"
	LanguageDetection           bool     `mapstructure:"language_detection"`
	LanguageConfidenceThreshold float64  `mapstructure:"language_confidence_threshold"`
	LanguageCode                string   `mapstructure:"language_code"`
	Punctuate                   bool     `mapstructure:"punctuate"`
	FormatText                  bool     `mapstructure:"format_text"`
	Disfluencies                bool     `mapstructure:"disfluencies"`
	FilterProfanity             bool     `mapstructure:"filter_profanity"`
	AudioStartFrom              int64    `mapstructure:"audio_start_from"`
	AudioEndAt                  int64    `mapstructure:"audio_end_at"`
	SpeechThreshold             float64  `mapstructure:"speech_threshold"`
	Multichannel                bool     `mapstructure:"multichannel"`
	BoostParam                  string   `mapstructure:"boost_param"`
	WordBoost                   []string `mapstructure:"word_boost"`
	// 这里仅演示一个简单示例的结构体，如果要自定义更多 mapping，可以创建单独结构体
	CustomSpelling []struct {
		From []string `mapstructure:"from"`
		To   string   `mapstructure:"to"`
	} `mapstructure:"custom_spelling"`
}

type VolcengineConfig struct {
	AccessKey  string `mapstructure:"access_key"`
	AppKey     string `mapstructure:"app_key"`
	WsURL      string `mapstructure:"ws_url"`
	ResourceID string `mapstructure:"resource_id"`
	UID        string `yaml:"uid"`
	Rate       int    `yaml:"rate"`
	Format     string `yaml:"format"`
	Bits       int    `yaml:"bits"`
	Channel    int    `yaml:"channel"`
	Codec      string `yaml:"codec"`
}

type AWSConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region          string `yaml:"region"`
}

type Config struct {
	Server struct {
		Port      int
		EnableTLS bool `mapstructure:"enable_tls"`
	}
	Web struct {
		Port int
	}
	STT struct {
		Provider string
	}
	TTS struct {
		Provider string
	}
	LLM struct {
		Provider string
	}
	AssemblyAI AssemblyAIConfig `mapstructure:"assemblyai"` // 新增
	OpenAI     struct {
		APIKey  string `mapstructure:"api_key"`
		BaseURL string `mapstructure:"base_url"`
	}
	Google struct {
		TTSKey string `mapstructure:"tts_key"`
		STTKey string `mapstructure:"stt_key"`
	}
	Azure struct {
		TTSKey string `mapstructure:"tts_key"`
		STTKey string `mapstructure:"stt_key"`
		Region string
	}
	AWS        AWSConfig `yaml:"aws"`
	Volcengine struct {
		STT struct {
			WsURL      string `mapstructure:"ws_url"`
			UID        string `mapstructure:"uid"`
			Rate       int    `mapstructure:"rate"`
			Format     string `mapstructure:"format"`
			Bits       int    `mapstructure:"bits"`
			Channel    int    `mapstructure:"channel"`
			Codec      string `mapstructure:"codec"`
			AccessKey  string `mapstructure:"access_key"`
			AppKey     string `mapstructure:"app_key"`
			ResourceID string `mapstructure:"resource_id"`
		} `mapstructure:"stt"`

		TTS struct {
			WsURL       string  `mapstructure:"ws_url"`
			AppID       string  `mapstructure:"app_id"`
			Token       string  `mapstructure:"token"`
			Cluster     string  `mapstructure:"cluster"`
			VoiceType   string  `mapstructure:"voice_type"`
			Encoding    string  `mapstructure:"encoding"`
			SpeedRatio  float64 `mapstructure:"speed_ratio"`
			VolumeRatio float64 `mapstructure:"volume_ratio"`
			PitchRatio  float64 `mapstructure:"pitch_ratio"`
		} `mapstructure:"tts"`
	} `mapstructure:"volcengine"`
	MinIO struct {
		Enabled     bool   `mapstructure:"enabled"`
		BucketName  string `mapstructure:"bucket_name"`
		Endpoint    string `mapstructure:"endpoint"`
		AccessKey   string `mapstructure:"access_key"`
		SecretKey   string `mapstructure:"secret_key"`
		UseSSL      bool   `mapstructure:"use_ssl"`
		Secure      bool   `mapstructure:"secure"`
		StoragePath string `mapstructure:"storage_path"`
	}
	Logging struct {
		Level        string
		Format       string
		Filename     string `mapstructure:"filename"`
		MaxSize      int    `mapstructure:"max_size"`
		MaxBackups   int    `mapstructure:"max_backups"`
		MaxAge       int    `mapstructure:"max_age"`
		Compress     bool   `mapstructure:"compress"`
		ReportCaller bool   `mapstructure:"report_caller"`
	}
}

var (
	cfg     *Config
	cfgOnce sync.Once
	cfgLock sync.RWMutex
)

// GetConfig 使用 sync.Once 确保配置只初始化一次
func GetConfig() (*Config, error) {
	var initErr error
	cfgOnce.Do(func() {
		cfg = &Config{}
		if err := viper.Unmarshal(cfg); err != nil {
			initErr = fmt.Errorf("无法解析配置结构体: %v", err)
		}
	})
	return cfg, initErr
}

func SetProvider(service string, provider string) error {
	cfgLock.Lock()
	defer cfgLock.Unlock()

	if cfg == nil {
		return fmt.Errorf("配置尚未初始化")
	}

	switch service {
	case "stt":
		cfg.STT.Provider = provider
	case "tts":
		cfg.TTS.Provider = provider
	case "llm":
		cfg.LLM.Provider = provider
	default:
		return fmt.Errorf("未知的服务: %s", service)
	}
	return nil
}
