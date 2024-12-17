// pkg/sttservice/service.go
package sttservice

import (
	"context"
	"fmt"
)

type Service interface {
	Recognize(audioData []byte) (string, error)
	StreamRecognize(ctx context.Context, audioDataChan <-chan []byte, transcriptChan chan<- string) error
}

// 需要一个全局的 STT 服务实例
var sttInstance Service

// 提供一个方法来设置 STT 服务实例
func SetService(s Service) {
	sttInstance = s
}

// 提供全局可调用的 Recognize 方法
func Recognize(audioData []byte) (string, error) {
	if sttInstance == nil {
		return "", fmt.Errorf("STT 服务未初始化")
	}
	return sttInstance.Recognize(audioData)
}
