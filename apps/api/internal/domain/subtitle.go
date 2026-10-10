package domain

import "time"

type SubtitleCue struct {
	Index   int    `json:"index"`
	StartMs int64  `json:"start_ms"`
	EndMs   int64  `json:"end_ms"`
	Text    string `json:"text"`
}

type VideoSubtitle struct {
	ID             string        `db:"id" json:"id"`
	VideoID        string        `db:"video_id" json:"video_id"`
	LanguageCode   string        `db:"language_code" json:"language_code"`
	SourceLanguage string        `db:"source_language" json:"source_language"`
	Format         string        `db:"format" json:"format"`
	Content        string        `db:"content" json:"content"`
	CuesJSON       string        `db:"cues_json" json:"-"`
	Cues           []SubtitleCue `db:"-" json:"cues"`
	CreatedAt      time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at" json:"updated_at"`
}

type AITranslationResult struct {
	SourceLanguage       string   `json:"source_language"`
	TargetLanguage       string   `json:"target_language"`
	OriginalTitle        string   `json:"original_title"`
	TranslatedTitle      string   `json:"translated_title"`
	OriginalDescription  string   `json:"original_description"`
	TranslatedDescription string  `json:"translated_description"`
	OriginalTags         []string `json:"original_tags"`
	TranslatedTags       []string `json:"translated_tags"`
}

type VideoDubbing struct {
	VideoID       string    `json:"video_id"`
	AudioURL      string    `json:"audio_url"`
	VideoURL      string    `json:"video_url,omitempty"`
	CueCount      int       `json:"cue_count"`
	StartMs       int64     `json:"start_ms"`
	EndMs         int64     `json:"end_ms"`
	DurationMs    int64     `json:"duration_ms"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
