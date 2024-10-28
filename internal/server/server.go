package server

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	upgrader websocket.Upgrader
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

func (s *Server) SetupRoutes(mux *http.ServeMux) {
	// WebSocket 路由
	mux.HandleFunc("/ws", s.handleConnections)

	// 配置更改的 RESTful API
	mux.HandleFunc("/config", s.HandleConfig)
}
