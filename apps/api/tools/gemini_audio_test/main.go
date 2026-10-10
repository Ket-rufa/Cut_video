package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "d:/CUT_VIDEO_MXH/apps/api/storage/dev.db")
	if err != nil {
		fmt.Printf("DB error: %v\n", err)
		return
	}
	defer db.Close()

	var apiKey string
	err = db.QueryRow("SELECT value FROM app_settings WHERE key = 'gemini_api_key'").Scan(&apiKey)
	if err != nil {
		fmt.Printf("Get key error: %v\n", err)
		return
	}
	fmt.Printf("Key found (len=%d)\n", len(apiKey))

	// Test audio generation via Gemini generateContent with responseModalities: ["AUDIO"]
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]interface{}{
					{"text": "Đọc to câu sau bằng tiếng Việt: Xin chào, hôm nay tôi sẽ review bộ phim này."},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseModalities": []string{"AUDIO"},
			"speechConfig": map[string]interface{}{
				"voiceConfig": map[string]interface{}{
					"prebuiltVoiceConfig": map[string]interface{}{
						"voiceName": "Puck",
					},
				},
			},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Printf("Post error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\nResponse: %s\n", resp.StatusCode, string(respBytes)[:min(len(respBytes), 500)])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
