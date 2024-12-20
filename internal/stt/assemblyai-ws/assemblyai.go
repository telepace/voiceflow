// assemblyai.go
package assemblyai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
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

func (a *AssemblyAI) Recognize(audioData []byte, audioURL string) (string, error) {
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

	// 打印转录文本
	logger.Infof("Transcription result: %s", transcriptText)

	return transcriptText, nil
}

func (a *AssemblyAI) uploadAudioData(audioData []byte) (string, error) {
	url := "https://api.assemblyai.com/v2/upload"

	req, err := http.NewRequest("POST", url, bytes.NewReader(audioData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", a.apiKey)
	req.Header.Set("Content-Type", "audio/wav")
	req.Header.Set("Transfer-Encoding", "chunked")

	logger.Infof("Uploading audio data, size: %d bytes", len(audioData))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logger.Infof("Upload response status: %d, body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		UploadURL string `json:"upload_url"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return result.UploadURL, nil
}

func (a *AssemblyAI) requestTranscription(uploadURL string) (string, error) {
	transcriptURL := "https://api.assemblyai.com/v2/transcript"

	logger.Infof("Sending transcription request for audio URL: %s", uploadURL)

	requestBody := map[string]interface{}{
		"audio_url": uploadURL,
		// "language_code": "zh",
		"punctuate":   true,
		"format_text": true,
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
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// 轮询等待转录完成
	pollURL := fmt.Sprintf("%s/%s", transcriptURL, result.ID)
	for i := 0; i < 30; i++ { // 最多等待30次，每次3秒
		time.Sleep(3 * time.Second)

		req, err := http.NewRequest("GET", pollURL, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", a.apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}

		var pollResult struct {
			Status string `json:"status"`
			Text   string `json:"text"`
			Error  string `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&pollResult); err != nil {
			resp.Body.Close()
			return "", err
		}
		resp.Body.Close()

		logger.Infof("Transcription status: %s", pollResult.Status)

		switch pollResult.Status {
		case "completed":
			// 打印最终转录文本
			logger.Infof("Final transcription text: %s", pollResult.Text)
			return pollResult.Text, nil
		case "error":
			return "", fmt.Errorf("transcription failed: %s", pollResult.Error)
		case "processing", "queued":
			continue
		default:
			return "", fmt.Errorf("unknown status: %s", pollResult.Status)
		}
	}

	return "", fmt.Errorf("transcription timeout after 90 seconds")
}

func wrapPCMDataToWAV(pcmData []byte) ([]byte, error) {
	outBuffer := bytes.NewBuffer(nil)

	const (
		sampleRate    = 16000
		numChannels   = 1
		bitsPerSample = 16
	)

	outBuffer.WriteString("RIFF")
	writeInt32(outBuffer, uint32(len(pcmData)+36))
	outBuffer.WriteString("WAVE")

	outBuffer.WriteString("fmt ")
	writeInt32(outBuffer, 16)
	writeInt16(outBuffer, 1)
	writeInt16(outBuffer, numChannels)
	writeInt32(outBuffer, sampleRate)
	writeInt32(outBuffer, sampleRate*numChannels*bitsPerSample/8)
	writeInt16(outBuffer, numChannels*bitsPerSample/8)
	writeInt16(outBuffer, bitsPerSample)

	outBuffer.WriteString("data")
	writeInt32(outBuffer, uint32(len(pcmData)))
	outBuffer.Write(pcmData)

	return outBuffer.Bytes(), nil
}

func writeInt16(w *bytes.Buffer, value uint16) {
	w.WriteByte(byte(value))
	w.WriteByte(byte(value >> 8))
}

func writeInt32(w *bytes.Buffer, value uint32) {
	w.WriteByte(byte(value))
	w.WriteByte(byte(value >> 8))
	w.WriteByte(byte(value >> 16))
	w.WriteByte(byte(value >> 24))
}
