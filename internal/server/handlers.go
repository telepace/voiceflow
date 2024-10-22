package server

import (
	"log"
	"net/http"

	"github.com/telepace/voiceflow/internal/stt"
	"github.com/telepace/voiceflow/internal/llm"
	"github.com/telepace/voiceflow/internal/storage"
	"github.com/telepace/voiceflow/internal/tts"
)

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade error: %v", err)
		return
	}
	defer ws.Close()

	// 创建 STT、TTS、LLM 和存储实例
	sttService := stt.NewService()
	ttsService := tts.NewService()
	llmService := llm.NewService()
	storageService := storage.NewService()

	// 开始处理 WebSocket 消息
	for {
		// 读取消息
		_, data, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		// 处理语音数据 -> STT
		text, err := sttService.Recognize(data)
		if err != nil {
			log.Printf("STT error: %v", err)
			continue
		}

		// 与 LLM 交互
		responseText, err := llmService.GetResponse(text)
		if err != nil {
			log.Printf("LLM error: %v", err)
			continue
		}

		// 文本转语音 -> TTS
		audioData, err := ttsService.Synthesize(responseText)
		if err != nil {
			log.Printf("TTS error: %v", err)
			continue
		}

		// 存储音频到 MinIO，并获取 URL
		audioURL, err := storageService.StoreAudio(audioData)
		if err != nil {
			log.Printf("Storage error: %v", err)
			continue
		}

		// 将音频 URL 返回给前端
		err = ws.WriteJSON(map[string]string{"audio_url": audioURL})
		if err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
}
