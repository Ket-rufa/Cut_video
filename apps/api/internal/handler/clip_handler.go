package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/service"
)

type ClipHandler struct {
	clipSvc *service.ClipService
	libSvc  *service.LibraryService
}

func NewClipHandler(clipSvc *service.ClipService, libSvc *service.LibraryService) *ClipHandler {
	return &ClipHandler{clipSvc: clipSvc, libSvc: libSvc}
}

func (h *ClipHandler) CreateClip(c *gin.Context) {
	videoID := c.Param("id")
	var req service.CreateClipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code": "INVALID_REQUEST", "message": err.Error(),
		}})
		return
	}
	clip, err := h.clipSvc.CreateClip(c.Request.Context(), videoID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "CREATE_CLIP_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": clip})
}

func (h *ClipHandler) ListClips(c *gin.Context) {
	videoID := c.Param("id")
	clips, err := h.clipSvc.ListByVideoID(c.Request.Context(), videoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "INTERNAL_ERROR", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": clips})
}

func (h *ClipHandler) DeleteClip(c *gin.Context) {
	id := c.Param("id")
	if err := h.clipSvc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "DELETE_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "deleted"}})
}

// Folders
type createFolderReq struct {
	Name  string `json:"name" binding:"required"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

func (h *ClipHandler) ListFolders(c *gin.Context) {
	folders, err := h.libSvc.ListFolders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "INTERNAL_ERROR", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": folders})
}

func (h *ClipHandler) CreateFolder(c *gin.Context) {
	var req createFolderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code": "INVALID_REQUEST", "message": err.Error(),
		}})
		return
	}
	f, err := h.libSvc.CreateFolder(c.Request.Context(), req.Name, req.Icon, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "CREATE_FOLDER_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": f})
}

func (h *ClipHandler) DeleteFolder(c *gin.Context) {
	id := c.Param("id")
	if err := h.libSvc.DeleteFolder(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "DELETE_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "deleted"}})
}

// Tags
type createTagReq struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

func (h *ClipHandler) ListTags(c *gin.Context) {
	tags, err := h.libSvc.ListTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "INTERNAL_ERROR", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tags})
}

func (h *ClipHandler) CreateTag(c *gin.Context) {
	var req createTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code": "INVALID_REQUEST", "message": err.Error(),
		}})
		return
	}
	t, err := h.libSvc.CreateTag(c.Request.Context(), req.Name, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "CREATE_TAG_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": t})
}

func (h *ClipHandler) DeleteTag(c *gin.Context) {
	id := c.Param("id")
	if err := h.libSvc.DeleteTag(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "DELETE_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "deleted"}})
}

// History
func (h *ClipHandler) ListHistory(c *gin.Context) {
	logs, err := h.libSvc.ListHistory(c.Request.Context(), 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "INTERNAL_ERROR", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (h *ClipHandler) CreateHistory(c *gin.Context) {
	var log domain.HistoryLog
	if err := c.ShouldBindJSON(&log); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code": "INVALID_REQUEST", "message": err.Error(),
		}})
		return
	}
	if err := h.libSvc.CreateHistory(c.Request.Context(), &log); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code": "CREATE_HISTORY_FAILED", "message": err.Error(),
		}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": log})
}
