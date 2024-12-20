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
	buffer, exists := sm.sessions[sessionID]
	audioData := buffer.Bytes()
	sm.mu.Unlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	go func() {
		audioURLChan := make(chan string, 1)
		errChan := make(chan error, 1)

		go func() {
			audioURL, err := storageService.StoreAudio(audioData)
			if err != nil {
				errChan <- err
				return
			}
			audioURLChan <- audioURL
		}()

		go func() {
			text, err := sttService.Recognize(audioData, "")
			if err != nil {
				ws.WriteJSON(map[string]interface{}{
					"type":       "recognition_error",
					"session_id": sessionID,
					"error":      err.Error(),
				})
				return
			}

			ws.WriteJSON(map[string]interface{}{
				"type":       "recognition_complete",
				"session_id": sessionID,
				"text":       text,
			})
		}()

		select {
		case audioURL := <-audioURLChan:
			ws.WriteJSON(map[string]interface{}{
				"type":       "audio_stored",
				"session_id": sessionID,
				"audio_url":  audioURL,
			})
		case err := <-errChan:
			ws.WriteJSON(map[string]interface{}{
				"type":  "storage_error",
				"error": err.Error(),
			})
		}
	}()

	sm.mu.Lock()
	delete(sm.sessions, sessionID)
	if sm.currentSession == sessionID {
		sm.currentSession = ""
	}
	sm.mu.Unlock()

	return nil
}
