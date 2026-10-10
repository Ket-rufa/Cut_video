package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/video_cutter?sslmode=disable")
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
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	_ = os.WriteFile("models_dump.json", b, 0644)
	fmt.Println("Written bytes:", len(b))
}
