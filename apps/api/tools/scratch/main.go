package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "modernc.org/sqlite"
)

type M struct {
	Name                       string   `json:"name"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
}
type Resp struct {
	Models []M `json:"models"`
}

func main() {
	db, err := sql.Open("sqlite", "social_video_cutter.db")
	if err != nil {
		fmt.Println("DB error:", err)
		return
	}
	defer db.Close()

	var key string
	err = db.QueryRow("SELECT value FROM app_settings WHERE key = 'gemini_api_key'").Scan(&key)
	if err != nil {
		fmt.Println("Query error:", err)
		return
	}

	resp, err := http.Get("https://generativelanguage.googleapis.com/v1beta/models?key=" + key)
	if err != nil {
		fmt.Println("HTTP error:", err)
		return
	}
	defer resp.Body.Close()

	var r Resp
	json.NewDecoder(resp.Body).Decode(&r)
	for _, m := range r.Models {
		for _, sm := range m.SupportedGenerationMethods {
			if sm == "generateContent" {
				fmt.Println(m.Name)
			}
		}
	}
}
