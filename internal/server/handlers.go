// handlers.go - 服务器处理函数
package server

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"

	"github.com/gorilla/websocket"
	"github.com/telepace/voiceflow/internal/storage"
	"github.com/telepace/voiceflow/internal/stt"
	"github.com/telepace/voiceflow/internal/tts"
)

var (
	// 服务实例和锁
	serviceLock sync.RWMutex
	sttService  stt.Service
	ttsService  tts.Service
	// llmService     llm.Service
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
	// llmService = llm.NewService(cfg.LLM.Provider)
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

	// 创建会话管理器
	sessionManager := NewSessionManager()

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
			var msg map[string]interface{}
			if err := json.Unmarshal(data, &msg); err != nil {
				logger.Error("解析消息失败", "error", err)
				continue
			}

			// 检查是否是音频相关的控制消息
			if msgType, ok := msg["type"].(string); ok {
				switch msgType {
				case "audio_start":
					sessionID, _ := msg["session_id"].(string)
					sessionManager.StartSession(sessionID)
				case "audio_end":
					sessionID, _ := msg["session_id"].(string)
					if err := sessionManager.EndSession(sessionID, ws); err != nil {
						logger.Error("处理会话结束失败", "error", err)
					}
				}
			} else {
				// 处理普通文本消息
				text, _ := msg["text"].(string)
				requireTTS, _ := msg["require_tts"].(bool)

				if requireTTS {
					// 调用 TTS 服务
					audio, err := ttsService.Synthesize(text)
					if err != nil {
						logger.Error("语音合成失败", "error", err)
						continue
					}

					// 存储音频文件
					audioURL, err := storageService.StoreAudio(audio)
					if err != nil {
						logger.Error("存储音频失败", "error", err)
						continue
					}

					// 发送响应给客户端
					response := map[string]interface{}{
						"type":      "tts_complete",
						"text":      text,
						"audio_url": audioURL,
					}

					if err := ws.WriteJSON(response); err != nil {
						logger.Error("发送响应失败", "error", err)
					}
				}
			}

		case websocket.BinaryMessage:
			currentSession := sessionManager.GetCurrentSession()
			if currentSession == "" {
				logger.Error("收到二进制数据但没有活动会话")
				continue
			}

			if err := sessionManager.AppendAudioData(currentSession, data); err != nil {
				logger.Error("追加音频数据失败", "error", err)
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
