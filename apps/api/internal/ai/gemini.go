package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"social-video-cutter/api/internal/config"
	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/repository"
)

type GeminiService struct {
	cfg        *config.Config
	subRepo    repository.SubtitleRepository
	httpClient *http.Client
}

func NewGeminiService(cfg *config.Config, subRepo repository.SubtitleRepository) *GeminiService {
	return &GeminiService{
		cfg:        cfg,
		subRepo:    subRepo,
		httpClient: &http.Client{Timeout: 3 * time.Minute},
	}
}

func (s *GeminiService) getEffectiveAPIKey(ctx context.Context) string {
	// First check DB setting
	if s.subRepo != nil {
		val, err := s.subRepo.GetSetting(ctx, "gemini_api_key")
		if err == nil && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	// Fallback to environment/config
	return strings.TrimSpace(s.cfg.AI.GeminiAPIKey)
}

func (s *GeminiService) GetCurrentAPIKey(ctx context.Context) string {
	return s.getEffectiveAPIKey(ctx)
}

func (s *GeminiService) SaveAPIKey(ctx context.Context, apiKey string) error {
	if s.subRepo == nil {
		return fmt.Errorf("subtitle repository not available")
	}
	return s.subRepo.SetSetting(ctx, "gemini_api_key", strings.TrimSpace(apiKey))
}

type geminiModelItem struct {
	Name                       string   `json:"name"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
}

type geminiModelsResponse struct {
	Models []geminiModelItem `json:"models"`
}

func (s *GeminiService) resolveWorkingModel(ctx context.Context, apiKey string) string {
	if s.cfg.AI.Model != "" && s.cfg.AI.Model != "gemini-1.5-flash" && s.cfg.AI.Model != "gemini-3.8-flash" && s.cfg.AI.Model != "gemini-flash-latest" {
		return s.cfg.AI.Model
	}

	// Query available models from Google AI API
	listURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err == nil {
		resp, err := s.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var res geminiModelsResponse
			if err := json.NewDecoder(resp.Body).Decode(&res); err == nil {
				// Priority candidates: gemini-3.5-flash is current best with generous free quota!
				candidates := []string{
					"gemini-3.5-flash",
					"gemini-3.5-transcribe",
					"gemini-3.5-flash-lite",
					"gemini-flash-lite-latest",
					"gemini-3.7-flash",
				}
				for _, cand := range candidates {
					for _, m := range res.Models {
						cleanName := strings.TrimPrefix(m.Name, "models/")
						if cleanName == cand {
							for _, method := range m.SupportedGenerationMethods {
								if method == "generateContent" {
									return cand
								}
							}
						}
					}
				}
				// Any flash model supporting generateContent
				for _, m := range res.Models {
					cleanName := strings.TrimPrefix(m.Name, "models/")
					if strings.Contains(cleanName, "3.5") && strings.Contains(cleanName, "flash") && !strings.Contains(cleanName, "tts") {
						for _, method := range m.SupportedGenerationMethods {
							if method == "generateContent" {
								return cleanName
							}
						}
					}
				}
			}
		}
	}

	return "gemini-3.5-flash"
}

func (s *GeminiService) TestConnection(ctx context.Context, apiKey string) (bool, string, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		key = s.getEffectiveAPIKey(ctx)
	}
	if key == "" {
		return false, "", fmt.Errorf("chưa cấu hình Gemini API Key")
	}

	model := s.resolveWorkingModel(ctx, key)

	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, key)
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": "Hãy trả lời đúng 2 từ: 'KẾT NỐI_OK'"},
				},
			},
		},
	}

	jsonBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("kết nối Google API thất bại: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("lỗi Gemini API (HTTP %d, Model: %s): %s", resp.StatusCode, model, string(bodyBytes))
	}

	return true, fmt.Sprintf("Kết nối Google Gemini thành công! Đang sử dụng mô hình %s.", model), nil
}

// TranslateMetadata translates title, description and tags from any detected language to Vietnamese.
func (s *GeminiService) TranslateMetadata(ctx context.Context, title, desc string, tags []string) (*domain.AITranslationResult, error) {
	apiKey := s.getEffectiveAPIKey(ctx)

	// If Gemini key is available, use Gemini 1.5 Flash for nuanced social media translation
	if apiKey != "" {
		result, err := s.translateMetadataViaGemini(ctx, apiKey, title, desc, tags)
		if err == nil && result != nil {
			return result, nil
		}
	}

	// Fallback: Use free Google Translate API so user always gets instant translation even with no key
	return s.translateMetadataViaGoogleFallback(ctx, title, desc, tags)
}

func (s *GeminiService) translateMetadataViaGemini(ctx context.Context, apiKey, title, desc string, tags []string) (*domain.AITranslationResult, error) {
	model := s.resolveWorkingModel(ctx, apiKey)

	prompt := fmt.Sprintf(`Bạn là một biên tập viên và dịch giả video mạng xã hội chuyên nghiệp (TikTok, Douyin, YouTube, Facebook).
Hãy tự động phát hiện ngôn ngữ gốc (ví dụ: tiếng Trung, tiếng Anh, v.v.) và dịch tiêu đề, mô tả, các thẻ hashtag sau đây sang tiếng Việt tự nhiên, hấp dẫn, đúng văn phong mạng xã hội và phim ảnh Việt Nam.

Dữ liệu đầu vào:
Tiêu đề: %s
Mô tả: %s
Thẻ (hashtags): %s

YÊU CẦU QUAN TRỌNG:
Chỉ trả về DUY NHẤT một chuỗi JSON hợp lệ (không kèm markdown, không kèm giải thích ngoài JSON) theo cấu trúc sau:
{
  "source_language": "Tên ngôn ngữ gốc phát hiện (ví dụ: Tiếng Trung, Tiếng Anh)",
  "translated_title": "Tiêu đề tiếng Việt",
  "translated_description": "Mô tả tiếng Việt",
  "translated_tags": ["tag1", "tag2", "tag3"]
}`, title, desc, strings.Join(tags, ", "))

	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":      0.2,
			"responseMimeType": "application/json",
		},
	}

	jsonBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API error (HTTP %d): %s", resp.StatusCode, string(b))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("không nhận được phản hồi từ Gemini")
	}

	rawText := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	// Strip markdown if present
	rawText = strings.TrimPrefix(rawText, "```json")
	rawText = strings.TrimPrefix(rawText, "```")
	rawText = strings.TrimSuffix(rawText, "```")
	rawText = strings.TrimSpace(rawText)

	var parsed struct {
		SourceLanguage        string   `json:"source_language"`
		TranslatedTitle       string   `json:"translated_title"`
		TranslatedDescription string   `json:"translated_description"`
		TranslatedTags        []string `json:"translated_tags"`
	}

	if err := json.Unmarshal([]byte(rawText), &parsed); err != nil {
		return nil, fmt.Errorf("lỗi đọc JSON trả về từ AI: %w", err)
	}

	return &domain.AITranslationResult{
		SourceLanguage:        parsed.SourceLanguage,
		TargetLanguage:        "Tiếng Việt",
		OriginalTitle:         title,
		TranslatedTitle:       parsed.TranslatedTitle,
		OriginalDescription:   desc,
		TranslatedDescription: parsed.TranslatedDescription,
		OriginalTags:          tags,
		TranslatedTags:        parsed.TranslatedTags,
	}, nil
}

func (s *GeminiService) translateMetadataViaGoogleFallback(ctx context.Context, title, desc string, tags []string) (*domain.AITranslationResult, error) {
	translatedTitle, detectedLang, _ := s.quickTranslateGoogle(ctx, title)
	translatedDesc, _, _ := s.quickTranslateGoogle(ctx, desc)
	var translatedTags []string
	for _, t := range tags {
		trTag, _, err := s.quickTranslateGoogle(ctx, t)
		if err == nil && trTag != "" {
			trTag = strings.ReplaceAll(trTag, " ", "")
			trTag = strings.ReplaceAll(trTag, "#", "")
			translatedTags = append(translatedTags, trTag)
		} else {
			cleanT := strings.ReplaceAll(t, "#", "")
			translatedTags = append(translatedTags, cleanT)
		}
	}

	langLabel := "Tiếng Trung"
	if strings.HasPrefix(detectedLang, "zh") {
		langLabel = "Tiếng Trung (Trung Quốc)"
	} else if strings.HasPrefix(detectedLang, "en") {
		langLabel = "Tiếng Anh"
	} else if strings.HasPrefix(detectedLang, "ja") {
		langLabel = "Tiếng Nhật"
	} else if strings.HasPrefix(detectedLang, "ko") {
		langLabel = "Tiếng Hàn"
	} else if detectedLang != "" {
		langLabel = detectedLang
	}

	return &domain.AITranslationResult{
		SourceLanguage:        langLabel,
		TargetLanguage:        "Tiếng Việt",
		OriginalTitle:         title,
		TranslatedTitle:       translatedTitle,
		OriginalDescription:   desc,
		TranslatedDescription: translatedDesc,
		OriginalTags:          tags,
		TranslatedTags:        translatedTags,
	}, nil
}

func (s *GeminiService) quickTranslateGoogle(ctx context.Context, text string) (string, string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", "", nil
	}
	urlStr := fmt.Sprintf("https://translate.googleapis.com/translate_a/single?client=gtx&sl=auto&tl=vi&dt=t&q=%s", url.QueryEscape(text))
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return text, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return text, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return text, "", fmt.Errorf("translate status: %d", resp.StatusCode)
	}

	var data []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return text, "", err
	}

	detectedLang := "zh-CN"
	if len(data) > 2 {
		if l, ok := data[2].(string); ok && l != "" {
			detectedLang = l
		}
	}

	if len(data) > 0 {
		if sentences, ok := data[0].([]interface{}); ok {
			var sb strings.Builder
			for _, item := range sentences {
				if sItem, ok := item.([]interface{}); ok && len(sItem) > 0 {
					if piece, ok := sItem[0].(string); ok {
						sb.WriteString(piece)
					}
				}
			}
			res := strings.TrimSpace(sb.String())
			if res != "" {
				return res, detectedLang, nil
			}
		}
	}
	return text, detectedLang, nil
}

