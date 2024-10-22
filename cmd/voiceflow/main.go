package main

import (
    "log"

    "github.com/telepace/voiceflow/internal/config"
    "github.com/telepace/voiceflow/internal/server"
    "github.com/telepace/voiceflow/pkg/logger"
)

func main() {
    // 初始化配置
    cfg := config.GetConfig()

    // 初始化日志
    logger.Init(cfg.Logging.Level)

    // 启动 WebSocket 服务器
    wsServer := server.NewServer()
    go wsServer.Start()

    // 可以在这里添加其他服务，如 RESTful API 服务器等

    // 阻塞主线程
    select {}
}