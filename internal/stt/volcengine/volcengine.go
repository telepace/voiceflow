// volcengine.go
package volcengine

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/telepace/voiceflow/pkg/config"
	"github.com/telepace/voiceflow/pkg/logger"
)

type STT struct {
	wsURL      string
	uid        string
	rate       int
	format     string
	bits       int
	channel    int
	codec      string
	accessKey  string
	appKey     string
	resourceID string
}

func NewVolcengineSTT() *STT {
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Fatalf("配置初始化失败: %v", err)
	}

	sttCfg := cfg.Volcengine.STT
	return &STT{
		wsURL:      sttCfg.WsURL,
		uid:        sttCfg.UID,
		rate:       sttCfg.Rate,
		format:     sttCfg.Format,
		bits:       sttCfg.Bits,
		channel:    sttCfg.Channel,
		codec:      sttCfg.Codec,
		accessKey:  sttCfg.AccessKey,
		appKey:     sttCfg.AppKey,
		resourceID: sttCfg.ResourceID,
	}
}

func (s *STT) validateAudioFormat(audioData []byte) error {
	if len(audioData) == 0 {
		return fmt.Errorf("音频数据为空")
	}

	// 添加其他音频格式检查逻辑
	// 例如检查采样率、位深度等

	return nil
}

// isFinalResponse 判断响应是否为最终响应
func isFinalResponse(result map[string]interface{}) bool {
	// 根据 VolcEngine 的响应结构调整逻辑
	if isFinal, ok := result["is_final"].(bool); ok {
		return isFinal
	}
	return false
}

// Recognize 调用 VolcEngine 的 STT API 将音频数据转换为文本
// 新增 audioURL 参数，但 VolcEngine 不使用该参数
func (s *STT) Recognize(audioData []byte, audioURL string) (string, error) {
	if audioURL != "" {
		logger.Infof("VolcEngine STT 不支持使用 audioURL，忽略该参数")
	}

	// 添加音频格式验证
	if err := s.validateAudioFormat(audioData); err != nil {
		logger.Errorf("音频格式验证失败: %v", err)
		return "", err
	}

	// reqID := uuid.New().String()
	connectID := uuid.New().String()

	header := http.Header{}
	header.Set("X-Api-Access-Key", s.accessKey)
	header.Set("X-Api-App-Key", s.appKey)
	header.Set("X-Api-Resource-Id", s.resourceID)
	// header.Set("X-Api-Request-Id", reqID)
	header.Set("X-Api-Connect-Id", connectID)

	logger.Infof("连接到 WebSocket URL: %s", s.wsURL)
	logger.Infof("请求头: %v", header)

	dialer := websocket.DefaultDialer
	conn, resp, err := dialer.Dial(s.wsURL, header)
	if err != nil {
		logger.Errorf("WebSocket 连接错误: %v", err)
		return "", err
	}
	defer conn.Close()

	// 检查并打印 X-Api-Connect-Id 和 X-Tt-Logid
	if connectID := resp.Header.Get("X-Api-Connect-Id"); connectID != "" {
		logger.Infof("连接追踪ID: X-Api-Connect-Id = %s", connectID)
	}
	if logID := resp.Header.Get("X-Tt-Logid"); logID != "" {
		logger.Infof("服务端返回的 logid: X-Tt-Logid = %s", logID)
	}

	// 构建并发送初始请求
	req := map[string]interface{}{
		"user": map[string]interface{}{
			"uid": s.uid,
		},
		"audio": map[string]interface{}{
			"format":   s.format,
			"rate":     s.rate,
			"bits":     s.bits,
			"channel":  s.channel,
			"codec":    s.codec,
			"language": "zh-CN",
		},
		"request": map[string]interface{}{
			"model_name":      "bigmodel",
			"enable_itn":      false,
			"enable_punc":     false,
			"enable_ddc":      false,
			"show_utterances": false,
			"result_type":     "full",
		},
	}

	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	// 不使用压缩，直接发送
	fullClientRequest := generateHeader(
		FULL_CLIENT_REQUEST,
		NOT_LAST_PACKAGE_NO_SEQUENCE,
		JSON_SERIALIZATION,
		NO_COMPRESSION,
		0x00,
	)
	payloadSize := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadSize, uint32(len(payloadBytes)))

	message := append(fullClientRequest, payloadSize...)
	message = append(message, payloadBytes...)

	err = conn.WriteMessage(websocket.BinaryMessage, message)
	if err != nil {
		logger.Errorf("发送初始消息错误: %v", err)
		return "", err
	}

	// 接收服务器的初始响应
	_, resData, err := conn.ReadMessage()
	if err != nil {
		logger.Errorf("读取响应错误: %v", err)
		return "", err
	}

	result, err := parseResponse(resData)
	if err != nil {
		logger.Errorf("解析响应错误: %v", err)
		return "", err
	}

	if errCode, ok := result["error_code"]; ok {
		logger.Errorf("服务器返回错误码 %v: %v", errCode, result["error_msg"])
		return "", fmt.Errorf("服务器返回错误码 %v: %v", errCode, result["error_msg"])
	}

	logger.Infof("初始响应: %+v", result)

	// 发送音频数据
	// 将音频数据按照块大小分片发送
	chunkSize := 3200 // 根据需求调整，每个包的音频时长约 100ms（16kHz 采样率，16 位深度，单声道）
	audioChunks := sliceData(audioData, chunkSize)

	for i, chunk := range audioChunks {
		isLast := i == len(audioChunks)-1

		flags := NOT_LAST_PACKAGE_NO_SEQUENCE
		if isLast {
			flags = LAST_PACKAGE_NO_SEQUENCE
		}

		audioRequest := generateHeader(
			AUDIO_ONLY_REQUEST,
			flags,
			NO_SERIALIZATION,
			NO_COMPRESSION,
			0x00,
		)

		payloadSize := make([]byte, 4)
		binary.BigEndian.PutUint32(payloadSize, uint32(len(chunk)))

		message := append(audioRequest, payloadSize...)
		message = append(message, chunk...)

		err = conn.WriteMessage(websocket.BinaryMessage, message)
		if err != nil {
			logger.Errorf("发送音频数据错误: %v", err)
			return "", err
		}

		logger.Debugf("发送音频数据包 %d", i+1)

		// 接收服务器响应
		if !isLast {
			// 非最后一包，尝试读取中间结果
			conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			_, resData, err = conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					logger.Errorf("读取响应错误: %v", err)
					return "", err
				} else {
					// 超时或非致命错误，继续发送
					continue
				}
			}

			result, err = parseResponse(resData)
			if err != nil {
				logger.Errorf("解析响应错误: %v", err)
				continue
			}

			if errCode, ok := result["error_code"]; ok {
				logger.Errorf("服务器返回错误码 %v: %v", errCode, result["error_msg"])
				return "", fmt.Errorf("服务器返回错误码 %v: %v", errCode, result["error_msg"])
			}

			logger.Infof("中间响应: %+v", result)
		}
	}

	// 接收服务器的最终响应
	var finalText string
	for {
		_, resData, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				logger.Errorf("读取最终响应错误: %v", err)
				return "", err
			}
			break
		}

		var result map[string]interface{}
		err = json.Unmarshal(resData, &result)
		if err != nil {
			logger.Errorf("解析最终响应错误: %v", err)
			continue
		}

		if errCode, ok := result["error_code"]; ok {
			logger.Errorf("服务器返回错误码 %v: %v", errCode, result["error_msg"])
			return "", fmt.Errorf("服务器返回错误码 %v: %v", errCode, result["error_msg"])
		}

		if text, ok := result["text"].(string); ok {
			finalText += text
			logger.Infof("识别文本: %s", text)
		}

		// 使用定义的 isFinalResponse 函数判断是否为最终响应
		if isFinalResponse(result) {
			break
		}
	}

	return finalText, nil
}

