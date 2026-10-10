package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/downloader"
	"social-video-cutter/api/internal/media"
	"social-video-cutter/api/internal/middleware"
	"social-video-cutter/api/internal/service"
)

type VideoHandler struct {
	videoSvc      *service.VideoService
	storageRoot   string
	jobManager    *downloader.JobManager
	registry      *downloader.Registry
	analysisCache *downloader.AnalysisCache
}

func NewVideoHandler(
	videoSvc *service.VideoService,
	storageRoot string,
	jobManager *downloader.JobManager,
	registry *downloader.Registry,
	analysisCache *downloader.AnalysisCache,
) *VideoHandler {
	return &VideoHandler{
		videoSvc:      videoSvc,
		storageRoot:   storageRoot,
		jobManager:    jobManager,
		registry:      registry,
		analysisCache: analysisCache,
	}
}

type analyzeRequest struct {
	URL string `json:"url" binding:"required"`
}

// Analyze handles POST /api/v1/videos/analyze
func (h *VideoHandler) Analyze(c *gin.Context) {
	var req analyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Vui lòng cung cấp URL video hợp lệ",
			},
		})
		return
	}

	rawURL := strings.TrimSpace(req.URL)
	// Auto-extract URL if user pasted a share message with text
	if idx := strings.Index(rawURL, "https://"); idx != -1 {
		rawURL = rawURL[idx:]
		if end := strings.IndexAny(rawURL, " \t\r\n"); end != -1 {
			rawURL = rawURL[:end]
		}
	} else if idx := strings.Index(rawURL, "http://"); idx != -1 {
		rawURL = rawURL[idx:]
		if end := strings.IndexAny(rawURL, " \t\r\n"); end != -1 {
			rawURL = rawURL[:end]
		}
	}

	// Find registered platform adapter
	dl, err := h.registry.FindDownloader(rawURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNSUPPORTED_PLATFORM",
				"message": "Nền tảng chưa được hỗ trợ. Hiện chỉ hỗ trợ YouTube, TikTok, Douyin và Link trực tiếp (MP4/CDN).",
			},
		})
		return
	}

	// Canonicalize and follow redirects if necessary
	canonicalURL, err := dl.CanonicalURL(c.Request.Context(), rawURL)
	if err != nil {
		code := "INVALID_URL"
		if errors.Is(err, downloader.ErrSSRFDetected) {
			code = "SSRF_BLOCKED"
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    code,
				"message": fmt.Sprintf("Không thể chuẩn hóa đường dẫn: %v", err),
			},
		})
		return
	}

	// Analyze video
	result, err := dl.Analyze(c.Request.Context(), canonicalURL)
	if err != nil {
		errStr := err.Error()
		code := "ANALYZE_FAILED"
		msg := "Không thể phân tích video này"
		if parts := strings.SplitN(errStr, ": ", 2); len(parts) == 2 {
			code = parts[0]
			msg = parts[1]
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    code,
				"message": msg,
			},
		})
		return
	}

	// Store in session cache and return analysisId
	analysisID := h.analysisCache.Store(result)
	result.AnalysisID = analysisID

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

type downloadRequest struct {
	AnalysisID string `json:"analysisId"`
	Quality    string `json:"quality"`
	URL        string `json:"url"` // optional fallback if analysisId expired
}

// Download handles POST /api/v1/videos/download
func (h *VideoHandler) Download(c *gin.Context) {
	var req downloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	userID := middleware.GetUserID(c)
	analysisID := strings.TrimSpace(req.AnalysisID)

	// If analysisId is missing but URL is provided, analyze on the fly
	if analysisID == "" && strings.TrimSpace(req.URL) != "" {
		dl, err := h.registry.FindDownloader(req.URL)
		if err == nil {
			canonical, err := dl.CanonicalURL(c.Request.Context(), req.URL)
			if err == nil {
				res, err := dl.Analyze(c.Request.Context(), canonical)
				if err == nil {
					analysisID = h.analysisCache.Store(res)
				}
			}
		}
	}

	if analysisID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "analysisId là bắt buộc. Hãy gọi phân tích video trước.",
			},
		})
		return
	}

	job, err := h.jobManager.CreateDownloadJob(c.Request.Context(), userID, analysisID, req.Quality)
	if err != nil {
		code := "CREATE_JOB_FAILED"
		if errors.Is(err, downloader.ErrVideoTooLong) {
			code = "VIDEO_TOO_LONG"
		} else if errors.Is(err, downloader.ErrFileTooLarge) {
			code = "FILE_TOO_LARGE"
		} else if errors.Is(err, downloader.ErrInsufficientDisk) {
			code = "INSUFFICIENT_DISK_SPACE"
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    code,
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"data": gin.H{
			"jobId":   job.ID,
			"status":  strings.ToUpper(string(job.Status)),
			"videoId": job.VideoID.String,
		},
	})
}

