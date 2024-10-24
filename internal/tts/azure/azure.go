package azure

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"

    "github.com/telepace/voiceflow/internal/config"
)

type AzureTTS struct {
    apiKey    string
    region    string
    endpoint  string
    voiceName string  // 可以根据需要增加配置
}

// NewAzureTTS 创建并返回一个新的 AzureTTS 实例
func NewAzureTTS() *AzureTTS {
    cfg := config.GetConfig()
    return &AzureTTS{
        apiKey:    cfg.Azure.TTSKey,
        region:    cfg.Azure.Region,
        endpoint:  fmt.Sprintf("https://%s.tts.speech.microsoft.com/cognitiveservices/v1", cfg.Azure.Region),
        voiceName: "en-US-AriaNeural",  // 设置默认语音，可以从配置文件读取
    }
}

// Synthesize 调用 Azure 的 TTS API，将文本转换为音频
func (a *AzureTTS) Synthesize(text string) ([]byte, error) {
    // 定义请求体
    requestBody, err := json.Marshal(map[string]interface{}{
        "text": text,
        "voiceName": a.voiceName,  // 使用指定的语音
        "locale": "en-US",         // 可以根据需要设置语言
        "format": "riff-16khz-16bit-mono-pcm",  // Azure TTS 音频格式
    })
    if err != nil {
        return nil, err
    }

    // 创建 HTTP 请求
    req, err := http.NewRequest("POST", a.endpoint, bytes.NewBuffer(requestBody))
    if err != nil {
        return nil, err
    }

    // 设置请求头
    req.Header.Set("Ocp-Apim-Subscription-Key", a.apiKey)
    req.Header.Set("Content-Type", "application/json")

    // 发送请求
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // 处理响应
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("Azure TTS error: %s", string(body))
    }

    return io.ReadAll(resp.Body)  // 返回音频数据
}