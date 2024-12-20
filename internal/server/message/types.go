package message

// MessageType 定义消息类型
type MessageType int

const (
    TextType MessageType = iota
    BinaryType
    SignalType
)

// SignalMessage 定义信号消息
type SignalMessage struct {
    Type    string `json:"type"`    // 信号类型："end" 等
    Session string `json:"session"` // 会话ID，用于关联音频片段
}

// TextMessage 保持不变
type TextMessage struct {
    Text       string `json:"text"`
    RequireTTS bool   `json:"require_tts"`
}

// BinaryMessage 添加会话信息
type BinaryMessage struct {
    Data []byte
}

type AudioStartMessage struct {
    Type      string `json:"type"`      // "audio_start"
    SessionID string `json:"session_id"`
}

type AudioEndMessage struct {
    Type      string `json:"type"`      // "audio_end"
    SessionID string `json:"session_id"`
}

type MessageHandler interface {
	HandleTextMessage(msg *TextMessage) error
	HandleBinaryMessage(msg *BinaryMessage) error
}
