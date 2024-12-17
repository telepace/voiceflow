// pkg/voiceprocessor/voiceprocessor.go
package voiceprocessor

import (
	"fmt"
	"github.com/telepace/voiceflow/pkg/sttservice"
	"os"
)

func StartRealtime() error {
	// 实现实时语音监听和翻译的逻辑
	fmt.Println("实时语音处理已启动。")
	// 例如，使用麦克风输入并处理音频流
	// 这里可以调用 sttservice 中的 StreamRecognize 方法
	return nil
}

func TranscribeFile(audioFile string) error {
	// 检查文件是否存在
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在：%s", audioFile)
	}

	// 读取音频文件数据
	audioData, err := os.ReadFile(audioFile)
	if err != nil {
		return fmt.Errorf("无法读取音频文件：%v", err)
	}

	// 调用 STT 服务进行转录
	transcript, err := sttservice.Recognize(audioData)
	if err != nil {
		return fmt.Errorf("转录失败：%v", err)
	}

	// 输出转录结果
	fmt.Printf("转录结果：\n%s\n", transcript)
	return nil
}
