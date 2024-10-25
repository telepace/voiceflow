package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

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
	OpenAI struct {
		APIKey string `mapstructure:"api_key"`
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
		Level string
	}
}

var (
	cfg     *Config
	cfgLock sync.RWMutex
)

func GetConfig() *Config {
	cfgLock.RLock()
	defer cfgLock.RUnlock()
	if cfg == nil {
		cfgLock.RUnlock()
		cfgLock.Lock()
		defer cfgLock.Unlock()
		if cfg == nil {
			cfg = &Config{}
			if err := viper.Unmarshal(cfg); err != nil {
				panic(fmt.Errorf("无法解析配置结构体: %v", err))
			}
		}
		cfgLock.RLock()
	}
	return cfg
}

func SetProvider(service string, provider string) error {
	cfgLock.Lock()
	defer cfgLock.Unlock()

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
