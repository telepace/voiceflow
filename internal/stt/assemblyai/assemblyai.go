// assemblyai.go
package assemblyai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
	"io"
	"net/http"
)

const WAVE_FORMAT_PCM = 1

type AssemblyAI struct {
	apiKey string
}

func NewAssemblyAI() *AssemblyAI {
	logger.Info("Using AssemblyAI STT provider")
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("Failed to get config: %v", err)
	}
	return &AssemblyAI{
		apiKey: cfg.AssemblyAI.APIKey,
	}
}

func (a *AssemblyAI) Recognize(audioData []byte) (string, error) {
	// 将 PCM 数据包装成 WAV 格式
	wavData, err := wrapPCMDataToWAV(audioData)
	if err != nil {
		return "", fmt.Errorf("failed to wrap audio data to WAV: %v", err)
	}
	// 上传音频数据
	uploadURL, err := a.uploadAudioData(wavData)
	if err != nil {
		return "", fmt.Errorf("failed to upload audio data: %v", err)
	}
	// 请求转录
	transcriptText, err := a.requestTranscription(uploadURL)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio: %v", err)
	}
	return transcriptText, nil
}

func (a *AssemblyAI) uploadAudioData(audioData []byte) (string, error) {
	uploadURL := "https://api.assemblyai.com/v2/upload"

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(audioData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", a.apiKey)
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		UploadURL string `json:"upload_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.UploadURL, nil
}

func (a *AssemblyAI) requestTranscription(uploadURL string) (string, error) {
	transcriptURL := "https://api.assemblyai.com/v2/transcript"

	requestBody := map[string]interface{}{
		"audio_url":           uploadURL,
		"language_code":       "en_us",
		"punctuate":           true,
		"format_text":         true,
		"wait_for_completion": true,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", transcriptURL, bytes.NewReader(requestBodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("transcription request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Text   string `json:"text"`
		Error  string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Status != "completed" {
		return "", fmt.Errorf("transcription failed with status %s: %s", result.Status, result.Error)
	}

	return result.Text, nil
}

func wrapPCMDataToWAV(pcmData []byte) ([]byte, error) {
	const (
		sampleRate  = 16000
		bitDepth    = 16
		numChannels = 1
	)

	buf := &BufferWriteSeeker{}

	encoder := wav.NewEncoder(buf, sampleRate, bitDepth, numChannels, WAVE_FORMAT_PCM)

	// 将 PCM 数据转换为 audio.IntBuffer
	intBuf := &audio.IntBuffer{
		Data:           make([]int, len(pcmData)/2),
		Format:         &audio.Format{SampleRate: sampleRate, NumChannels: numChannels},
		SourceBitDepth: bitDepth,
	}

	// 假设 PCM 数据是 16 位有符号整数（小端序）
	for i := 0; i+1 < len(pcmData); i += 2 {
		sample := int16(pcmData[i]) | int16(pcmData[i+1])<<8
		intBuf.Data[i/2] = int(sample)
	}

	// 写入缓冲区
	if err := encoder.Write(intBuf); err != nil {
		return nil, err
	}

	// 关闭编码器以刷新数据
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