// GetJob handles GET /api/v1/videos/jobs/:jobId
func (h *VideoHandler) GetJob(c *gin.Context) {
	jobID := c.Param("jobId")
	userID := middleware.GetUserID(c)

	job, err := h.jobManager.GetJob(c.Request.Context(), jobID, userID)
	if err != nil {
		if errors.Is(err, downloader.ErrJobNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "JOB_NOT_FOUND",
					"message": "Không tìm thấy tiến trình này",
				},
			})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "Bạn không có quyền xem tiến trình này",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":              job.ID,
			"videoId":         job.VideoID.String,
			"platform":        job.Platform,
			"status":          strings.ToUpper(string(job.Status)),
			"progress":        job.Progress,
			"downloadedBytes": job.DownloadedBytes,
			"totalBytes":      job.TotalBytes,
			"speed":           job.Speed,
			"eta":             job.ETA,
			"error":           job.ErrorMessage,
		},
	})
}

// CancelJob handles POST /api/v1/videos/jobs/:jobId/cancel
func (h *VideoHandler) CancelJob(c *gin.Context) {
	jobID := c.Param("jobId")
	userID := middleware.GetUserID(c)

	if err := h.jobManager.CancelJob(c.Request.Context(), jobID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CANCEL_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status": "cancelled",
		},
	})
}

// GetVideoFile handles GET /api/v1/videos/:id/file (Streaming and Download)
func (h *VideoHandler) GetVideoFile(c *gin.Context) {
	id := c.Param("id")
	video, err := h.videoSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VIDEO_NOT_FOUND",
				"message": "Không tìm thấy video",
			},
		})
		return
	}

	// Ownership check (Rule #3)
	if !middleware.CheckOwnership(c, video.UserID) {
		return
	}

	filePath := video.SourcePath
	cleanPath := filepath.Clean(filePath)

	// If cleanPath does not exist directly, try resolving relative to storageRoot
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		candidate := filepath.Join(h.storageRoot, cleanPath)
		if _, err := os.Stat(candidate); err == nil {
			cleanPath = candidate
		} else {
			cleanRoot := filepath.Clean(h.storageRoot)
			if strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
				rel := strings.TrimPrefix(cleanPath, cleanRoot+string(filepath.Separator))
				candidate2 := filepath.Join(h.storageRoot, rel)
				if _, err := os.Stat(candidate2); err == nil {
					cleanPath = candidate2
				}
			}
		}
	}

	// Security: prevent path traversal (Rule #15)
	absRoot, _ := filepath.Abs(h.storageRoot)
	absPath, _ := filepath.Abs(cleanPath)
	if !strings.HasPrefix(absPath, absRoot+string(filepath.Separator)) && absPath != absRoot {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ACCESS_DENIED",
				"message": "Đường dẫn không hợp lệ",
			},
		})
		return
	}

	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FILE_NOT_FOUND",
				"message": "Tệp tin video không tồn tại trên hệ thống",
			},
		})
		return
	}

	// Dynamic trimming support (e.g. for clip streaming or downloading)
	startMsStr := c.Query("start_ms")
	endMsStr := c.Query("end_ms")
	if startMsStr != "" && endMsStr != "" {
		startMs, errStart := strconv.ParseInt(startMsStr, 10, 64)
		endMs, errEnd := strconv.ParseInt(endMsStr, 10, 64)
		if errStart == nil && errEnd == nil && endMs > startMs {
			speed, _ := strconv.ParseFloat(c.Query("speed"), 64)
			var filterOpts *media.FilterOptions
			filterJSON := c.Query("filter")
			if filterJSON != "" {
				var fo media.FilterOptions
				if err := json.Unmarshal([]byte(filterJSON), &fo); err == nil {
					filterOpts = &fo
				}
			}

			clipsDir := filepath.Join(h.storageRoot, "clips")
			_ = os.MkdirAll(clipsDir, 0755)

			clipFilename := fmt.Sprintf("clip_%s_%d_%d_spd%.2f.mp4", id, startMs, endMs, speed)
			trimmedPath := filepath.Join(clipsDir, clipFilename)

			// Generate trimmed clip with FFmpeg if not already cached
			if _, statErr := os.Stat(trimmedPath); os.IsNotExist(statErr) {
				opts := media.TrimOptions{
					InputPath:  cleanPath,
					OutputPath: trimmedPath,
					StartMs:    startMs,
					EndMs:      endMs,
					Quality:    c.DefaultQuery("quality", "original"),
					Format:     "mp4",
					Speed:      speed,
					FilterOpts: filterOpts,
				}
				args := media.BuildTrimArgs(opts)
				cmd := exec.CommandContext(c.Request.Context(), "ffmpeg", args...)
				if err := cmd.Run(); err != nil {
					_ = os.Remove(trimmedPath)
				}
			}

			if _, statErr := os.Stat(trimmedPath); statErr == nil {
				cleanPath = trimmedPath
			}
		}
	}

	// If download flag is requested, set Content-Disposition
	if c.Query("download") == "1" {
		title := c.Query("title")
		if title == "" {
			title = filepath.Base(cleanPath)
		}
		if !strings.HasSuffix(strings.ToLower(title), ".mp4") {
			title += ".mp4"
		}
		title = strings.ReplaceAll(title, "/", "_")
		title = strings.ReplaceAll(title, "\\", "_")
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, title))
	}

	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("Content-Type", "video/mp4")
	// c.File supports HTTP Range requests (206 Partial Content) natively for video players
	c.File(cleanPath)
}

