package message

type Message interface {
    Process() error
}

type TextMessage struct {
    Text       string `json:"text"`
    RequireTTS bool   `json:"require_tts"` 
}

type BinaryMessage struct {
    Data []byte
}

type MessageHandler interface {
	HandleTextMessage(msg *TextMessage) error
	HandleBinaryMessage(msg *BinaryMessage) error
}
