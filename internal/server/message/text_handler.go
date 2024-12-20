package message

import (
	"github.com/gorilla/websocket"
	"github.com/telepace/voiceflow/internal/storage"
	"github.com/telepace/voiceflow/internal/tts"
	"fmt"
)

type TextMessageHandler struct {
	tts     tts.Service
	storage storage.Service
}

func NewTextMessageHandler(tts tts.Service, storage storage.Service) *TextMessageHandler {
	return &TextMessageHandler{
		tts:     tts,
		storage: storage,
	}
}

func (h *TextMessageHandler) Handle(conn *websocket.Conn, msg *TextMessage) error {
	// 如果需要TTS,直接合成语音
	if msg.RequireTTS {
		audio, err := h.tts.Synthesize(msg.Text)
		if err != nil {
			return fmt.Errorf("failed to synthesize speech: %w", err)
		}
		
		audioURL, err := h.storage.StoreAudio(audio)
		if err != nil {
			return fmt.Errorf("failed to store audio: %w", err)
		}
		
		return conn.WriteJSON(map[string]string{
			"text": msg.Text,
			"audio_url": audioURL,
		})
	}
	
	// 如果不需要TTS,直接返回文本
	return conn.WriteJSON(map[string]string{
		"text": msg.Text,
	})
}
