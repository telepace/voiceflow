package local

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/telepace/voiceflow/pkg/logger"
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
func (l *LocalSTT) Recognize(audioData []byte, audioURL string) (string, error) {
	if audioURL != "" {
		logger.Infof("本地 STT 不支持使用 audioURL，忽略该参数")
	}

	// 将音频数据写入临时文件
	tempFilePath := "/tmp/input.wav"
	err := writeTempFile(tempFilePath, audioData)
	if err != nil {
		return "", fmt.Errorf("写入临时音频文件失败: %v", err)
	}

	// 调用 VOSK 或其他本地 STT 工具进行识别
	// 示例：调用 VOSK 命令行接口
	cmd := exec.Command("vosk", "--model", l.modelPath, "--input", tempFilePath)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		logger.Errorf("本地 STT 命令执行错误: %v, stderr: %s", err, stderr.String())
		return "", fmt.Errorf("本地 STT 命令执行错误: %v, stderr: %s", err, stderr.String())
	}

	recognizedText := out.String()
	if recognizedText == "" {
		return "", fmt.Errorf("本地 STT 未能识别出文本")
	}

	return recognizedText, nil
}

// writeTempFile 将音频数据写入指定路径的临时文件
func writeTempFile(filePath string, data []byte) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo '%s' > %s", bytesToString(data), filePath))
	return cmd.Run()
}

// bytesToString 将字节数组转换为字符串，确保数据安全
func bytesToString(data []byte) string {
	return fmt.Sprintf("%s", data)
}
