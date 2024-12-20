package server

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type SessionManager struct {
	sessions       map[string]*bytes.Buffer
	currentSession string
	mu             sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*bytes.Buffer),
	}
}

func (sm *SessionManager) StartSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sessions[sessionID] = &bytes.Buffer{}
	sm.currentSession = sessionID
}

func (sm *SessionManager) GetCurrentSession() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentSession
}

func (sm *SessionManager) AppendAudioData(sessionID string, data []byte) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	buffer, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	_, err := buffer.Write(data)
	return err
}

func (sm *SessionManager) EndSession(sessionID string, ws *websocket.Conn) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	buffer, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// 获取音频数据
	audioData := buffer.Bytes()

	// 1. 首先存储音频文件
	audioURL, err := storageService.StoreAudio(audioData)
	if err != nil {
		return fmt.Errorf("failed to store audio: %w", err)
	}

	// 2. 发送音频存储成功的消息
	if err := ws.WriteJSON(map[string]interface{}{
		"type":      "audio_stored",
		"audio_url": audioURL,
	}); err != nil {
		return fmt.Errorf("failed to send audio storage response: %w", err)
	}

	// 3. 进行语音识别
	text, err := sttService.Recognize(audioData)
	if err != nil {
		// 发送识别错误消息
		return ws.WriteJSON(map[string]interface{}{
			"type":  "recognition_error",
			"error": err.Error(),
		})
	}

	// 4. 发送识别完成的消息
	if err := ws.WriteJSON(map[string]interface{}{
		"type": "recognition_complete",
		"text": text,
	}); err != nil {
		return fmt.Errorf("failed to send recognition result: %w", err)
	}

	// 5. 清理会话数据
	delete(sm.sessions, sessionID)
	if sm.currentSession == sessionID {
		sm.currentSession = ""
	}

	return nil
}
