package server

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/telepace/voiceflow/pkg/config"
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
	buffer, exists := sm.sessions[sessionID]
	sm.mu.Unlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	audioData := buffer.Bytes()
	var audioURL string

	// 判断当前 Provider 是否为 AssemblyAI
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("配置获取失败: %v", err)
	}

	if cfg.STT.Provider == "assemblyai" {
		// 上传音频到 MinIO
		audioURL, err = storageService.StoreAudio(audioData)
		if err != nil {
			return fmt.Errorf("failed to store audio: %w", err)
		}

		// 发送音频 URL 给前端
		ws.WriteJSON(map[string]interface{}{
			"type":       "audio_stored",
			"session_id": sessionID,
			"audio_url":  audioURL,
		})
	}

	// 调用 STT 服务
	text, err := sttService.Recognize(audioData, audioURL)
	if err != nil {
		// 发送错误响应
		ws.WriteJSON(map[string]interface{}{
			"type":       "recognition_error",
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return err
	}

	// 发送识别结果
	ws.WriteJSON(map[string]interface{}{
		"type":       "recognition_complete",
		"session_id": sessionID,
		"text":       text,
	})

	// 清理会话
	sm.mu.Lock()
	delete(sm.sessions, sessionID)
	if sm.currentSession == sessionID {
		sm.currentSession = ""
	}
	sm.mu.Unlock()

	return nil
}