// TranslateTextToVietnamese translates text to Vietnamese, using Gemini AI first or Google Translate POST fallback.
func (s *GeminiService) TranslateTextToVietnamese(ctx context.Context, text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil
	}

	apiKey := s.getEffectiveAPIKey(ctx)
	if apiKey != "" {
		tr, err := s.translateTextViaGemini(ctx, apiKey, text)
		if err == nil && strings.TrimSpace(tr) != "" {
			return strings.TrimSpace(tr), nil
		}
	}

	// Fallback to Google Translate POST (handles arbitrary long texts without URI limit)
	tr, err := s.quickTranslateGooglePOST(ctx, text)
	if err == nil && strings.TrimSpace(tr) != "" {
		return strings.TrimSpace(tr), nil
	}

	return text, err
}

func (s *GeminiService) translateTextViaGemini(ctx context.Context, apiKey, text string) (string, error) {
	model := s.resolveWorkingModel(ctx, apiKey)
	prompt := fmt.Sprintf(`Bạn là biên tập viên dịch thuật video mạng xã hội chuyên nghiệp.
Hãy dịch đoạn văn bản sau đây sang Tiếng Việt tự nhiên, chuẩn phong cách mạng xã hội và phim ảnh Việt Nam.
BẮT BUỘC: Chỉ trả về DUY NHẤT nội dung bản dịch tiếng Việt thuần túy, không thêm lời chào, không giải thích, không đặt trong dấu ngoặc kép hay markdown.

Nội dung cần dịch:
%s`, text)

	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.2,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API error (HTTP %d): %s", resp.StatusCode, string(b))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("không nhận được phản hồi từ Gemini")
	}

	rawText := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	return rawText, nil
}

