package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"social-video-cutter/api/internal/ai"
	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/repository"
	"social-video-cutter/api/internal/service"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	geminiSvc   *ai.GeminiService
	dubbingSvc  *ai.DubbingService
	videoSvc    *service.VideoService
	subRepo     repository.SubtitleRepository
	storageRoot string
}

func NewAIHandler(
	geminiSvc *ai.GeminiService,
	dubbingSvc *ai.DubbingService,
	videoSvc *service.VideoService,
	subRepo repository.SubtitleRepository,
	storageRoot string,
) *AIHandler {
	return &AIHandler{
		geminiSvc:   geminiSvc,
		dubbingSvc:  dubbingSvc,
		videoSvc:    videoSvc,
		subRepo:     subRepo,
		storageRoot: storageRoot,
	}
}

// TranslateMeta handles POST /api/v1/videos/:id/translate-meta
func (h *AIHandler) TranslateMeta(c *gin.Context) {
	id := c.Param("id")
	video, err := h.videoSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "VIDEO_NOT_FOUND", "message": "Không tìm thấy video"},
		})
		return
	}

	var tags []string
	if video.MetadataJSON != "" {
		var meta struct {
			Tags []string `json:"tags"`
		}
		if err := json.Unmarshal([]byte(video.MetadataJSON), &meta); err == nil && len(meta.Tags) > 0 {
			tags = meta.Tags
		}
	}
	if len(tags) == 0 {
		// Extract from hashtags in title / description
		words := strings.Fields(video.Title + " " + video.Description)
		for _, w := range words {
			if strings.HasPrefix(w, "#") {
				tags = append(tags, strings.TrimPrefix(w, "#"))
			}
		}
	}

	result, err := h.geminiSvc.TranslateMetadata(c.Request.Context(), video.Title, video.Description, tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "TRANSLATION_FAILED", "message": err.Error()},
		})
		return
	}

	// Auto-apply if requested
	if c.Query("apply") == "1" || c.Query("apply") == "true" {
		video.Title = result.TranslatedTitle
		video.Description = result.TranslatedDescription
		_ = h.videoSvc.Update(c.Request.Context(), video)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

type generateSubtitlesRequest struct {
	StartMs int64 `json:"start_ms"`
	EndMs   int64 `json:"end_ms"`
}

// GenerateSubtitles handles POST /api/v1/videos/:id/subtitles/generate
func (h *AIHandler) GenerateSubtitles(c *gin.Context) {
	id := c.Param("id")
	video, err := h.videoSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "VIDEO_NOT_FOUND", "message": "Không tìm thấy video"},
		})
		return
	}

	var req generateSubtitlesRequest
	_ = c.ShouldBindJSON(&req)

	// Resolve video cleanPath
	cleanPath := filepath.Clean(video.SourcePath)
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		candidate := filepath.Join(h.storageRoot, cleanPath)
		if _, err := os.Stat(candidate); err == nil {
			cleanPath = candidate
		} else {
			cleanRoot := filepath.Clean(h.storageRoot)
			if strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
				rel := strings.TrimPrefix(cleanPath, cleanRoot+string(filepath.Separator))
				cand2 := filepath.Join(h.storageRoot, rel)
				if _, err := os.Stat(cand2); err == nil {
					cleanPath = cand2
				}
			}
		}
	}

	absPath, err := filepath.Abs(cleanPath)
	if err != nil || !filepath.IsAbs(absPath) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_PATH", "message": "Đường dẫn video không hợp lệ"},
		})
		return
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "FILE_NOT_FOUND", "message": "Không tìm thấy tệp video trên đĩa để tách âm thanh"},
		})
		return
	}

	startMs := req.StartMs
	endMs := req.EndMs
	if endMs <= startMs && video.DurationMs > 0 {
		endMs = video.DurationMs
	}

	sub, err := h.geminiSvc.GenerateSubtitles(c.Request.Context(), video.ID, absPath, startMs, endMs)
	if err != nil {
		fmt.Printf("[SUBTITLE ERROR] %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "SUBTITLE_GEN_FAILED", "message": err.Error()},
		})
		return
	}
	h.invalidateDubbing(id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sub,
	})
}

// GetSubtitles handles GET /api/v1/videos/:id/subtitles
func (h *AIHandler) GetSubtitles(c *gin.Context) {
	id := c.Param("id")
	sub, err := h.subRepo.GetByVideoID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "DB_ERROR", "message": err.Error()},
		})
		return
	}

	if sub == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sub,
	})
}

// DownloadSubtitles handles GET /api/v1/videos/:id/subtitles/download
func (h *AIHandler) DownloadSubtitles(c *gin.Context) {
	id := c.Param("id")
	sub, err := h.subRepo.GetByVideoID(c.Request.Context(), id)
	if err != nil || sub == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "SUBTITLES_NOT_FOUND", "message": "Chưa có phụ đề cho video này"},
		})
		return
	}

	video, _ := h.videoSvc.GetByID(c.Request.Context(), id)
	filename := fmt.Sprintf("subtitles_%s.srt", id)
	if video != nil && video.Title != "" {
		cleanTitle := strings.ReplaceAll(video.Title, " ", "_")
		if len(cleanTitle) > 50 {
			cleanTitle = cleanTitle[:50]
		}
		filename = fmt.Sprintf("%s_subtitles.srt", cleanTitle)
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, sub.Content)
}

