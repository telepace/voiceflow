// root.go
package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/telepace/voiceflow/pkg/config"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	serverpkg "github.com/telepace/voiceflow/internal/server"
	"github.com/telepace/voiceflow/pkg/logger"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "voiceflow",
	Short: "VoiceFlow is a voice processing server",
	Long:  `VoiceFlow is a server application for processing voice data.`,
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

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
	mux.Handle("/", http.FileServer(http.Dir("./web")))
	mux.Handle("/audio_files/", http.StripPrefix("/audio_files/", http.FileServer(http.Dir("./audio_files"))))

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
}

func initConfig() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		logger.Fatal("No .env file found or failed to load, proceeding without it")
	} else {
		logger.Info(".env file loaded")
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

	// 其他服务配置...
	viper.SetDefault("web.port", 18090)
	viper.SetDefault("minio.enabled", true)
	viper.SetDefault("minio.endpoint", "localhost:9000")
}
