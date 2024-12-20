package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

var (
	enumMessageType = map[byte]string{
		11: "audio-only server response",
		12: "frontend server response",
		15: "error message from server",
	}
	enumMessageTypeSpecificFlags = map[byte]string{
		0: "no sequence number",
		1: "sequence number > 0",
		2: "last message from server (seq < 0)",
		3: "sequence number < 0",
	}
	enumMessageSerializationMethods = map[byte]string{
		0:  "no serialization",
		1:  "JSON",
		15: "custom type",
	}
	enumMessageCompression = map[byte]string{
		0:  "no compression",
		1:  "gzip",
		15: "custom compression method",
	}
)

const (
	optQuery  string = "query"
	optSubmit string = "submit"
)

var addr = "openspeech.bytedance.com"
var u = url.URL{Scheme: "wss", Host: addr, Path: "/api/v1/tts/ws_binary"}

var appid = os.Getenv("VOICEFLOW_VOLCENGINE_TTS_APP_ID")
var token = os.Getenv("VOICEFLOW_VOLCENGINE_TTS_TOKEN")
var header = http.Header{"Authorization": []string{fmt.Sprintf("Bearer;%s", token)}}

type synResp struct {
	Audio  []byte
	IsLast bool
}

// version: b0001 (4 bits)
// header size: b0001 (4 bits)
// message type: b0001 (Full client request) (4bits)
// message type specific flags: b0000 (none) (4bits)
// message serialization method: b0001 (JSON) (4 bits)
// message compression: b0001 (gzip) (4bits)
// reserved data: 0x00 (1 byte)
var defaultHeader = []byte{0x11, 0x10, 0x11, 0x00}

func setupInput(text, voiceType, opt string) []byte {
	reqID := uuid.Must(uuid.NewV4(), nil).String()
	params := make(map[string]map[string]interface{})
	params["app"] = make(map[string]interface{})
	// 平台上查看具体appid
	params["app"]["appid"] = appid
	params["app"]["token"] = "N81BXUVgn7fJ2ySnNlHKEd2WLMVKmM" // 根据 API 文档设置正确的值
	// 平台上查看具体集群名称
	params["app"]["cluster"] = "volcano_tts"
	params["user"] = make(map[string]interface{})
	params["user"]["uid"] = "uid"
	params["audio"] = make(map[string]interface{})
	params["audio"]["voice_type"] = voiceType
	params["audio"]["encoding"] = "mp3"
	params["audio"]["speed_ratio"] = 1.0
	params["audio"]["volume_ratio"] = 1.0
	params["audio"]["pitch_ratio"] = 1.0
	params["request"] = make(map[string]interface{})
	params["request"]["reqid"] = reqID
	params["request"]["text"] = text
	params["request"]["text_type"] = "plain"
	params["request"]["operation"] = opt
	fmt.Println("Request parameters:", params)
	resStr, _ := json.Marshal(params)
	return resStr
}

func gzipCompress(input []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(input)
	w.Close()
	return b.Bytes()
}

func gzipDecompress(input []byte) []byte {
	b := bytes.NewBuffer(input)
	r, _ := gzip.NewReader(b)
	out, _ := ioutil.ReadAll(r)
	r.Close()
	return out
}

