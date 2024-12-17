// handlers.go - 服务器处理函数
package server

import (
	"encoding/json"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/telepace/voiceflow/internal/llm"
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

	// 升级 WebSocket 连接
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("WebSocket Upgrade error: %v", err)
		return
	}
	defer ws.Close()

	for {
		mt, data, err := ws.ReadMessage()
		if err != nil {
			logger.Errorf("Read error: %v", err)
			break
		}

		// 获取最新的服务实例
		serviceLock.RLock()
		currentSTTService := sttService
		currentTTSService := ttsService
		currentLLMService := llmService
		currentStorageService := storageService
		serviceLock.RUnlock()

		if mt == websocket.TextMessage {
			logger.Debug("Received text message")
			var msg TextMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				logger.Error("JSON parse error: %v", err)
				continue
			}
			
			// 调用 TTS 服务
			audioData, err := currentTTSService.Synthesize(msg.Text)
			if err != nil {
				logger.Error("TTS error: %v", err)
				continue
			}
			
			// 存储音频并获取 URL
			audioURL, err := currentStorageService.StoreAudio(audioData)
			if err != nil {
				logger.Error("Storage error: %v", err)
				continue
			}
			
			// 返回文本和音频 URL
			response := map[string]string{
				"text": msg.Text,
				"audio_url": audioURL,
			}
			ws.WriteJSON(response)
		} else if mt == websocket.BinaryMessage {
			logger.Debug("Received binary message")
			// 处理音频消息
			// 使用 STT 服务将语音转换为文字
			text, err := currentSTTService.Recognize(data)
			if err != nil {
				logger.Errorf("STT error: %v", err)
				continue
			}

			// 调用 LLM 服务获取响应
			responseText, err := currentLLMService.GetResponse(text)
			if err != nil {
				logger.Errorf("LLM error: %v", err)
				continue
			}

			// 返回文本响应给前端
			ws.WriteJSON(map[string]string{"text": responseText})
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