func (s *GeminiService) quickTranslateGooglePOST(ctx context.Context, text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil
	}

	postData := url.Values{}
	postData.Set("client", "gtx")
	postData.Set("sl", "auto")
	postData.Set("tl", "vi")
	postData.Set("dt", "t")
	postData.Set("q", text)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://translate.googleapis.com/translate_a/single", strings.NewReader(postData.Encode()))
	if err != nil {
		return text, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return text, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return text, fmt.Errorf("translate status: %d", resp.StatusCode)
	}

	var data []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return text, err
	}

	if len(data) > 0 {
		if sentences, ok := data[0].([]interface{}); ok {
			var sb strings.Builder
			for _, item := range sentences {
				if sItem, ok := item.([]interface{}); ok && len(sItem) > 0 {
					if piece, ok := sItem[0].(string); ok {
						sb.WriteString(piece)
					}
				}
			}
			res := strings.TrimSpace(sb.String())
			if res != "" {
				return res, nil
			}
		}
	}

	return text, nil
}

// ContainsHanzi checks if a string contains any Chinese / Han characters.
func ContainsHanzi(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// SplitLongCueIfNeeded splits a cue if its duration is overly long (> 10s) and contains multiple sentences.
func SplitLongCueIfNeeded(cue domain.SubtitleCue) []domain.SubtitleCue {
	dur := cue.EndMs - cue.StartMs
	if dur <= 7000 {
		return []domain.SubtitleCue{cue}
	}

	// Split by punctuation: Vietnamese/English/Chinese sentence ends
	re := regexp.MustCompile(`([.!?…;。！？]+|\n)`)
	rawParts := re.Split(cue.Text, -1)
	var sentences []string
	for _, p := range rawParts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			sentences = append(sentences, trimmed)
		}
	}

	if len(sentences) <= 1 {
		words := strings.Fields(cue.Text)
		targetParts := int((dur + 5999) / 6000)
		if targetParts > len(words) {
			targetParts = len(words)
		}
		if targetParts <= 1 {
			cue.EndMs = cue.StartMs + min(dur, 7000)
			return []domain.SubtitleCue{cue}
		}
		sentences = make([]string, 0, targetParts)
		for i := 0; i < targetParts; i++ {
			from := i * len(words) / targetParts
			to := (i + 1) * len(words) / targetParts
			sentences = append(sentences, strings.Join(words[from:to], " "))
		}
	}

	totalRunes := 0
	for _, s := range sentences {
		totalRunes += len([]rune(s))
	}
	if totalRunes == 0 {
		return []domain.SubtitleCue{cue}
	}

	var result []domain.SubtitleCue
	curStart := cue.StartMs
	for i, s := range sentences {
		sRunes := len([]rune(s))
		var pieceDur int64
		if i == len(sentences)-1 {
			pieceDur = cue.EndMs - curStart
		} else {
			pieceDur = int64(float64(dur) * (float64(sRunes) / float64(totalRunes)))
			if pieceDur < 1500 {
				pieceDur = 1500
			}
		}
		curEnd := curStart + pieceDur
		if curEnd > cue.EndMs {
			curEnd = cue.EndMs
		}

		result = append(result, domain.SubtitleCue{
			StartMs: curStart,
			EndMs:   curEnd,
			Text:    s,
		})
		curStart = curEnd
	}

	var normalized []domain.SubtitleCue
	for _, part := range result {
		if part.EndMs-part.StartMs > 7000 {
			normalized = append(normalized, SplitLongCueIfNeeded(part)...)
			continue
		}
		normalized = append(normalized, part)
	}
	return normalized
}