// 定义协议相关的常量和函数
const (
	PROTOCOL_VERSION    byte = 0x01
	DEFAULT_HEADER_SIZE byte = 0x01

	// 消息类型
	FULL_CLIENT_REQUEST   byte = 0x01
	AUDIO_ONLY_REQUEST    byte = 0x02
	FULL_SERVER_RESPONSE  byte = 0x09
	SERVER_ERROR_RESPONSE byte = 0x0F

	// Message Type Specific Flags
	NOT_LAST_PACKAGE_NO_SEQUENCE byte = 0x00
	LAST_PACKAGE_NO_SEQUENCE     byte = 0x02

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

	headerLength := int(headerSize) * 4
	if len(data) < headerLength {
		return nil, fmt.Errorf("数据长度不足以包含完整的头部")
	}

	payload := data[headerLength:]

	result := make(map[string]interface{})

	if messageType == FULL_SERVER_RESPONSE {
		if len(payload) < 8 {
			return nil, fmt.Errorf("payload 长度不足以包含序列号和大小信息")
		}
		sequence := binary.BigEndian.Uint32(payload[0:4])
		payloadSize := binary.BigEndian.Uint32(payload[4:8])

		if len(payload) < int(8+payloadSize) {
			return nil, fmt.Errorf("payload 长度不足以包含完整的消息")
		}

		payloadMsg := payload[8 : 8+payloadSize]

		if compressionType == GZIP_COMPRESSION {
			// 本例中不使用压缩，保留此代码以备后用
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

		result["sequence"] = sequence

		if messageTypeSpecificFlags&0x02 != 0 {
			result["is_last_package"] = true
		} else {
			result["is_last_package"] = false
		}
	} else if messageType == SERVER_ERROR_RESPONSE {
		// 解析错误响应
		if len(payload) < 8 {
			return nil, fmt.Errorf("payload 长度不足以包含错误代码和大小信息")
		}
		errorCode := binary.BigEndian.Uint32(payload[:4])
		errorMsgSize := binary.BigEndian.Uint32(payload[4:8])
		if len(payload) < int(8+errorMsgSize) {
			return nil, fmt.Errorf("payload 长度不足以包含完整的错误消息")
		}
		errorMsg := payload[8 : 8+errorMsgSize]

		if compressionType == GZIP_COMPRESSION {
			// 本例中不使用压缩，保留此代码以备后用
		}

		result["error_code"] = errorCode
		result["error_msg"] = string(errorMsg)
	} else {
		logger.Warn("收到未知的消息类型")
	}

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
