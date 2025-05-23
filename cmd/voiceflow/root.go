// cmd/voiceflow/root.go
package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/sttservice"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	serverpkg "github.com/telepace/voiceflow/internal/server"
	"github.com/telepace/voiceflow/pkg/logger"
)

var cfgFile string

//go:embed web/*
var webFS embed.FS

// setupFileServers sets up the file servers for both embedded and local files
func setupFileServers(mux *http.ServeMux) error {
	// Setup web content from embedded files
	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(webContent)))

	// Setup audio files from local directory
	// This allows for dynamic audio file serving without embedding
	mux.Handle("/audio_files/", http.StripPrefix("/audio_files/",
		http.FileServer(http.Dir("audio_files"))))

	return nil
}

// ensureDirectories creates necessary directories if they don't exist
func ensureDirectories() error {
	dirs := []string{
		"audio_files",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "voiceflow",
	Short: "VoiceFlow is a voice processing server",
	Long:  `VoiceFlow is a server application for processing voice data.`,
	RunE:  run,
}

// 添加新的子命令 transcribe
var transcribeCmd = &cobra.Command{
	Use:   "transcribe",
	Short: "Transcribe an audio file using STT service",
	Long:  `Transcribe an audio file by specifying its path and using the configured STT service.`,
	RunE:  runTranscribe,
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if err := ensureDirectories(); err != nil {
		logger.Fatalf("Failed to ensure directories: %v", err)
	}
	// Load configuration
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Initialize logger
	logCfg := logger.Config{
		Level:        cfg.Logging.Level,
		Format:       cfg.Logging.Format,
		Filename:     cfg.Logging.Filename,
		MaxSize:      cfg.Logging.MaxSize,
		MaxBackups:   cfg.Logging.MaxBackups,
		MaxAge:       cfg.Logging.MaxAge,
		Compress:     cfg.Logging.Compress,
		ReportCaller: cfg.Logging.ReportCaller,
	}

	fields := logger.StandardFields{
		ServiceID:  "voiceflow",
		InstanceID: fmt.Sprintf("instance-%d", time.Now().Unix()),
	}

	if err := logger.Init(logCfg, fields); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// 记录启动信息
	logger.InfoContextf(ctx, "Starting VoiceFlow server with config: %+v", cfg)

	serverpkg.InitServices()

	// Set up HTTP server
	mux := http.NewServeMux()
	if err := setupFileServers(mux); err != nil {
		logger.Fatalf("Failed to setup file servers: %v", err)
	}

	// Initialize WebSocket server
	wsServer := serverpkg.NewServer()
	if wsServer == nil {
		logger.Fatal("Failed to create Server instance")
	}

	wsServer.SetupRoutes(mux)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: loggingMiddleware(mux),
	}

	// Start server
	logger.InfoContext(ctx, "Server starting", "address", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.ErrorContext(ctx, "Server failed to start", "error", err)
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// 简单的日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		start := time.Now()

		logger.Info(ctx, "Request started",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		next.ServeHTTP(w, r)

		logger.Info(ctx, "Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).String(),
		)
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error(context.Background(), "Command execution failed", "error", err)
		os.Exit(1)
	}
}

var transcribeFile string

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")

	// 基础配置项
	rootCmd.PersistentFlags().Int("server.port", 18080, "server port")
	rootCmd.PersistentFlags().Bool("server.enable_tls", false, "enable TLS")

	// 日志相关配置项
	// Logging flags
	rootCmd.PersistentFlags().String("logging.level", "info", "log level (debug, info, warn, error, fatal)")
	rootCmd.PersistentFlags().String("logging.format", "json", "log format (json/text)")
	rootCmd.PersistentFlags().String("logging.filename", "", "log file path")
	rootCmd.PersistentFlags().Int("logging.max_size", 100, "maximum size in MB before log file rotation")
	rootCmd.PersistentFlags().Int("logging.max_backups", 3, "maximum number of old log files to retain")
	rootCmd.PersistentFlags().Int("logging.max_age", 28, "maximum number of days to retain old log files")
	rootCmd.PersistentFlags().Bool("logging.compress", true, "whether to compress old log files")
	rootCmd.PersistentFlags().Bool("logging.report_caller", true, "whether to include caller information in logs")

	// 绑定到 viper
	viper.BindPFlags(rootCmd.PersistentFlags())

	// 配置 transcribe 子命令的标志
	transcribeCmd.Flags().StringVarP(&transcribeFile, "file", "f", "", "Path to the audio file to transcribe")
	transcribeCmd.MarkFlagRequired("file") // 标记为必需

	// 将 transcribe 子命令添加到 rootCmd
	rootCmd.AddCommand(transcribeCmd)
}

func initConfig() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found or failed to load, proceeding without it")
	} else {
		envPath, _ := os.Getwd()
		logger.Info(fmt.Sprintf(".env file loaded from: %s/.env", envPath))
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("/etc/voiceflow/")
		viper.AddConfigPath("$HOME/.voiceflow")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
	}

	viper.SetEnvPrefix("VOICEFLOW")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	} else {
		fmt.Println("No config file found, using defaults")
	}

	setDefaults()
}

func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.port", 18080)
	viper.SetDefault("server.enable_tls", false)

	// 日志默认配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.filename", "")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)
	viper.SetDefault("logging.compress", true)
	viper.SetDefault("logging.report_caller", true)

	// AWS 默认配置
	viper.SetDefault("aws.region", "us-east-2")

	// 其他服务配置...
	viper.SetDefault("web.port", 18090)
	viper.SetDefault("minio.enabled", true)
	viper.SetDefault("minio.endpoint", "localhost:9000")
}

// runTranscribe 处理 transcribe 子命令的逻辑
func runTranscribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 初始化配置
	if err := ensureDirectories(); err != nil {
		logger.Fatalf("Failed to ensure directories: %v", err)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// 初始化日志
	logCfg := logger.Config{
		Level:        cfg.Logging.Level,
		Format:       cfg.Logging.Format,
		Filename:     cfg.Logging.Filename,
		MaxSize:      cfg.Logging.MaxSize,
		MaxBackups:   cfg.Logging.MaxBackups,
		MaxAge:       cfg.Logging.MaxAge,
		Compress:     cfg.Logging.Compress,
		ReportCaller: cfg.Logging.ReportCaller,
	}

	fields := logger.StandardFields{
		ServiceID:  "voiceflow",
		InstanceID: fmt.Sprintf("instance-%d", time.Now().Unix()),
	}

	if err := logger.Init(logCfg, fields); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// 记录启动信息
	logger.InfoContextf(ctx, "Starting VoiceFlow transcribe command with config: %+v", cfg)

	// 初始化服务
	serverpkg.InitServices()

	// 读取音频文件
	audioData, err := ioutil.ReadFile(transcribeFile)
	if err != nil {
		logger.Errorf("Failed to read audio file: %v", err)
		return fmt.Errorf("failed to read audio file: %w", err)
	}

	// 调用 STT 服务进行转录
	transcript, err := sttservice.Recognize(audioData)
	if err != nil {
		logger.Errorf("STT Recognize error: %v", err)
		return fmt.Errorf("STT Recognize error: %w", err)
	}

	// 输出转录结果
	fmt.Printf("Transcript:\n%s\n", transcript)

	return nil
}