func (s *GeminiService) probeVideoDuration(ctx context.Context, videoAbsPath string) int64 {
	cmd := exec.CommandContext(ctx, s.cfg.Downloader.FFprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoAbsPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	durSec, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0
	}
	return int64(durSec * 1000)
}

func (s *GeminiService) extractAudioSegment(ctx context.Context, videoAbsPath, audioPath string, startMs, endMs int64) error {
	ffmpegArgs := []string{"-y"}
	if startMs > 0 {
		startSec := float64(startMs) / 1000.0
		ffmpegArgs = append(ffmpegArgs, "-ss", fmt.Sprintf("%.3f", startSec))
	}
	ffmpegArgs = append(ffmpegArgs, "-i", videoAbsPath)
	if endMs > startMs {
		durSec := float64(endMs-startMs) / 1000.0
		ffmpegArgs = append(ffmpegArgs, "-t", fmt.Sprintf("%.3f", durSec))
	}
	ffmpegArgs = append(ffmpegArgs,
		"-vn",
		"-ac", "1",
		"-ar", "16000",
		"-b:a", "32k",
		audioPath,
	)

	cmd := exec.CommandContext(ctx, s.cfg.Downloader.FFmpegPath, ffmpegArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("lỗi trích xuất audio bằng FFmpeg: %v (output: %s)", err, string(out))
	}
	return nil
}

