// models.go
package models

import (
	"time"
)

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Sender    string    `json:"sender"` // "user" 或 "assistant"
	Content   string    `json:"content"`
	AudioURL  string    `json:"audio_url"`
	CreatedAt time.Time `json:"created_at"`
}
