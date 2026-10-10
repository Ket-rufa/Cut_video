package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func generateSSML(text, voice string) string {
	return fmt.Sprintf(`<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='vi-VN'>
<voice name='%s'>
<prosody pitch='+0Hz' rate='+0%%' volume='+0%%'>
%s
</prosody>
</voice>
</speak>`, voice, text)
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func synthesizeEdgeTTS(text, voice string) ([]byte, error) {
	connID := randomHex(16)
	url := fmt.Sprintf("wss://speech.platform.bing.com/consumer/speech/synthesize/readaloud/edge/v1?TrustedClientToken=6A5AA1D4EA654070A64E8512D3E42A69&ConnectionId=%s", connID)

	header := http.Header{}
	header.Set("Origin", "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold")
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36 Edg/130.0.0.0")
	header.Set("Pragma", "no-cache")
	header.Set("Cache-Control", "no-cache")

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		return nil, fmt.Errorf("dial error: %w", err)
	}
	defer conn.Close()

	// 1. Send speech.config
	timestamp := time.Now().UTC().Format("Mon Jan 02 2006 15:04:05 GMT-0700 (Coordinated Universal Time)")
	configMsg := fmt.Sprintf("X-Timestamp:%s\r\nContent-Type:application/json; charset=utf-8\r\nPath:speech.config\r\n\r\n{\"context\":{\"synthesis\":{\"audio\":{\"metadataoptions\":{\"sentenceBoundaryEnabled\":\"false\",\"wordBoundaryEnabled\":\"false\"},\"outputFormat\":\"audio-24khz-48kbitrate-mono-mp3\"}}}}", timestamp)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(configMsg)); err != nil {
		return nil, fmt.Errorf("write config error: %w", err)
	}

	// 2. Send SSML request
	reqID := randomHex(16)
	ssml := generateSSML(text, voice)
	ssmlMsg := fmt.Sprintf("X-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:%s\r\nPath:ssml\r\n\r\n%s", reqID, timestamp, ssml)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(ssmlMsg)); err != nil {
		return nil, fmt.Errorf("write ssml error: %w", err)
	}

	// 3. Receive audio data
	var audioBuf bytes.Buffer
	for {
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if msgType == websocket.TextMessage {
			str := string(message)
			if strings.Contains(str, "Path:turn.end") {
				break
			}
		} else if msgType == websocket.BinaryMessage {
			// Binary message contains 2 bytes header length, then text header, then audio binary
			if len(message) > 2 {
				headerLen := int(message[0])<<8 | int(message[1])
				if len(message) > 2+headerLen {
					audioData := message[2+headerLen:]
					audioBuf.Write(audioData)
				}
			}
		}
	}

	if audioBuf.Len() == 0 {
		return nil, fmt.Errorf("no audio received")
	}

	return audioBuf.Bytes(), nil
}

func main() {
	text := "Xin chào bạn, hôm nay tôi sẽ review bộ phim vô cùng kịch tính này."
	
	// Test female voice (HoaiMy)
	fmt.Println("Testing vi-VN-HoaiMyNeural...")
	audio1, err := synthesizeEdgeTTS(text, "vi-VN-HoaiMyNeural")
	if err != nil {
		fmt.Printf("HoaiMy error: %v\n", err)
	} else {
		fmt.Printf("HoaiMy SUCCESS! Bytes: %d\n", len(audio1))
		os.WriteFile("test_hoaimy.mp3", audio1, 0644)
	}

	// Test male voice (NamMinh)
	fmt.Println("Testing vi-VN-NamMinhNeural...")
	audio2, err := synthesizeEdgeTTS(text, "vi-VN-NamMinhNeural")
	if err != nil {
		fmt.Printf("NamMinh error: %v\n", err)
	} else {
		fmt.Printf("NamMinh SUCCESS! Bytes: %d\n", len(audio2))
		os.WriteFile("test_namminh.mp3", audio2, 0644)
	}
}