type resolveRequest struct {
	URL string `json:"url" binding:"required"`
}

// Resolve handles legacy POST /api/v1/videos/resolve
func (h *VideoHandler) Resolve(c *gin.Context) {
	var req resolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code": "INVALID_REQUEST", "message": err.Error(),
		}})
		return
	}

	userID := middleware.GetUserID(c)
	video, err := h.videoSvc.Resolve(c.Request.Context(), req.URL, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code": "RESOLVE_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": video})
}

func (h *VideoHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	video, err := h.videoSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code": "VIDEO_NOT_FOUND", "message": err.Error(),
		}})
		return
	}

	// Ownership check (Rule #3)
	if !middleware.CheckOwnership(c, video.UserID) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": video})
}

func (h *VideoHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	platform := c.Query("platform")
	folderID := c.Query("folder_id")
	search := c.Query("search")

	videos, err := h.videoSvc.ListFiltered(c.Request.Context(), userID, platform, folderID, search, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "INTERNAL_ERROR", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": videos})
}

type updateVideoRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	FolderID    *string `json:"folder_id"`
}

func (h *VideoHandler) Update(c *gin.Context) {
	id := c.Param("id")
	video, err := h.videoSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code": "VIDEO_NOT_FOUND", "message": err.Error(),
		}})
		return
	}

	// Ownership check
	if !middleware.CheckOwnership(c, video.UserID) {
		return
	}

	var req updateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code": "INVALID_REQUEST", "message": err.Error(),
		}})
		return
	}

	if req.Title != nil {
		video.Title = *req.Title
	}
	if req.Description != nil {
		video.Description = *req.Description
	}
	if req.FolderID != nil {
		video.FolderID = *req.FolderID
	}

	if err := h.videoSvc.Update(c.Request.Context(), video); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "UPDATE_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": video})
}

func (h *VideoHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	video, err := h.videoSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code": "VIDEO_NOT_FOUND", "message": err.Error(),
		}})
		return
	}

	// Ownership check
	if !middleware.CheckOwnership(c, video.UserID) {
		return
	}

	if err := h.videoSvc.Delete(c.Request.Context(), id, video.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "DELETE_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "deleted"}})
}

func (h *VideoHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code": "INVALID_FILE", "message": "file is required",
		}})
		return
	}

	userID := middleware.GetUserID(c)
	uploadDir := filepath.Join(h.storageRoot, "source")
	_ = os.MkdirAll(uploadDir, 0755)

	id := uuid.NewString()
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".mp4"
	}
	filename := fmt.Sprintf("%s%s", id, ext)
	dst := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "SAVE_FILE_FAILED", "message": err.Error(),
		}})
		return
	}

	video := &domain.Video{
		ID:           id,
		UserID:       userID,
		SourceURL:    file.Filename,
		Platform:     domain.PlatformGeneric,
		ExternalID:   id,
		Title:        file.Filename,
		Description:  "Video tải lên trực tiếp từ thiết bị",
		ThumbnailURL: "https://images.unsplash.com/photo-1528127269322-539801943592?auto=format&fit=crop&w=640&q=80",
		DurationMs:   120000,
		Width:        1920,
		Height:       1080,
		SourcePath:   dst,
		Status:       domain.VideoStatusResolved,
		FolderID:     "f-all",
		FileSize:     file.Size,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_ = h.videoSvc.Update(c.Request.Context(), video)

	c.JSON(http.StatusOK, gin.H{"data": video})
}
