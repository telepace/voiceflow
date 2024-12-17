// internal/server/server.go
package server

import (
	"github.com/telepace/voiceflow/pkg/logger"
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
				return true
			},
		},
	}
}

func (s *Server) SetupRoutes(mux *http.ServeMux) {
	if s == nil {
		logger.Error("Server instance is nil in SetupRoutes")
	} else {
		logger.Info("Server instance is not nil in SetupRoutes")
	}

	// 使用闭包来包装方法调用，确保正确捕获接收者 s
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.handleConnections(w, r)
	})

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		s.HandleConfig(w, r)
	})
}