func parseResponse(res []byte) (resp synResp, err error) {
	protoVersion := res[0] >> 4
	headSize := res[0] & 0x0f
	messageType := res[1] >> 4
	messageTypeSpecificFlags := res[1] & 0x0f
	serializationMethod := res[2] >> 4
	messageCompression := res[2] & 0x0f
	reserve := res[3]
	headerExtensions := res[4 : headSize*4]
	payload := res[headSize*4:]

	fmt.Printf("            Protocol version: %x - version %d\n",
		protoVersion, protoVersion)
	fmt.Printf("                 Header size: %x - %d bytes\n",
		headSize, headSize*4)
	fmt.Printf("                Message type: %x - %s\n", messageType,
		enumMessageType[messageType])
	fmt.Printf(" Message type specific flags: %x - %s\n", messageTypeSpecificFlags,
		enumMessageTypeSpecificFlags[messageTypeSpecificFlags])
	fmt.Printf("Message serialization method: %x - %s\n",
		serializationMethod, enumMessageSerializationMethods[serializationMethod])
	fmt.Printf("         Message compression: %x - %s\n",
		messageCompression, enumMessageCompression[messageCompression])
	fmt.Printf("                    Reserved: %d\n", reserve)
	if headSize != 1 {
		fmt.Printf("           Header extensions: %s\n",
			headerExtensions)
	}
	// audio-only server response
	if messageType == 0xb {
		// no sequence number as ACK
		if messageTypeSpecificFlags == 0 {
			fmt.Println("                Payload size: 0")
		} else {
			sequenceNumber := int32(binary.BigEndian.Uint32(payload[0:4]))
			payloadSize := int32(binary.BigEndian.Uint32(payload[4:8]))
			payload = payload[8:]
			resp.Audio = append(resp.Audio, payload...)
			fmt.Printf("             Sequence number: %d\n",
				sequenceNumber)
			fmt.Printf("                Payload size: %d\n", payloadSize)
			if sequenceNumber < 0 {
				resp.IsLast = true
			}
		}
	} else if messageType == 0xf {
		code := int32(binary.BigEndian.Uint32(payload[0:4]))
		errMsg := payload[8:]
		if messageCompression == 1 {
			errMsg = gzipDecompress(errMsg)
		}
		fmt.Printf("                  Error code: %d\n", code)
		fmt.Printf("                   Error msg: %s\n", string(errMsg))
		err = errors.New(string(errMsg))
		return
	} else if messageType == 0xc {
		// msgSize = int32(binary.BigEndian.Uint32(payload[0:4]))
		payload = payload[4:]
		if messageCompression == 1 {
			payload = gzipDecompress(payload)
		}
		fmt.Printf("            Frontend message: %s\n", string(payload))
	} else {
		fmt.Printf("          wrong message type:%d\n", messageType)
		err = errors.New("wrong message type")
		return
	}
	return
}

// 一次性合成
func nonStreamSynth(text, voiceType, outFile string) {
	input := setupInput(text, voiceType, optQuery)
	fmt.Println(string(input))
	input = gzipCompress(input)
	payloadSize := len(input)
	payloadArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadArr, uint32(payloadSize))
	clientRequest := make([]byte, len(defaultHeader))
	copy(clientRequest, defaultHeader)
	clientRequest = append(clientRequest, payloadArr...)
	clientRequest = append(clientRequest, input...)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		fmt.Println("dial err:", err)
		return
	}
	defer c.Close()
	err = c.WriteMessage(websocket.BinaryMessage, clientRequest)
	if err != nil {
		fmt.Println("write message fail, err:", err.Error())
		return
	}
	_, message, err := c.ReadMessage()
	if err != nil {
		fmt.Println("read message fail, err:", err.Error())
		return
	}
	resp, err := parseResponse(message)
	if err != nil {
		fmt.Println("parse response fail, err:", err.Error())
		return
	}
	err = ioutil.WriteFile(outFile, resp.Audio, 0644)
	if err != nil {
		fmt.Println("write audio to fail fail, err:", err.Error())
		return
	}
}

// 流式合成
func streamSynth(text, voiceType, outFile string) {
	input := setupInput(text, voiceType, optSubmit)
	fmt.Println(string(input))
	input = gzipCompress(input)
	payloadSize := len(input)
	payloadArr := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadArr, uint32(payloadSize))
	clientRequest := make([]byte, len(defaultHeader))
	copy(clientRequest, defaultHeader)
	clientRequest = append(clientRequest, payloadArr...)
	clientRequest = append(clientRequest, input...)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		fmt.Println("dial err:", err)
		return
	}
	defer c.Close()
	err = c.WriteMessage(websocket.BinaryMessage, clientRequest)
	if err != nil {
		fmt.Println("write message fail, err:", err.Error())
		return
	}
	var audio []byte
	for {
		var message []byte
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read message fail, err:", err.Error())
			break
		}
		resp, err := parseResponse(message)
		if err != nil {
			fmt.Println("parse response fail, err:", err.Error())
			break
		}
		audio = append(audio, resp.Audio...)
		if resp.IsLast {
			break
		}
	}
	if err != nil {
		fmt.Println("stream synthesis fail, err:", err.Error())
		return
	}
	err = ioutil.WriteFile(outFile, audio, 0644)
	if err != nil {
		fmt.Println("write audio to fail fail, err:", err.Error())
		return
	}
}

func main() {
	fmt.Println("appid:", appid)
	fmt.Println("token:", token)
	// 此处替换成需要调用的音色
	streamSynth("我想测试下语音合成的效果", "zh_male_beijingxiaoye_moon_bigtts", "test.mp3")
}
