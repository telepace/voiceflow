package local

import (
	"bytes"
	"io"
	"os/exec"
)

type LocalTTS struct {
	voice string // 本地 TTS 的语音配置
}

// NewLocalTTS 创建并返回一个新的 LocalTTS 实例
func NewLocalTTS() *LocalTTS {
	return &LocalTTS{
		voice: "en", // 本地 eSpeak 使用的默认语言
	}
}

// Synthesize 使用本地 TTS 生成语音（例如 eSpeak）
func (l *LocalTTS) Synthesize(text string) ([]byte, error) {
	// 使用 eSpeak 工具将文本转换为音频
	cmd := exec.Command("espeak", "-v", l.voice, "--stdout", text)
	audioData, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return io.ReadAll(bytes.NewReader(audioData))
}
