package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/telepace/voiceflow/internal/config"
	"github.com/telepace/voiceflow/internal/server"
	"github.com/telepace/voiceflow/pkg/logger"
)

func main() {
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
}
