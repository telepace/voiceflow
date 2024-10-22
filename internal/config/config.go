package config

import (
    "log"
    "sync"

    "github.com/joho/godotenv"
    "github.com/spf13/viper"
)

type Config struct {
    Server struct {
        Port     int
        EnableTLS bool `mapstructure:"enable_tls"`
    }
    MinIO struct {
        Enabled    bool
        BucketName string `mapstructure:"bucket_name"`
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
    Logging struct {
        Level string
    }
    // 更多配置...
}

var (
    cfg  *Config
    once sync.Once
)

func GetConfig() *Config {
    once.Do(func() {
        // 加载 .env 文件
        if err := godotenv.Load(); err != nil {
            log.Println("No .env file found")
        }

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
    })
    return cfg
}