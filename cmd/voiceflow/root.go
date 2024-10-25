package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/telepace/voiceflow/internal/config"
	"github.com/telepace/voiceflow/internal/server"
	"github.com/telepace/voiceflow/pkg/logger"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "voiceflow",
	Short: "VoiceFlow is a voice processing server",
	Long:  `VoiceFlow is a server application for processing voice data.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化配置
		cfg := config.GetConfig()

		// 初始化日志
		logger.Init(cfg.Logging.Level)

		// 创建一个新的 ServeMux
		mux := http.NewServeMux()

		// 提供静态文件服务
		mux.Handle("/", http.FileServer(http.Dir("./web")))
		mux.Handle("/audio_files/", http.StripPrefix("/audio_files/", http.FileServer(http.Dir("./audio_files"))))

		// 初始化服务器并设置路由
		wsServer := server.NewServer()
		wsServer.SetupRoutes(mux)

		// 启动服务器
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Printf("Server started on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 定义持久化标志（全局）
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件 (默认查找路径: ./config.yaml)")

	// 绑定常用标志
	rootCmd.PersistentFlags().Int("server.port", 18080, "服务器端口")
	rootCmd.PersistentFlags().Bool("server.enable_tls", false, "是否启用 TLS")
	rootCmd.PersistentFlags().String("logging.level", "info", "日志级别")

	// 绑定标志到 Viper 配置
	viper.BindPFlag("server.port", rootCmd.PersistentFlags().Lookup("server.port"))
	viper.BindPFlag("server.enable_tls", rootCmd.PersistentFlags().Lookup("server.enable_tls"))
	viper.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("logging.level"))
}

func initConfig() {
	if cfgFile != "" {
		// 使用命令行标志指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else if os.Getenv("VOICEFLOW_CONFIG") != "" {
		// 使用环境变量指定的配置文件
		viper.SetConfigFile(os.Getenv("VOICEFLOW_CONFIG"))
	} else {
		// 默认配置文件路径
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc/voiceflow/")
		viper.AddConfigPath("$HOME/.voiceflow")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs") // 添加项目的 configs 目录
	}

	viper.SetEnvPrefix("VOICEFLOW")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // 自动读取匹配的环境变量

	// 读取配置文件
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("使用的配置文件:", viper.ConfigFileUsed())
	} else {
		fmt.Println("未找到配置文件，使用默认配置")
	}

	// 设置默认值
	setDefaults()
}

func setDefaults() {
	viper.SetDefault("server.port", 18080)
	viper.SetDefault("server.enable_tls", false)
	viper.SetDefault("web.port", 18090)
	viper.SetDefault("minio.enabled", true)
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.access_key", "minioadmin")
	viper.SetDefault("minio.secret_key", "minioadmin")
	viper.SetDefault("minio.bucket_name", "telepace-pipeline")
	viper.SetDefault("minio.secure", true)
	viper.SetDefault("stt.provider", "azure")
	viper.SetDefault("tts.provider", "google")
	viper.SetDefault("llm.provider", "openai")
	viper.SetDefault("azure.stt_key", "your_azure_stt_key")
	viper.SetDefault("azure.tts_key", "your_azure_tts_key")
	viper.SetDefault("azure.region", "eastus")
	viper.SetDefault("google.stt_key", "your_google_stt_key")
	viper.SetDefault("google.tts_key", "your_google_tts_key")
	viper.SetDefault("openai.api_key", "your_openai_api_key")
	viper.SetDefault("logging.level", "debug")
}
