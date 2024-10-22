package server

import (
    "log"
    "net/http"
	"fmt"

    "github.com/gorilla/websocket"
    "github.com/telepace/voiceflow/internal/config"
)

type Server struct {
    upgrader websocket.Upgrader
    // 其他需要的字段
}

func NewServer() *Server {
    return &Server{
        upgrader: websocket.Upgrader{
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
            CheckOrigin: func(r *http.Request) bool {
                return true // 根据需要进行跨域处理
            },
        },
    }
}

func (s *Server) Start() {
    http.HandleFunc("/ws", s.handleConnections)
    cfg := config.GetConfig()
    addr := fmt.Sprintf(":%d", cfg.Server.Port)
    log.Printf("WebSocket server started on %s", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}