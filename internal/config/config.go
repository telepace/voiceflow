package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/joho/godotenv"
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
	MinIO struct { // 添加 MinIO 配置结构体
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
	once    sync.Once
)

func GetConfig() *Config {
	cfgLock.RLock()
	defer cfgLock.RUnlock()
	return cfg
}

func loadConfig() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	log.Println("Loading config file")

	// 加载 config.yaml
	viper.SetConfigName("config")
	viper.AddConfigPath("./configs")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// 解析配置
	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}
}

func init() {
	once.Do(func() {
		cfgLock.Lock()
		defer cfgLock.Unlock()
		loadConfig()
	})
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
		return fmt.Errorf("unknown service: %s", service)
	}
	return nil
}
