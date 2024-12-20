package volcengine

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
)

type VolcengineTTS struct {
	wsURL      string
	appID      string
	token      string
	cluster    string
	voiceType  string
	encoding   string
	speedRatio float64
	volume     float64
	pitch      float64
}

func NewVolcengineTTS() *VolcengineTTS {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("配置初始化失败: %v", err)
	}

	ttsCfg := cfg.Volcengine.TTS
	return &VolcengineTTS{
		wsURL:      ttsCfg.WsURL,
		appID:      ttsCfg.AppID,
		token:      ttsCfg.Token,
		cluster:    ttsCfg.Cluster,
		voiceType:  ttsCfg.VoiceType,
		encoding:   ttsCfg.Encoding,
		speedRatio: ttsCfg.SpeedRatio,
		volume:     ttsCfg.VolumeRatio,
		pitch:      ttsCfg.PitchRatio,
	}
}

func (v *VolcengineTTS) Synthesize(text string) ([]byte, error) {
	// 构建 WebSocket URL
	u, err := url.Parse(v.wsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid WebSocket URL: %v", err)
	}

	// 设置请求头
	header := http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer;%s", v.token)},
	}

	// 建立 WebSocket 连接
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, fmt.Errorf("WebSocket连接失败: %v", err)
	}
	defer conn.Close()

	// 修改请求参数
	params := map[string]map[string]interface{}{
		"app": {
			"appid":   v.appID,
			"token":   v.token,
			"cluster": v.cluster,
		},
		"user": {
			"uid": fmt.Sprintf("user_%d", time.Now().UnixNano()),
		},
		"audio": {
			"voice_type":   v.voiceType,
			"encoding":     v.encoding,
			"speed_ratio":  v.speedRatio,
			"volume_ratio": v.volume,
			"pitch_ratio":  v.pitch,
		},
		"request": {
			"reqid":     generateReqID(),
			"text":      text,
			"text_type": "plain",
			"operation": "submit",
		},
	}

	// 序列化并压缩请求数据
	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
	}

	compressedData := gzipCompress(jsonData)

	// 构建二进制消息头
	message := buildMessage(compressedData)

	// 发送请求
	if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}

	// 修改响应处理
	var audioBuffer bytes.Buffer
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				break
			}
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}

		// 解析响应
		resp, err := parseResponse(message)
		if err != nil {
			return nil, fmt.Errorf("解析响应失败: %v", err)
		}

		// 检查错误
		if resp.Code != 0 {
			return nil, fmt.Errorf("服务端错误(code=%d): %s", resp.Code, resp.Message)
		}

		// 如果有音频数据,追加到 buffer
		if len(resp.Audio) > 0 {
			audioBuffer.Write(resp.Audio)
		}

		// 如果是最后一包数据,退出循环
		if resp.IsLast {
			break
		}
	}

	return audioBuffer.Bytes(), nil
}

// 工具函数
func generateReqID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func gzipCompress(input []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(input)
	w.Close()
	return b.Bytes()
}

func buildMessage(payload []byte) []byte {
	header := []byte{0x11, 0x10, 0x11, 0x00} // 默认消息头
	payloadSize := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSize, uint32(len(payload)))

	message := make([]byte, 0, len(header)+len(payloadSize)+len(payload))
	message = append(message, header...)
	message = append(message, payloadSize...)
	message = append(message, payload...)

	return message
}

// 添加响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Audio   []byte
	IsLast  bool
}

func parseResponse(res []byte) (*Response, error) {
	if len(res) < 4 {
		return nil, fmt.Errorf("响应数据长度不足")
	}

	// 解析二进制协议头
	// protoVersion := res[0] >> 4
	headSize := res[0] & 0x0f
	messageType := res[1] >> 4
	messageTypeSpecificFlags := res[1] & 0x0f
	// serializationMethod := res[2] >> 4
	messageCompression := res[2] & 0x0f
	payload := res[headSize*4:]

	resp := &Response{}

	// audio-only server response
	if messageType == 0xb {
		if messageTypeSpecificFlags == 0 {
			return resp, nil
		}

		sequenceNumber := int32(binary.BigEndian.Uint32(payload[0:4]))
		// payloadSize := int32(binary.BigEndian.Uint32(payload[4:8]))
		resp.Audio = append(resp.Audio, payload[8:]...)

		if sequenceNumber < 0 {
			resp.IsLast = true
		}
		return resp, nil
	}

	// error response
	if messageType == 0xf {
		code := int32(binary.BigEndian.Uint32(payload[0:4]))
		errMsg := payload[8:]
		if messageCompression == 1 {
			var err error
			errMsg, err = gzipDecompress(errMsg)
			if err != nil {
				return nil, fmt.Errorf("解压错误消息失败: %v", err)
			}
		}
		resp.Code = int(code)
		resp.Message = string(errMsg)
		return resp, fmt.Errorf("服务端错误(code=%d): %s", code, errMsg)
	}

	return nil, fmt.Errorf("未知的消息类型: %d", messageType)
}

func gzipDecompress(input []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
