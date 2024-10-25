package storage

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestMinIOService(t *testing.T) {
	// 初始化 MinIOService
	service := NewMinIOService()

	// 模拟音频数据
	audioData := []byte("test audio data")

	// 测试上传音频文件
	t.Run("StoreAudio", func(t *testing.T) {
		// 上传音频文件
		url, err := service.StoreAudio(audioData)
		assert.NoError(t, err, "Failed to store audio")
		assert.Contains(t, url, "http", "The returned URL should be valid")

		// 提取文件名，用于删除测试
		objectName := extractFileNameFromURL(url)

		// 测试删除音频文件
		t.Run("DeleteAudio", func(t *testing.T) {
			err := service.DeleteAudio(objectName)
			assert.NoError(t, err, "Failed to delete audio")
		})
	})
}

// extractFileNameFromURL 从 URL 中提取文件名（假设最后一个部分为文件名）
func extractFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