func (s *GeminiService) generateSRTForAudio(ctx context.Context, apiKey, audioPath string, expectedDurationMs int64) (string, error) {
	audioBytes, err := os.ReadFile(audioPath)
	if err != nil {
		return "", fmt.Errorf("lỗi đọc file audio: %w", err)
	}
	if len(audioBytes) == 0 {
		return "", fmt.Errorf("file audio rỗng")
	}

	base64Audio := base64.StdEncoding.EncodeToString(audioBytes)

	prompt := `Bạn là chuyên gia nghe lời thoại và dịch phụ đề phim/video mạng xã hội xuất sắc.
Nhiệm vụ của bạn là nghe toàn bộ âm thanh trong tệp này và tạo phụ đề TIẾNG VIỆT chuẩn xác:
1. Tự động nhận diện ngôn ngữ nói của diễn viên/nhân vật (tiếng Trung, tiếng Anh, v.v.).
2. BẮT BUỘC DỊCH TOÀN BỘ 100% SANG TIẾNG VIỆT tự nhiên, đúng ngữ cảnh mạng xã hội Việt Nam. TUYỆT ĐỐI KHÔNG ĐƯỢC GIỮ LẠI BẤT KỲ KÝ TỰ CHỮ HÁN, TIẾNG TRUNG HOẶC TIẾNG NƯỚC NGOÀI GỐC NÀO.
3. CHIA NHỎ TỪNG CÂU PHỤ ĐỀ: Mỗi đoạn phụ đề chỉ từ 2 đến 5 giây (tối đa 7 giây), bám sát chính xác từng câu nói ngắn và nhịp ngắt nghỉ của nhân vật. TUYỆT ĐỐI KHÔNG gộp nhiều câu thành một đoạn dài hay gom cả phút vào một mốc thời gian.
4. Xuất kết quả dưới định dạng chuẩn SubRip (.srt) với số thứ tự tăng dần (1, 2, 3...) và mốc thời gian chính xác dạng 00:00:00,000 --> 00:00:00,000.

YÊU CẦU ĐẶC BIỆT:
- Tất cả nội dung phụ đề phải là TIẾNG VIỆT có dấu hoàn chỉnh.
- Mỗi mục chỉ chứa một câu thoại, tối đa 120 ký tự; tuyệt đối không chèn số thứ tự hay timestamp vào nội dung câu.
- Chỉ trả về nội dung tệp .srt thuần túy, bắt đầu ngay bằng số 1, không có lời chào hỏi, không dùng markdown hay thẻ code.`

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
					{
						"inlineData": map[string]interface{}{
							"mimeType": "audio/mp3",
							"data":     base64Audio,
						},
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.2,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	primaryModel := s.resolveWorkingModel(ctx, apiKey)
	modelsToTry := []string{"gemini-3.5-flash-lite", "gemini-3-flash-preview", primaryModel}

	// Deduplicate
	var uniqueModels []string
	seen := make(map[string]bool)
	for _, m := range modelsToTry {
		if !seen[m] && m != "" {
			seen[m] = true
			uniqueModels = append(uniqueModels, m)
		}
	}

	var lastErr error
	var attemptErrors []string
	for _, model := range uniqueModels {
		maxAttempts := 1
		if model == "gemini-3.6-flash" {
			maxAttempts = 3
		}
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
			req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBytes))
			if err != nil {
				return "", err
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := s.httpClient.Do(req)
			if err != nil {
				lastErr = fmt.Errorf("kết nối Gemini API thất bại: %w", err)
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("Gemini API error (HTTP %d, Model: %s): %s", resp.StatusCode, model, string(bodyBytes))
				attemptErrors = append(attemptErrors, lastErr.Error())
				if resp.StatusCode == 503 && attempt < maxAttempts {
					time.Sleep(time.Duration(attempt*2) * time.Second)
					continue
				}
				break
			}

			var geminiResp struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text    string `json:"text"`
							Thought bool   `json:"thought"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
			}

			if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
				lastErr = err
				continue
			}

			if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
				var contentBuilder strings.Builder
				for _, part := range geminiResp.Candidates[0].Content.Parts {
					if part.Thought || strings.TrimSpace(part.Text) == "" {
						continue
					}
					if contentBuilder.Len() > 0 {
						contentBuilder.WriteString("\n")
					}
					contentBuilder.WriteString(part.Text)
				}
				srtContent := strings.TrimSpace(contentBuilder.String())
				srtContent = strings.TrimPrefix(srtContent, "```srt")
				srtContent = strings.TrimPrefix(srtContent, "```")
				srtContent = strings.TrimSuffix(srtContent, "```")
				srtContent = strings.TrimSpace(srtContent)
				if err := validateSRTChunk(srtContent, expectedDurationMs); err != nil {
					previewRunes := []rune(strings.ReplaceAll(srtContent, "\n", `\n`))
					if len(previewRunes) > 500 {
						previewRunes = previewRunes[:500]
					}
					lastErr = fmt.Errorf("model %s returned invalid subtitles: %w; output preview: %q", model, err, string(previewRunes))
					attemptErrors = append(attemptErrors, lastErr.Error())
					continue
				}
				if srtContent != "" {
					return srtContent, nil
				}
			}
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("%w (các lần thử: %s)", lastErr, strings.Join(attemptErrors, " | "))
	}
	return "", fmt.Errorf("AI không trả về kết quả phụ đề")
}

