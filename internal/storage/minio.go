// storage.go
package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
)

type MinIOService struct {
	client      *minio.Client
	bucketName  string
	storagePath string
}

// NewMinIOService 创建并返回 MinIO 客户端
func NewMinIOService() *MinIOService {
	cfg, err := config.GetConfig()
	ctx := context.Background()
	if err != nil {
		logger.Fatal(ctx, "配置初始化失败:", err)
	}

	minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.Secure,
	})
	if err != nil {
		logger.Fatal(ctx, "Failed to create MinIO client:", err)
	}

	return &MinIOService{
		client:      minioClient,
		bucketName:  cfg.MinIO.BucketName,
		storagePath: cfg.MinIO.StoragePath,
	}
}

// StoreAudio 实现了 Service 接口，用存储音频数据
func (m *MinIOService) StoreAudio(audioData []byte) (string, error) {
	ctx := context.Background()

	// 生成唯一文件名，并添加存储路径前缀
	objectName := fmt.Sprintf("%s%s.wav", m.storagePath, uuid.New().String())

	// 上传音频数据
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, bytes.NewReader(audioData), int64(len(audioData)), minio.PutObjectOptions{
		ContentType: "audio/wav",
	})
	if err != nil {
		return "", fmt.Errorf("上传到 MinIO 失败: %v", err)
	}

	// 生成预签名 URL
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, time.Hour*24, nil)
	if err != nil {
		return "", fmt.Errorf("生成预签名 URL 失败: %v", err)
	}

	return presignedURL.String(), nil
}

// DeleteAudio 删除音频文件
func (m *MinIOService) DeleteAudio(objectName string) error {
	ctx := context.Background()
	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("error deleting file from MinIO: %v", err)
	}
	log.Printf("Successfully deleted %s from MinIO\n", objectName)
	return nil
}
