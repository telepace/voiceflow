// local.go
package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type LocalStorageService struct {
	storagePath string // 本地存储的根目录
}

// NewLocalStorageService 创建并返回一个新的本地存储服务
func NewLocalStorageService() *LocalStorageService {
	return &LocalStorageService{
		storagePath: "./audio_files", // 设置本地存储路径
	}
}

// StoreAudio 实现了 Service 接口，将音频文件存储到本地
func (l *LocalStorageService) StoreAudio(audioData []byte) (string, error) {
	// 检查存储目录是否存在，不存在则创建
	if _, err := os.Stat(l.storagePath); os.IsNotExist(err) {
		os.MkdirAll(l.storagePath, os.ModePerm)
	}

	// 创建文件名（可以基于时间戳生成唯一的文件名）
	fileName := fmt.Sprintf("audio_%d.wav", time.Now().UnixNano())
	filePath := filepath.Join(l.storagePath, fileName)

	// 将音频数据写入文件
	err := ioutil.WriteFile(filePath, audioData, 0644)
	if err != nil {
		return "", fmt.Errorf("error writing audio file to local storage: %v", err)
	}

	// 返回文件的相对路径
	return filePath, nil
}
