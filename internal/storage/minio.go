package storage

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/telepace/voiceflow/internal/config"
)

type MinIOService struct {
	client     *minio.Client
	bucketName string
}

func NewMinIOService() *MinIOService {
	cfg := config.GetConfig()
	minioClient, err := minio.New(
		cfg.MinIO.Endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
			Secure: cfg.MinIO.UseSSL,
		},
	)
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}
	return &MinIOService{
		client:     minioClient,
		bucketName: cfg.MinIO.BucketName,
	}
}

func (m *MinIOService) StoreAudio(audioData []byte) (string, error) {
	// 存储音频数据到 MinIO，返回可访问的 URL
}
