package message

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/telepace/voiceflow/internal/storage"
	"github.com/telepace/voiceflow/internal/stt"
	"github.com/telepace/voiceflow/internal/tts"
)

type BinaryMessageHandler struct {
	stt     stt.Service
	tts     tts.Service
	storage storage.Service
}

func NewBinaryMessageHandler(stt stt.Service, tts tts.Service, storage storage.Service) *BinaryMessageHandler {
	return &BinaryMessageHandler{
		stt:     stt,
		tts:     tts,
		storage: storage,
	}
}

func (h *BinaryMessageHandler) Handle(conn *websocket.Conn, msg *BinaryMessage) error {
	// 1. STT 转换
	text, err := h.stt.Recognize(msg.Data)
	if err != nil {
		return fmt.Errorf("speech recognition failed: %w", err)
	}

	// 2. 直接返回识别的文本
	return conn.WriteJSON(map[string]string{
		"text": text,
	})
}
