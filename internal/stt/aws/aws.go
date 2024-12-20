// internal/stt/aws/aws.go
package aws

import (
	"context"
	"fmt"
	"github.com/telepace/voiceflow/pkg/sttservice"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	transcribe "github.com/aws/aws-sdk-go/service/transcribestreamingservice"
	"github.com/aws/aws-sdk-go/service/transcribestreamingservice/transcribestreamingserviceiface"
	"github.com/telepace/voiceflow/pkg/config"
)

type Service struct {
	client transcribestreamingserviceiface.TranscribeStreamingServiceAPI
	config *config.AWSConfig
}

// 确保 Service 实现了 sttservice.Service 接口
var _ sttservice.Service = (*Service)(nil)

// NewService 创建新的 AWS STT 服务
func NewService(cfg *config.AWSConfig) (sttservice.Service, error) {
	awsConfig := &aws.Config{
		Region:      aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	}
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("无法创建 AWS 会话：%v", err)
	}
	client := transcribe.New(sess)
	return &Service{
		client: client,
		config: cfg,
	}, nil
}

// Recognize 实现了 stt.Service 接口的 Recognize 方法
func (s *Service) Recognize(audioData []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	input := &transcribe.StartStreamTranscriptionInput{
		LanguageCode:         aws.String("en-US"),
		MediaEncoding:        aws.String("pcm"),
		MediaSampleRateHertz: aws.Int64(16000),
	}

	output, err := s.client.StartStreamTranscriptionWithContext(ctx, input)
	if err != nil {
		return "", fmt.Errorf("无法开始转录流：%v", err)
	}

	eventStream := output.GetStream()

	// 发送音频数据
	go func() {
		defer eventStream.Close()
		err := eventStream.Send(ctx, &transcribe.AudioEvent{
			AudioChunk: audioData,
		})
		if err != nil {
			fmt.Printf("发送音频数据时出错：%v\n", err)
			return
		}
		// 发送完成后关闭发送方向的流
		eventStream.Close()
	}()

	// 接收转录结果
	var transcript string
	for event := range eventStream.Events() {
		switch e := event.(type) {
		case *transcribe.TranscriptEvent:
			results := e.Transcript.Results
			for _, result := range results {
				if !aws.BoolValue(result.IsPartial) {
					for _, alt := range result.Alternatives {
						transcript += aws.StringValue(alt.Transcript)
					}
				}
			}
		}
	}

	if err := eventStream.Err(); err != nil {
		return "", fmt.Errorf("转录错误：%v", err)
	}

	return transcript, nil
}

// StreamRecognize 实现了 stt.Service 接口的 StreamRecognize 方法
func (s *Service) StreamRecognize(ctx context.Context, audioDataChan <-chan []byte, transcriptChan chan<- string) error {
	input := &transcribe.StartStreamTranscriptionInput{
		LanguageCode:         aws.String("en-US"),
		MediaEncoding:        aws.String("pcm"),
		MediaSampleRateHertz: aws.Int64(16000),
	}

	output, err := s.client.StartStreamTranscriptionWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("无法开始转录流：%v", err)
	}

	eventStream := output.GetStream()

	// 发送音频数据的协程
	go func() {
		defer eventStream.Close()
		for {
			select {
			case audioChunk, ok := <-audioDataChan:
				if !ok {
					// 音频数据通道已关闭，结束发送
					return
				}
				err := eventStream.Send(ctx, &transcribe.AudioEvent{
					AudioChunk: audioChunk,
				})
				if err != nil {
					fmt.Printf("发送音频块时出错：%v\n", err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// 接收转录结果
	for event := range eventStream.Events() {
		switch e := event.(type) {
		case *transcribe.TranscriptEvent:
			results := e.Transcript.Results
			for _, result := range results {
				for _, alt := range result.Alternatives {
					transcript := aws.StringValue(alt.Transcript)
					// 发送部分转录结果
					transcriptChan <- transcript
				}
			}
		}
	}

	if err := eventStream.Err(); err != nil {
		return fmt.Errorf("转录错误：%v", err)
	}

	return nil
}