// GenerateSubtitles extracts audio and uses Gemini AI to generate Vietnamese SRT subtitles with chunking for long videos.
func (s *GeminiService) GenerateSubtitles(ctx context.Context, videoID, videoAbsPath string, startMs, endMs int64) (*domain.VideoSubtitle, error) {
	apiKey := s.getEffectiveAPIKey(ctx)
	if apiKey == "" {
		return nil, fmt.Errorf("Vui lòng cấu hình Google Gemini API Key trong Cài đặt hoặc file .env để sử dụng tính năng tạo phụ đề AI")
	}

	tempDir := filepath.Join(s.cfg.Storage.Root, "temp")
	_ = os.MkdirAll(tempDir, 0755)

	if endMs <= startMs {
		probedDuration := s.probeVideoDuration(ctx, videoAbsPath)
		if probedDuration > startMs {
			endMs = probedDuration
		} else {
			endMs = startMs + 120000 // default 2 minutes
		}
	}

	// Four minutes stays below Gemini Transcribe's free-tier audio-token limit
	// while retaining enough context for natural sentence boundaries.
	const chunkDuration int64 = 4 * 60 * 1000
	type chunkJob struct {
		index int
		start int64
		end   int64
	}
	type chunkResult struct {
		index int
		cues  []domain.SubtitleCue
		err   error
	}

	var jobsList []chunkJob
	for curStart := startMs; curStart < endMs; curStart += chunkDuration {
		curEnd := curStart + chunkDuration
		if curEnd > endMs {
			curEnd = endMs
		}
		jobsList = append(jobsList, chunkJob{index: len(jobsList), start: curStart, end: curEnd})
	}

	workerCount := 2
	if len(jobsList) < workerCount {
		workerCount = len(jobsList)
	}
	jobs := make(chan chunkJob)
	results := make(chan chunkResult, len(jobsList))
	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for job := range jobs {
				if ctx.Err() != nil {
					return
				}
				chunkAudioPath := filepath.Join(tempDir, fmt.Sprintf("chunk_%s_%d_%d.mp3", videoID, job.start, job.end))
				if err := s.extractAudioSegment(ctx, videoAbsPath, chunkAudioPath, job.start, job.end); err != nil {
					_ = os.Remove(chunkAudioPath)
					results <- chunkResult{index: job.index, err: err}
					continue
				}
				srtChunk, err := s.generateSRTForAudio(ctx, apiKey, chunkAudioPath, job.end-job.start)
				_ = os.Remove(chunkAudioPath)
				if err != nil {
					results <- chunkResult{index: job.index, err: err}
					continue
				}
				parsedCues := ParseSRT(srtChunk, job.start)
				cues := make([]domain.SubtitleCue, 0, len(parsedCues))
				for _, cue := range parsedCues {
					if cue.StartMs >= job.end || cue.EndMs <= job.start {
						continue
					}
					cues = append(cues, cue)
				}
				for i := range cues {
					if cues[i].StartMs < job.start {
						cues[i].StartMs = job.start
					}
					if cues[i].EndMs > job.end {
						cues[i].EndMs = job.end
					}
				}
				results <- chunkResult{index: job.index, cues: cues}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for i, job := range jobsList {
			// Stagger starts to avoid request bursts while allowing the Flash-Lite
			// model to process a long video in a reasonable time.
			if i > 0 {
				timer := time.NewTimer(20 * time.Second)
				select {
				case <-timer.C:
				case <-ctx.Done():
					timer.Stop()
					return
				}
			}
			select {
			case jobs <- job:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		workers.Wait()
		close(results)
	}()

	orderedChunks := make([][]domain.SubtitleCue, len(jobsList))
	var chunkErrors []string
	for result := range results {
		if result.err != nil {
			chunkErrors = append(chunkErrors, fmt.Sprintf("%d-%d ms: %v", jobsList[result.index].start, jobsList[result.index].end, result.err))
			continue
		}
		orderedChunks[result.index] = result.cues
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(chunkErrors) > 0 {
		return nil, fmt.Errorf("không thể tạo phụ đề đạt chất lượng cho %d đoạn audio: %s", len(chunkErrors), strings.Join(chunkErrors, "; "))
	}
	var allCues []domain.SubtitleCue
	for _, cues := range orderedChunks {
		allCues = append(allCues, cues...)
	}

	// Models sometimes stop a little early at a chunk boundary. Re-run only
	// those short boundary gaps instead of leaving missing dialogue or
	// regenerating the entire video.
	var boundaries []int64
	for i := 1; i < len(jobsList); i++ {
		boundaries = append(boundaries, jobsList[i].start)
	}
	for _, gap := range findBoundaryGaps(allCues, boundaries, 15000, 60000) {
		repairAudioPath := filepath.Join(tempDir, fmt.Sprintf("repair_%s_%d_%d.mp3", videoID, gap.start, gap.end))
		if err := s.extractAudioSegment(ctx, videoAbsPath, repairAudioPath, gap.start, gap.end); err != nil {
			_ = os.Remove(repairAudioPath)
			continue
		}
		srtChunk, err := s.generateSRTForAudio(ctx, apiKey, repairAudioPath, gap.end-gap.start)
		_ = os.Remove(repairAudioPath)
		if err != nil {
			continue
		}
		allCues = append(allCues, ParseSRT(srtChunk, gap.start)...)
	}

	if len(allCues) == 0 {
		return nil, fmt.Errorf("không phát hiện thấy lời thoại rõ ràng hoặc AI không nhận diện được phụ đề")
	}

	// Post-processing:
	// 1. Detect any Hanzi that was not translated and auto-translate to Vietnamese
	// 2. Split any overly long cues (> 10s) with multiple sentences
	// Gemini normally returns Vietnamese already. Translate rare leftover Hanzi
	// concurrently through the lightweight endpoint instead of serial AI calls.
	translationSlots := make(chan struct{}, 6)
	var translationWorkers sync.WaitGroup
	for i := range allCues {
		if !ContainsHanzi(allCues[i].Text) {
			continue
		}
		translationWorkers.Add(1)
		go func(index int) {
			defer translationWorkers.Done()
			translationSlots <- struct{}{}
			defer func() { <-translationSlots }()
			trText, err := s.quickTranslateGooglePOST(ctx, allCues[index].Text)
			if err == nil && strings.TrimSpace(trText) != "" {
				allCues[index].Text = strings.TrimSpace(trText)
			}
		}(i)
	}
	translationWorkers.Wait()

	// Defensive ordering in case individual requests finish out of order.
	sort.SliceStable(allCues, func(i, j int) bool { return allCues[i].StartMs < allCues[j].StartMs })
	var processedCues []domain.SubtitleCue
	for _, c := range allCues {

		splitCues := SplitLongCueIfNeeded(c)
		processedCues = append(processedCues, splitCues...)
	}
	allCues = processedCues

	// Renumber cues
	for i := range allCues {
		allCues[i].Index = i + 1
	}

	// Rebuild clean SRT text
	var sb strings.Builder
	for _, c := range allCues {
		sb.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n\n",
			c.Index,
			FormatMsToSRTTime(c.StartMs),
			FormatMsToSRTTime(c.EndMs),
			c.Text,
		))
	}
	finalSRT := strings.TrimSpace(sb.String())

	sub := &domain.VideoSubtitle{
		VideoID:        videoID,
		LanguageCode:   "vi",
		SourceLanguage: "Tự động (Tiếng Trung / Gốc)",
		Format:         "srt",
		Content:        finalSRT,
		Cues:           allCues,
	}

	// Save to database
	if s.subRepo != nil {
		_ = s.subRepo.SaveSubtitle(ctx, sub)
	}

	return sub, nil
}

// FormatMsToSRTTime converts millisecond timestamp to SRT formatted time (HH:MM:SS,mmm).
func FormatMsToSRTTime(ms int64) string {
	h := ms / 3600000
	ms = ms % 3600000
	m := ms / 60000
	ms = ms % 60000
	s := ms / 1000
	rem := ms % 1000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, rem)
}