type updateSubtitlesRequest struct {
	Cues []domain.SubtitleCue `json:"cues"`
}

// UpdateSubtitles handles PUT /api/v1/videos/:id/subtitles
func (h *AIHandler) UpdateSubtitles(c *gin.Context) {
	id := c.Param("id")
	sub, err := h.subRepo.GetByVideoID(c.Request.Context(), id)
	if err != nil || sub == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "SUBTITLES_NOT_FOUND", "message": "Chưa có phụ đề cho video này"},
		})
		return
	}

	var req updateSubtitlesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_INPUT", "message": "Dữ liệu phụ đề không hợp lệ"},
		})
		return
	}

	var normalizedCues []domain.SubtitleCue
	for _, cue := range req.Cues {
		normalizedCues = append(normalizedCues, ai.SplitLongCueIfNeeded(cue)...)
	}
	for i := range normalizedCues {
		normalizedCues[i].Index = i + 1
	}
	sub.Cues = normalizedCues

	var sb strings.Builder
	for _, cue := range sub.Cues {
		sb.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n\n",
			cue.Index,
			ai.FormatMsToSRTTime(cue.StartMs),
			ai.FormatMsToSRTTime(cue.EndMs),
			cue.Text,
		))
	}
	sub.Content = strings.TrimSpace(sb.String())

	if err := h.subRepo.SaveSubtitle(c.Request.Context(), sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "SAVE_ERROR", "message": err.Error()},
		})
		return
	}
	h.invalidateDubbing(id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sub,
	})
}

func (h *AIHandler) invalidateDubbing(videoID string) {
	dubbingDir := filepath.Join(h.storageRoot, "dubbing")
	_ = os.Remove(filepath.Join(dubbingDir, fmt.Sprintf("dubbed_audio_%s.mp3", videoID)))
	_ = os.Remove(filepath.Join(dubbingDir, fmt.Sprintf("dubbed_video_%s.mp4", videoID)))
}

type translateCueRequest struct {
	Index   int    `json:"index"`
	Text    string `json:"text" binding:"required"`
	StartMs int64  `json:"start_ms"`
	EndMs   int64  `json:"end_ms"`
}

// TranslateCue handles POST /api/v1/videos/:id/subtitles/translate-cue
func (h *AIHandler) TranslateCue(c *gin.Context) {
	var req translateCueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_INPUT", "message": "Nội dung câu cần dịch không được để trống"},
		})
		return
	}

	translated, err := h.geminiSvc.TranslateTextToVietnamese(c.Request.Context(), req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "TRANSLATE_FAILED", "message": err.Error()},
		})
		return
	}

	var splitted []domain.SubtitleCue
	if req.EndMs > req.StartMs && (req.EndMs-req.StartMs) > 10000 {
		dummyCue := domain.SubtitleCue{
			Index:   req.Index,
			StartMs: req.StartMs,
			EndMs:   req.EndMs,
			Text:    translated,
		}
		splitted = ai.SplitLongCueIfNeeded(dummyCue)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"original_text":   req.Text,
			"translated_text": translated,
			"splitted_cues":   splitted,
		},
	})
}

// GetAIConfig handles GET /api/v1/settings/ai
func (h *AIHandler) GetAIConfig(c *gin.Context) {
	key := h.geminiSvc.GetCurrentAPIKey(c.Request.Context())
	hasKey := key != ""
	maskedKey := ""
	if hasKey {
		if len(key) > 8 {
			maskedKey = key[:4] + "••••••••" + key[len(key)-4:]
		} else {
			maskedKey = "••••••••"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"has_key":    hasKey,
			"masked_key": maskedKey,
			"model":      "gemini-3.5-flash-lite",
		},
	})
}

type updateAIConfigRequest struct {
	APIKey string `json:"api_key"`
}

// UpdateAIConfig handles POST /api/v1/settings/ai
func (h *AIHandler) UpdateAIConfig(c *gin.Context) {
	var req updateAIConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_REQUEST", "message": err.Error()},
		})
		return
	}

	if err := h.geminiSvc.SaveAPIKey(c.Request.Context(), req.APIKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "SAVE_FAILED", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Đã lưu khóa API Gemini thành công!",
	})
}

// TestAIConnection handles POST /api/v1/settings/ai/test
func (h *AIHandler) TestAIConnection(c *gin.Context) {
	var req updateAIConfigRequest
	_ = c.ShouldBindJSON(&req)

	ok, msg, err := h.geminiSvc.TestConnection(c.Request.Context(), req.APIKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "CONNECTION_FAILED", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": ok,
		"message": msg,
	})
}

type generateDubbingRequest struct {
	StartMs  int64   `json:"start_ms"`
	EndMs    int64   `json:"end_ms"`
	BGVolume float64 `json:"bg_volume"`
	Voice    string  `json:"voice"`
}

