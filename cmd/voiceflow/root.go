package main

import (
	"context"
	"fmt"
	"github.com/telepace/voiceflow/pkg/config"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/telepace/voiceflow/internal/server"
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

	// 初始化日志配置
	logCfg := logger.Config{
		Level:        viper.GetString("logging.level"),
		Format:       viper.GetString("logging.format"),
		Filename:     viper.GetString("logging.filename"),
		MaxSize:      viper.GetInt("logging.max_size"),
		MaxBackups:   viper.GetInt("logging.max_backups"),
		MaxAge:       viper.GetInt("logging.max_age"),
		Compress:     viper.GetBool("logging.compress"),
		ReportCaller: true,
	}

	// 服务标识信息
	fields := logger.StandardFields{
		ServiceID:  "voiceflow",
		InstanceID: fmt.Sprintf("instance-%d", time.Now().Unix()),
	}

	// 初始化日志系统
	if err := logger.Init(logCfg, fields); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// 记录启动信息
	logger.Info(ctx, "Starting VoiceFlow server")

	// 初始化配置
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error(ctx, "Failed to load configuration", "error", err)
		return fmt.Errorf("failed to get config: %w", err)
	}

	// 记录配置信息
	logger.Info(ctx, "Configuration loaded",
		"server_port", cfg.Server.Port,
		"enable_tls", cfg.Server.EnableTLS,
	)

	// 创建服务器实例
	mux := http.NewServeMux()

	// 注册路由处理器
	logger.Debug(ctx, "Registering HTTP handlers")
	mux.Handle("/", http.FileServer(http.Dir("./web")))
	mux.Handle("/audio_files/", http.StripPrefix("/audio_files/", http.FileServer(http.Dir("./audio_files"))))

	// 初始化 WebSocket 服务器
	wsServer := server.NewServer()
	wsServer.SetupRoutes(mux)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: loggingMiddleware(mux),
	}

	// 启动服务器
	logger.Info(ctx, "Server starting", "address", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(ctx, "Server failed to start", "error", err)
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
	rootCmd.PersistentFlags().String("logging.level", "info", "log level")
	rootCmd.PersistentFlags().String("logging.format", "json", "log format (json/text)")
	rootCmd.PersistentFlags().String("logging.filename", "", "log file path")

	// 绑定到 viper
	viper.BindPFlags(rootCmd.PersistentFlags())
}

func initConfig() {
	//ctx := context.Background()

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
	viper.SetDefault("logging.filename", "") // 默认输出到标准输出
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)
	viper.SetDefault("logging.compress", true)

	// 其他服务配置...
	viper.SetDefault("web.port", 18090)
	viper.SetDefault("minio.enabled", true)
	viper.SetDefault("minio.endpoint", "localhost:9000")
	// ... 其他配置项
}