// ParseSRT parses SRT formatted text into structured SubtitleCue items, offset by startMs.
func ParseSRT(srt string, offsetMs int64) []domain.SubtitleCue {
	var cues []domain.SubtitleCue
	normalized := strings.ReplaceAll(srt, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "```srt", "")
	normalized = strings.ReplaceAll(normalized, "```", "")

	// Gemini does not always put a blank line between SRT entries. Parse by
	// timestamp boundaries so a whole chunk cannot become one oversized cue.
	timeRegex := regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2})[,.](\d{3})\s*-->\s*(\d{2}):(\d{2}):(\d{2})[,.](\d{3})`)
	timeLocations := timeRegex.FindAllStringIndex(normalized, -1)
	indexSuffixRegex := regexp.MustCompile(`(?m)\n\s*\d+[.)]?\s*$`)

	for i, location := range timeLocations {
		match := timeRegex.FindStringSubmatch(normalized[location[0]:location[1]])
		if len(match) != 9 {
			continue
		}

		startMs := parseTimeToMs(match[1], match[2], match[3], match[4]) + offsetMs
		endMs := parseTimeToMs(match[5], match[6], match[7], match[8]) + offsetMs

		textEnd := len(normalized)
		if i+1 < len(timeLocations) {
			textEnd = timeLocations[i+1][0]
		}
		text := normalized[location[1]:textEnd]
		if i+1 < len(timeLocations) {
			text = indexSuffixRegex.ReplaceAllString(text, "")
		}
		text = strings.Join(strings.Fields(text), " ")

		if text != "" && endMs > startMs {
			cues = append(cues, domain.SubtitleCue{
				Index:   len(cues) + 1,
				StartMs: startMs,
				EndMs:   endMs,
				Text:    text,
			})
		}
	}

	return cues
}

func validateSRTChunk(srt string, expectedDurationMs int64) error {
	cues := ParseSRT(srt, 0)
	if len(cues) == 0 {
		return fmt.Errorf("không có câu phụ đề hợp lệ")
	}

	embeddedTimestamp := regexp.MustCompile(`\d{2}:\d{2}:\d{2}[,.]\d{3}\s*-->`)
	validCueCount := 0
	for _, cue := range cues {
		if cue.StartMs >= expectedDurationMs || cue.EndMs <= 0 {
			continue
		}
		validCueCount++
		endMs := cue.EndMs
		if endMs > expectedDurationMs {
			endMs = expectedDurationMs
		}
		duration := endMs - cue.StartMs
		if duration <= 0 || duration > 90000 {
			return fmt.Errorf("câu %d có thời lượng %d ms", cue.Index, duration)
		}
		if cue.StartMs < 0 {
			return fmt.Errorf("câu %d có mốc thời gian ngoài đoạn audio", cue.Index)
		}
		if len([]rune(cue.Text)) > 280 {
			return fmt.Errorf("câu %d quá dài (%d ký tự)", cue.Index, len([]rune(cue.Text)))
		}
		if embeddedTimestamp.MatchString(cue.Text) {
			return fmt.Errorf("câu %d chứa timestamp lồng trong nội dung", cue.Index)
		}
	}
	if validCueCount == 0 {
		return fmt.Errorf("không có câu phụ đề trong phạm vi audio")
	}

	return nil
}

type subtitleGap struct {
	start int64
	end   int64
}

func findBoundaryGaps(cues []domain.SubtitleCue, boundaries []int64, minGapMs, maxSideMs int64) []subtitleGap {
	var gaps []subtitleGap
	for _, boundary := range boundaries {
		beforeEnd := int64(-1)
		afterStart := int64(-1)
		for _, cue := range cues {
			if cue.EndMs <= boundary && cue.EndMs > beforeEnd {
				beforeEnd = cue.EndMs
			}
			if cue.StartMs >= boundary && (afterStart == -1 || cue.StartMs < afterStart) {
				afterStart = cue.StartMs
			}
		}
		if beforeEnd < 0 || afterStart < 0 || afterStart-beforeEnd < minGapMs {
			continue
		}
		if boundary-beforeEnd > maxSideMs || afterStart-boundary > maxSideMs {
			continue
		}
		gaps = append(gaps, subtitleGap{start: beforeEnd, end: afterStart})
	}
	return gaps
}

func parseTimeToMs(hStr, mStr, sStr, msStr string) int64 {
	h, _ := strconv.ParseInt(hStr, 10, 64)
	m, _ := strconv.ParseInt(mStr, 10, 64)
	s, _ := strconv.ParseInt(sStr, 10, 64)
	ms, _ := strconv.ParseInt(msStr, 10, 64)
	return h*3600000 + m*60000 + s*1000 + ms
}