// GenerateDubbing handles POST /api/v1/videos/:id/dubbing/generate
func (h *AIHandler) GenerateDubbing(c *gin.Context) {
	id := c.Param("id")
	video, err := h.videoSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "VIDEO_NOT_FOUND", "message": "Không tìm thấy video"},
		})
		return
	}

	sub, err := h.subRepo.GetByVideoID(c.Request.Context(), id)
	if err != nil || sub == nil || len(sub.Cues) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "NO_SUBTITLES", "message": "Vui lòng tạo phụ đề tiếng Việt trước khi thực hiện lồng tiếng AI"},
		})
		return
	}

	var req generateDubbingRequest
	_ = c.ShouldBindJSON(&req)

	startMs := req.StartMs
	endMs := req.EndMs
	if endMs <= startMs {
		endMs = video.DurationMs
	}
	if req.BGVolume <= 0 {
		req.BGVolume = 0.15
	}
	if req.Voice == "" {
		req.Voice = "male_deep"
	}

	dubbing, err := h.dubbingSvc.GenerateDubbingTrack(c.Request.Context(), video, sub.Cues, startMs, endMs, req.BGVolume, req.Voice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "DUBBING_FAILED", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dubbing,
	})
}

// GetDubbing handles GET /api/v1/videos/:id/dubbing
func (h *AIHandler) GetDubbing(c *gin.Context) {
	id := c.Param("id")
	dubbingDir := filepath.Join(h.storageRoot, "dubbing")
	audioFile := filepath.Join(dubbingDir, fmt.Sprintf("dubbed_audio_%s.mp3", id))
	videoFile := filepath.Join(dubbingDir, fmt.Sprintf("dubbed_video_%s.mp4", id))

	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "NOT_FOUND", "message": "Chưa có bản lồng tiếng cho video này"},
		})
		return
	}

	var videoURL string
	if _, err := os.Stat(videoFile); err == nil {
		videoURL = fmt.Sprintf("/api/v1/videos/%s/dubbing/video", id)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": domain.VideoDubbing{
			VideoID:   id,
			AudioURL:  fmt.Sprintf("/api/v1/videos/%s/dubbing/audio", id),
			VideoURL:  videoURL,
			Status:    "ready",
			CreatedAt: time.Now(),
		},
	})
}

// StreamDubbedAudio handles GET /api/v1/videos/:id/dubbing/audio
func (h *AIHandler) StreamDubbedAudio(c *gin.Context) {
	id := c.Param("id")
	audioFile := filepath.Join(h.storageRoot, "dubbing", fmt.Sprintf("dubbed_audio_%s.mp3", id))
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy file audio lồng tiếng"})
		return
	}
	c.Header("Content-Type", "audio/mpeg")
	c.File(audioFile)
}

// StreamDubbedVideo handles GET /api/v1/videos/:id/dubbing/video
func (h *AIHandler) StreamDubbedVideo(c *gin.Context) {
	id := c.Param("id")
	videoFile := filepath.Join(h.storageRoot, "dubbing", fmt.Sprintf("dubbed_video_%s.mp4", id))
	if _, err := os.Stat(videoFile); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy file video lồng tiếng"})
		return
	}
	c.Header("Content-Type", "video/mp4")
	c.File(videoFile)
}

// StreamCueAudio handles GET /api/v1/videos/:id/subtitles/cues/:index/audio?voice=male_deep
func (h *AIHandler) StreamCueAudio(c *gin.Context) {
	id := c.Param("id")
	indexStr := c.Param("index")
	var idx int
	fmt.Sscanf(indexStr, "%d", &idx)

	voice := c.DefaultQuery("voice", "male_deep")

	sub, err := h.subRepo.GetByVideoID(c.Request.Context(), id)
	if err != nil || sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy phụ đề"})
		return
	}

	var targetCue *domain.SubtitleCue
	for _, cue := range sub.Cues {
		if cue.Index == idx {
			targetCue = &cue
			break
		}
	}

	if targetCue == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy câu phụ đề"})
		return
	}

	tempDir := filepath.Join(h.storageRoot, "temp")
	_ = os.MkdirAll(tempDir, 0755)
	cueAudioPath := filepath.Join(tempDir, fmt.Sprintf("preview_%s_%d_%s.mp3", id, idx, voice))

	targetDur := float64(targetCue.EndMs-targetCue.StartMs) / 1000.0
	if err := h.dubbingSvc.SynthesizeSyncedCueTTS(c.Request.Context(), targetCue.Text, targetDur, cueAudioPath, voice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "audio/mpeg")
	c.File(cueAudioPath)
}

// GetDubbingVoices handles GET /api/v1/dubbing/voices
func (h *AIHandler) GetDubbingVoices(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ai.AvailableVoices,
	})
}

// PreviewVoice handles GET /api/v1/dubbing/preview?voice=male_deep
func (h *AIHandler) PreviewVoice(c *gin.Context) {
	voice := c.DefaultQuery("voice", "male_deep")
	audioBytes, err := h.dubbingSvc.SynthesizeSampleVoice(c.Request.Context(), voice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "audio/mpeg")
	c.Data(http.StatusOK, "audio/mpeg", audioBytes)
}
