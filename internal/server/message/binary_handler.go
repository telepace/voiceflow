package message

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/telepace/voiceflow/internal/storage"
	"github.com/telepace/voiceflow/internal/stt"
	"github.com/telepace/voiceflow/internal/tts"
	"github.com/telepace/voiceflow/pkg/logger"
)

type BinaryMessageHandler struct {
	stt          stt.Service
	tts          tts.Service
	storage      storage.Service
	audioBuffers map[string]*bytes.Buffer
	bufferMutex  sync.RWMutex
	oneShot      bool
}

func NewBinaryMessageHandler(stt stt.Service, tts tts.Service, storage storage.Service) *BinaryMessageHandler {
	return &BinaryMessageHandler{
		stt:          stt,
		tts:          tts,
		storage:      storage,
		audioBuffers: make(map[string]*bytes.Buffer),
		bufferMutex:  sync.RWMutex{},
	}
}

// HandleStart 处理音频开始信号
func (h *BinaryMessageHandler) HandleStart(sessionID string, oneShot bool) error {
	h.bufferMutex.Lock()
	defer h.bufferMutex.Unlock()

	if _, exists := h.audioBuffers[sessionID]; exists {
		return fmt.Errorf("session already exists: %s", sessionID)
	}

	h.audioBuffers[sessionID] = &bytes.Buffer{}
	h.oneShot = oneShot
	return nil
}

// HandleAudioData 处理音频二进制数据
func (h *BinaryMessageHandler) HandleAudioData(sessionID string, data []byte) error {
	h.bufferMutex.Lock()
	defer h.bufferMutex.Unlock()

	buffer, exists := h.audioBuffers[sessionID]
	if !exists {
		return fmt.Errorf("no active session found: %s", sessionID)
	}

	_, err := buffer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write audio data: %w", err)
	}

	return nil
}

// HandleEnd 处理音频结束信号
func (h *BinaryMessageHandler) HandleEnd(sessionID string, conn *websocket.Conn) error {
	// 1. 获取并清理音频数据
	audioData, err := h.getAndCleanAudioData(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get audio data: %w", err)
	}

	// 2. 存储到 MinIO
	audioURL, err := h.storage.StoreAudio(audioData)
	if err != nil {
		return fmt.Errorf("failed to store audio: %w", err)
	}

	// 3. 立即发送存储成功的响应
	err = conn.WriteJSON(map[string]interface{}{
		"type":       "audio_stored",
		"session_id": sessionID,
		"audio_url":  audioURL,
	})
	if err != nil {
		return fmt.Errorf("failed to send audio storage response: %w", err)
	}

	// 4. 异步进行语音识别
	go func() {
		text, err := h.stt.Recognize(audioData, audioURL)
		if err != nil {
			// 检查是否是最终错误（重试后仍然失败）
			if strings.Contains(err.Error(), "使用默认语言重试失败") {
				conn.WriteJSON(map[string]interface{}{
					"type":       "recognition_error",
					"session_id": sessionID,
					"error":      err.Error(),
				})
				return
			}

			// 如果是其他错误，记录日志但继续等待重试结果
			logger.Warnf("语音识别出现错误（可能正在重试）: %v", err)
			return
		}

		// 发送识别结果
		conn.WriteJSON(map[string]interface{}{
			"type":       "recognition_complete",
			"session_id": sessionID,
			"text":       text,
		})
	}()

	if h.oneShot {
		defer conn.Close()
	}

	return nil
}

// getAndCleanAudioData 获取并清理音频数据
func (h *BinaryMessageHandler) getAndCleanAudioData(sessionID string) ([]byte, error) {
	h.bufferMutex.Lock()
	defer h.bufferMutex.Unlock()

	buffer, exists := h.audioBuffers[sessionID]
	if !exists {
		return nil, fmt.Errorf("no audio data found for session: %s", sessionID)
	}

	audioData := buffer.Bytes()
	delete(h.audioBuffers, sessionID)
	return audioData, nil
}

// CleanupSession 清理指定会话的资源
func (h *BinaryMessageHandler) CleanupSession(sessionID string) {
	h.bufferMutex.Lock()
	defer h.bufferMutex.Unlock()
	delete(h.audioBuffers, sessionID)
}
