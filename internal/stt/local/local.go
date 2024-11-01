package local

import (
	"os/exec"
)

type LocalSTT struct {
	modelPath string // 本地模型路径
}

// NewLocalSTT 创建并返回一个新的 LocalSTT 实例
func NewLocalSTT() *LocalSTT {
	return &LocalSTT{
		modelPath: "/path/to/your/local/model", // 可以是 VOSK、DeepSpeech 等模型
	}
}

// Recognize 使用本地 STT 模型将音频转换为文本
func (l *LocalSTT) Recognize(audioData []byte) (string, error) {
	// 示例：调用 VOSK 命令行接口
	cmd := exec.Command("vosk", "--model", l.modelPath, "--input", "/path/to/input.wav")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
