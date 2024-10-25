package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/telepace/voiceflow/internal/config"
)

type MinIOService struct {
	client     *minio.Client
	bucketName string
}

// NewMinIOService 创建并返回 MinIO 客户端
func NewMinIOService() *MinIOService {
	cfg := config.GetConfig()

	minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.Secure,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	return &MinIOService{
		client:     minioClient,
		bucketName: cfg.MinIO.BucketName,
	}
}

// StoreAudio 实现了 Service 接口，用于存储音频数据
func (m *MinIOService) StoreAudio(audioData []byte) (string, error) {
	ctx := context.Background()

	// 创建临时文件存储音频数据
	tempFile, err := ioutil.TempFile("", "audio-*.wav")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %v", err)
	}
	defer tempFile.Close()

	// 将音频数据写入临时文件
	_, err = tempFile.Write(audioData)
	if err != nil {
		return "", fmt.Errorf("error writing audio to temp file: %v", err)
	}

	objectName := tempFile.Name() // 生成文件名或使用自定义文件名

	// 上传文件到 MinIO
	info, err := m.client.FPutObject(ctx, m.bucketName, objectName, tempFile.Name(), minio.PutObjectOptions{
		ContentType: "audio/wav",
	})
	if err != nil {
		return "", fmt.Errorf("error uploading file to MinIO: %v", err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)

	// 生成预签名 URL
	reqParams := url.Values{}
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, time.Duration(24)*time.Hour, reqParams)
	if err != nil {
		return "", fmt.Errorf("error generating presigned URL: %v", err)
	}

	return presignedURL.String(), nil
}