package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"social-video-cutter/api/internal/service"
)

type JobHandler struct {
	exportSvc   *service.ExportService
	storageRoot string
}

func NewJobHandler(exportSvc *service.ExportService, storageRoot string) *JobHandler {
	return &JobHandler{exportSvc: exportSvc, storageRoot: storageRoot}
}

type trimRequest struct {
	StartMs int64  `json:"start_ms" binding:"required"`
	EndMs   int64  `json:"end_ms"   binding:"required"`
	Format  string `json:"format"`
	Quality string `json:"quality"`
}

func (h *JobHandler) CreateTrimJob(c *gin.Context) {
	videoID := c.Param("id")
	var req trimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code": "INVALID_REQUEST", "message": err.Error(),
		}})
		return
	}
	job, err := h.exportSvc.CreateTrimJob(c.Request.Context(), service.TrimRequest{
		VideoID: videoID,
		StartMs: req.StartMs,
		EndMs:   req.EndMs,
		Format:  req.Format,
		Quality: req.Quality,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code": "CREATE_JOB_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": job})
}

func (h *JobHandler) GetJob(c *gin.Context) {
	jobID := c.Param("id")
	job, err := h.exportSvc.GetJob(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code": "JOB_NOT_FOUND", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": job})
}

func (h *JobHandler) GetJobProgress(c *gin.Context) {
	jobID := c.Param("id")
	job, err := h.exportSvc.GetJob(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code": "JOB_NOT_FOUND", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"job_id":   job.ID,
		"status":   job.Status,
		"progress": job.Progress,
	}})
}

func (h *JobHandler) ListJobs(c *gin.Context) {
	jobs, err := h.exportSvc.ListJobs(c.Request.Context(), 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "INTERNAL_ERROR", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": jobs})
}

func (h *JobHandler) GetExport(c *gin.Context) {
	id := c.Param("id")

	// Try by export ID first
	export, err := h.exportSvc.GetExport(c.Request.Context(), id)
	if err != nil {
		// Fall back: treat id as job ID
		export2, err2 := h.exportSvc.GetExportByJobID(c.Request.Context(), id)
		if err2 != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
				"code": "EXPORT_NOT_FOUND", "message": "export not found",
			}})
			return
		}
		export = export2
	}
	c.JSON(http.StatusOK, gin.H{"data": export})
}

func (h *JobHandler) DownloadExport(c *gin.Context) {
	id := c.Param("id")
	export, err := h.exportSvc.GetExport(c.Request.Context(), id)
	if err != nil {
		// Fall back: treat id as job ID
		export2, err2 := h.exportSvc.GetExportByJobID(c.Request.Context(), id)
		if err2 != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
				"code": "EXPORT_NOT_FOUND", "message": "export not found",
			}})
			return
		}
		export = export2
	}

	// Security: prevent path traversal
	cleanPath := filepath.Clean(export.OutputPath)
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
	// Ensure file is within storage root (prevent path traversal)
	absRoot, _ := filepath.Abs(h.storageRoot)
	absPath, _ := filepath.Abs(cleanPath)
	if !strings.HasPrefix(absPath, absRoot+string(filepath.Separator)) && absPath != absRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": gin.H{
			"code": "ACCESS_DENIED", "message": "invalid path",
		}})
		return
	}

	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code": "FILE_NOT_FOUND", "message": "export file not found on disk",
		}})
		return
	}
	filename := filepath.Base(cleanPath)
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Header("Content-Type", "video/mp4")
	c.File(cleanPath)
}
