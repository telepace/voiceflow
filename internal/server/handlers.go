// handlers.go - 服务器处理函数
package server

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"

	"github.com/gorilla/websocket"
	"github.com/telepace/voiceflow/internal/llm"
	"github.com/telepace/voiceflow/internal/server/message"
	"github.com/telepace/voiceflow/internal/storage"
	"github.com/telepace/voiceflow/internal/stt"
	"github.com/telepace/voiceflow/internal/tts"
)

var (
	// 服务实例和锁
	serviceLock    sync.RWMutex
	sttService     stt.Service
	ttsService     tts.Service
	llmService     llm.Service
	storageService storage.Service
)

// 初始化服务实例
func InitServices() {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("配置初始化失败: %v", err)
	}
	sttService = stt.NewService(cfg.STT.Provider)
	ttsService = tts.NewService(cfg.TTS.Provider)
	llmService = llm.NewService(cfg.LLM.Provider)
	storageService = storage.NewService()
}

// 修改消息结构
type TextMessage struct {
	Text       string `json:"text"`
	RequireTTS bool   `json:"require_tts"`
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	// 初始化消息处理器 - 移除 llmService
	textHandler := message.NewTextMessageHandler(ttsService, storageService)
	binaryHandler := message.NewBinaryMessageHandler(sttService, ttsService, storageService)

	for {
		mt, data, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.Info("WebSocket connection closed normally")
			} else {
				logger.Error("WebSocket read error", "error", err)
			}
			break
		}

		switch mt {
		case websocket.TextMessage:
			var msg message.TextMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				logger.Error("Failed to parse text message", "error", err)
				ws.WriteJSON(map[string]string{"error": "Invalid message format"})
				continue
			}

			if err := textHandler.Handle(ws, &msg); err != nil {
				logger.Error("Failed to handle text message", "error", err)
				ws.WriteJSON(map[string]string{
					"error":   "Failed to process message",
					"details": err.Error(),
				})
			}

		case websocket.BinaryMessage:
			msg := &message.BinaryMessage{Data: data}
			if err := binaryHandler.Handle(ws, msg); err != nil {
				logger.Error("Failed to handle binary message", "error", err)
				ws.WriteJSON(map[string]string{
					"error":   "Failed to process audio",
					"details": err.Error(),
				})
			}
		}
	}
}

// 配置更新处理函数
func (s *Server) HandleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Service  string `json:"service"`
		Provider string `json:"provider"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 更新配置，使用写锁保护
	serviceLock.Lock()
	defer serviceLock.Unlock()

	err = config.SetProvider(req.Service, req.Provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 根据新的配置重新初始化服务实例
	//initServices()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Configuration updated"))
}
