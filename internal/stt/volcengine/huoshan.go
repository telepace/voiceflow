// volcengine.go
package volcengine

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
	"io"
	"net/http"
)

type STT struct {
	wsURL     string
	uid       string
	rate      int
	format    string
	bits      int
	channel   int
	codec     string
	accessKey string
	appKey    string
}

func NewVolcengineSTT() *STT {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("配置初始化失败: %v", err)
	}
	return &STT{
		wsURL:     cfg.Volcengine.WsURL,
		uid:       cfg.Volcengine.UID,
		rate:      cfg.Volcengine.Rate,
		format:    cfg.Volcengine.Format,
		bits:      cfg.Volcengine.Bits,
		channel:   cfg.Volcengine.Channel,
		codec:     cfg.Volcengine.Codec,
		accessKey: cfg.Volcengine.AccessKey,
		appKey:    cfg.Volcengine.AppKey,
	}
}

func (s *STT) Recognize(audioData []byte) (string, error) {
	reqID := uuid.New().String()
	connectorID := uuid.New().String()

	header := http.Header{}
	header.Set("X-Api-Resource-Id", "volc.bigasr.sauc.duration")
	header.Set("X-Api-Access-Key", s.accessKey)
	header.Set("X-Api-App-Key", s.appKey)
	header.Set("X-Api-Connect-Id", connectorID)

	dialer := websocket.DefaultDialer
	conn, resp, err := dialer.Dial(s.wsURL, header)
	if err != nil {
		logger.Error("WebSocket 连接错误:", "reqID:", reqID, "ws URL:", s.wsURL, "api Key:", s.accessKey, err)
		return "", err
	}
	defer conn.Close()

	// 检查并打印 X-Api-Connect-Id 和 X-Tt-Logid
	if connectID := resp.Header.Get("X-Api-Connect-Id"); connectID != "" {
		logger.Infof("连接追踪ID: X-Api-Connect-Id = %s", connectID)
	}
	if logID := resp.Header.Get("X-Tt-Logid"); logID != "" {
		logger.Infof("服务端返回的logid: X-Tt-Logid = %s", logID)
	}

	// 构建并发送初始请求
	req := map[string]interface{}{
		"user": map[string]interface{}{
			"uid": s.uid,
		},
		"audio": map[string]interface{}{
			"format":  s.format,
			"rate":    s.rate,
			"bits":    s.bits,
			"channel": s.channel,
			"codec":   s.codec,
		},
	}

	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	compressedPayload, err := gzipCompress(payloadBytes)
	if err != nil {
		return "", err
	}

	err = s.sendMessage(conn, FULL_CLIENT_REQUEST, POS_SEQUENCE, JSON_SERIALIZATION, compressedPayload, 1)
	if err != nil {
		logger.Errorf("发送初始消息错误: %v", err)
		return "", err
	}

	// 处理响应
	var finalText string
	for {
		_, respData, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				logger.Info("WebSocket 连接正常关闭")
				break
			} else {
				logger.Errorf("读取响应错误: %v", err)
				return "", err
			}
		}

		result, err := parseResponse(respData)
		if err != nil {
			logger.Errorf("解析响应错误: %v", err)
			return "", err
		}

		if payloadMsg, ok := result["payload_msg"]; ok {
			if payloadMap, ok := payloadMsg.(map[string]interface{}); ok {
				if resultMap, ok := payloadMap["result"].(map[string]interface{}); ok {
					if text, ok := resultMap["text"].(string); ok {
						logger.Infof("识别结果: %s", text)
						finalText = text
					}
				}
			}
		}

		if isLast, ok := result["is_last_package"].(bool); ok && isLast {
			break
		}
	}

	if finalText == "" {
		return "", fmt.Errorf("未在响应中找到识别结果")
	}
	return finalText, nil
}

func (s *STT) sendMessage(conn *websocket.Conn, messageType, flags, serialization byte, payload []byte, sequence int32) error {
	header := generateHeader(messageType, flags, serialization, GZIP_COMPRESSION, 0x00)
	beforePayload := generateBeforePayload(sequence)
	payloadSize := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSize, uint32(len(payload)))

	fullMessage := bytes.NewBuffer(header)
	fullMessage.Write(beforePayload)
	fullMessage.Write(payloadSize)
	fullMessage.Write(payload)

	return conn.WriteMessage(websocket.BinaryMessage, fullMessage.Bytes())
}

// gzipCompress 压缩数据
func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
} // 定义协议相关的常量和函数

