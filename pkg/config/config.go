// config.go
package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

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
	AssemblyAI struct {
		APIKey string `mapstructure:"api_key"`
	}
	OpenAI struct {
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
			VoiceType   string  `mapstructure:"voice_type"`
			Encoding    string  `mapstructure:"encoding"`
			SpeedRatio  float64 `mapstructure:"speed_ratio"`
			VolumeRatio float64 `mapstructure:"volume_ratio"`
			PitchRatio  float64 `mapstructure:"pitch_ratio"`
		} `mapstructure:"tts"`
	} `mapstructure:"volcengine"`
	MinIO struct {
		Enabled    bool   `mapstructure:"enabled"`
		BucketName string `mapstructure:"bucket_name"`
		Endpoint   string `mapstructure:"endpoint"`
		AccessKey  string `mapstructure:"access_key"`
		SecretKey  string `mapstructure:"secret_key"`
		UseSSL     bool   `mapstructure:"use_ssl"`
		Secure     bool   `mapstructure:"secure"`
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
