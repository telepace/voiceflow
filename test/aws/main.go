package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	transcribe "github.com/aws/aws-sdk-go/service/transcribestreamingservice"
	"github.com/gordonklaus/portaudio"
)

func main() {
	// 创建 AWS 会话
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"), // 根据您的实际区域修改
	})
	if err != nil {
		log.Fatal("无法创建 AWS 会话：", err)
	}

	// 创建 AWS Transcribe Streaming 客户端
	client := transcribe.New(sess)

	// 初始化 PortAudio
	err = portaudio.Initialize()
	if err != nil {
		log.Fatal("无法初始化 PortAudio：", err)
	}
	defer portaudio.Terminate()

	// 音频流参数
	const sampleRate = 16000
	const channels = 1
	const framesPerBuffer = 512 // 设置较小的缓冲区

	// 创建音频数据通道，带缓冲区防止阻塞
	audioChan := make(chan []int16, 100)

	// 创建 PortAudio 输入流，使用回调函数
	stream, err := portaudio.OpenDefaultStream(channels, 0, sampleRate, framesPerBuffer, func(in []int16) {
		// 复制输入数据
		data := make([]int16, len(in))
		copy(data, in)
		// 将数据发送到通道，如果通道已满则丢弃数据以防止阻塞
		select {
		case audioChan <- data:
		default:
			// 通道已满，丢弃数据
		}
	})
	if err != nil {
		log.Fatal("无法打开音频流：", err)
	}
	defer stream.Close()

	// 启动音频流
	err = stream.Start()
	if err != nil {
		log.Fatal("无法启动音频流：", err)
	}
	defer stream.Stop()

	fmt.Println("请开始说话... 按下 Ctrl+C 结束")

	// 设置 AWS Transcribe Streaming 输入参数
	input := &transcribe.StartStreamTranscriptionInput{
		LanguageCode:         aws.String("zh-CN"),
		MediaEncoding:        aws.String("pcm"),
		MediaSampleRateHertz: aws.Int64(sampleRate),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 开始转录流
	output, err := client.StartStreamTranscriptionWithContext(ctx, input)
	if err != nil {
		log.Fatal("无法开始转录流：", err)
	}

	eventStream := output.GetStream()

	// 处理系统信号，支持优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建 WaitGroup 等待 Goroutine 完成
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine：从通道读取音频数据并发送到 AWS Transcribe
	go func() {
		defer wg.Done()
		for {
			select {
			case data := <-audioChan:
				// 将 []int16 转换为 []byte
				audioBytes := int16SliceToByteSlice(data)
				// 发送音频数据到 AWS Transcribe
				err := eventStream.Send(ctx, &transcribe.AudioEvent{
					AudioChunk: audioBytes,
				})
				if err != nil {
					log.Println("发送音频事件失败：", err)
					cancel()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Goroutine：接收并处理转录结果
	go func() {
		defer wg.Done()
		for event := range eventStream.Events() {
			switch e := event.(type) {
			case *transcribe.TranscriptEvent:
				results := e.Transcript.Results
				for _, result := range results {
					if !aws.BoolValue(result.IsPartial) {
						for _, alt := range result.Alternatives {
							fmt.Println("转录结果：", aws.StringValue(alt.Transcript))
						}
					}
				}
			default:
				// 处理其他事件
			}
		}
		if err := eventStream.Err(); err != nil {
			log.Println("事件流出错：", err)
			cancel()
		}
	}()

	// 等待退出信号
	<-sigChan
	fmt.Println("录音结束")

	// 取消上下文，停止 Goroutine
	cancel()

	// 等待 Goroutine 完成
	wg.Wait()

	// 关闭事件流
	eventStream.Close()
}

func int16SliceToByteSlice(data []int16) []byte {
	buf := make([]byte, len(data)*2)
	for i, v := range data {
		buf[i*2] = byte(v)
		buf[i*2+1] = byte(v >> 8)
	}
	return buf
}