const (
	PROTOCOL_VERSION    byte = 0x01
	DEFAULT_HEADER_SIZE byte = 0x01

	// 消息类型
	FULL_CLIENT_REQUEST   byte = 0x01
	AUDIO_ONLY_REQUEST    byte = 0x02
	FULL_SERVER_RESPONSE  byte = 0x09
	SERVER_ACK            byte = 0x0B
	SERVER_ERROR_RESPONSE byte = 0x0F

	POS_SEQUENCE      byte = 0x01
	NEG_SEQUENCE      byte = 0x02
	NEG_WITH_SEQUENCE byte = 0x03

	// 序列化方法
	NO_SERIALIZATION   byte = 0x00
	JSON_SERIALIZATION byte = 0x01

	// 压缩类型
	NO_COMPRESSION   byte = 0x00
	GZIP_COMPRESSION byte = 0x01
)

func generateHeader(
	messageType byte,
	messageTypeSpecificFlags byte,
	serialMethod byte,
	compressionType byte,
	reservedData byte,
) []byte {
	protocolVersion := PROTOCOL_VERSION
	headerSize := DEFAULT_HEADER_SIZE
	header := []byte{
		(protocolVersion << 4) | headerSize,
		(messageType << 4) | messageTypeSpecificFlags,
		(serialMethod << 4) | compressionType,
		reservedData,
	}
	return header
}

func generateBeforePayload(sequence int32) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, sequence)
	if err != nil {
		logger.Errorf("Error in generateBeforePayload: %v", err)
		return nil
	}
	return buf.Bytes()
}

func parseResponse(data []byte) (map[string]interface{}, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("响应数据过短")
	}
	//protocolVersion := data[0] >> 4
	headerSize := data[0] & 0x0F
	messageType := data[1] >> 4
	messageTypeSpecificFlags := data[1] & 0x0F
	serializationMethod := data[2] >> 4
	compressionType := data[2] & 0x0F
	// reserved := data[3]

	payloadData := data[headerSize*4:]

	result := make(map[string]interface{})
	result["is_last_package"] = false

	if messageTypeSpecificFlags&0x01 != 0 {
		// 带序列号的帧
		if len(payloadData) < 4 {
			return nil, fmt.Errorf("payload 长度不足以包含序列号")
		}
		var seq int32
		buf := bytes.NewReader(payloadData[:4])
		err := binary.Read(buf, binary.BigEndian, &seq)
		if err != nil {
			return nil, err
		}
		result["payload_sequence"] = seq
		payloadData = payloadData[4:]
	}

	if messageTypeSpecificFlags&0x02 != 0 {
		// 最后一个包
		result["is_last_package"] = true
	}

	var payloadMsg []byte
	var payloadSize uint32
	if messageType == FULL_SERVER_RESPONSE {
		if len(payloadData) < 4 {
			return nil, fmt.Errorf("payload 长度不足以包含大小信息")
		}
		payloadSize = binary.BigEndian.Uint32(payloadData[:4])
		payloadMsg = payloadData[4:]
	} else if messageType == SERVER_ACK {
		if len(payloadData) < 4 {
			return nil, fmt.Errorf("payload 长度不足以包含序列号")
		}
		var seq int32
		buf := bytes.NewReader(payloadData[:4])
		err := binary.Read(buf, binary.BigEndian, &seq)
		if err != nil {
			return nil, err
		}
		result["seq"] = seq
		if len(payloadData) >= 8 {
			payloadSize = binary.BigEndian.Uint32(payloadData[4:8])
			payloadMsg = payloadData[8:]
		}
	} else if messageType == SERVER_ERROR_RESPONSE {
		if len(payloadData) < 8 {
			return nil, fmt.Errorf("payload 长度不足以包含错误代码和大小信息")
		}
		code := binary.BigEndian.Uint32(payloadData[:4])
		result["code"] = code
		payloadSize = binary.BigEndian.Uint32(payloadData[4:8])
		payloadMsg = payloadData[8:]
	}

	if payloadMsg != nil {
		if compressionType == GZIP_COMPRESSION {
			gr, err := gzip.NewReader(bytes.NewReader(payloadMsg))
			if err != nil {
				return nil, err
			}
			decompressedData, err := io.ReadAll(gr)
			gr.Close()
			if err != nil {
				return nil, err
			}
			payloadMsg = decompressedData
		}

		if serializationMethod == JSON_SERIALIZATION {
			var payloadObj interface{}
			if err := json.Unmarshal(payloadMsg, &payloadObj); err != nil {
				return nil, err
			}
			result["payload_msg"] = payloadObj
		} else if serializationMethod != NO_SERIALIZATION {
			result["payload_msg"] = string(payloadMsg)
		}
		result["payload_size"] = payloadSize
	}

	// 打印解析后的响应内容
	logger.Infof("解析后的响应内容: %+v", result)

	return result, nil
}

func sliceData(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	dataLen := len(data)
	for i := 0; i < dataLen; i += chunkSize {
		end := i + chunkSize
		if end > dataLen {
			end = dataLen
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}
