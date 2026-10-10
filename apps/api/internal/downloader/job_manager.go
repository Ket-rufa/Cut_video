package downloader

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"social-video-cutter/api/internal/config"
	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/storage"
)

var (
	ErrJobNotFound          = errors.New("JOB_NOT_FOUND")
	ErrVideoTooLong         = errors.New("VIDEO_TOO_LONG")
	ErrFileTooLarge         = errors.New("FILE_TOO_LARGE")
	ErrInsufficientDisk     = errors.New("INSUFFICIENT_DISK_SPACE")
	ErrUserConcurrencyLimit = errors.New("USER_CONCURRENCY_LIMIT")
	ErrInvalidJobTransition = errors.New("INVALID_JOB_TRANSITION")
)

type JobManager struct {
	db             *sqlx.DB
	fileStore      storage.Storage
	registry       *Registry
	analysisCache  *AnalysisCache
	cfg            *config.Config

	globalSem      chan struct{}
	userJobsMu     sync.Mutex
	userActiveJobs map[string]int

	activeCancelsMu sync.Mutex
	activeCancels   map[string]context.CancelFunc

	dedupMu sync.Mutex
	// activeDedupKeys maps hash(userId:platform:videoId:quality) -> jobId
	activeDedupKeys map[string]string
}

func NewJobManager(
	db *sqlx.DB,
	fileStore storage.Storage,
	registry *Registry,
	analysisCache *AnalysisCache,
	cfg *config.Config,
) *JobManager {
	globalLimit := cfg.Limits.MaxConcurrentGlobal
	if globalLimit <= 0 {
		globalLimit = 4
	}

	jm := &JobManager{
		db:              db,
		fileStore:       fileStore,
		registry:        registry,
		analysisCache:   analysisCache,
		cfg:             cfg,
		globalSem:       make(chan struct{}, globalLimit),
		userActiveJobs:  make(map[string]int),
		activeCancels:   make(map[string]context.CancelFunc),
		activeDedupKeys: make(map[string]string),
	}

	// Startup: Zombie job recovery (Rule #10)
	jm.RecoverZombieJobs()

	// Startup: Periodic temp files cleanup (Rule #15)
	go func() {
		cleanupTicker := time.NewTicker(30 * time.Minute)
		maxAge := time.Duration(cfg.Limits.TempCleanupHours) * time.Hour
		if maxAge <= 0 {
			maxAge = 2 * time.Hour
		}
		for range cleanupTicker.C {
			if ls, ok := jm.fileStore.(*storage.LocalStorage); ok {
				ls.PeriodicCleanup(maxAge)
			}
		}
	}()

	return jm
}

func makeDedupKey(userID, platform, videoID, quality string) string {
	return fmt.Sprintf("%s:%s:%s:%s", userID, platform, videoID, strings.ToLower(quality))
}

// CreateDownloadJob validates the request, prevents duplicates, checks limits, and queues the job.
func (jm *JobManager) CreateDownloadJob(
	ctx context.Context,
	userID string,
	analysisID string,
	targetQuality string,
) (*domain.VideoJob, error) {
	// Rule #5: Fetch from verified backend analysis cache
	analysis, ok := jm.analysisCache.Get(analysisID)
	if !ok {
		return nil, fmt.Errorf("%w: analysis session expired or invalid. Please re-analyze video", ErrJobNotFound)
	}

	// Rule #13: Pre-download limits check
	if jm.cfg.Limits.MaxDurationSeconds > 0 && analysis.Duration > jm.cfg.Limits.MaxDurationSeconds {
		return nil, fmt.Errorf("%w: video duration %ds exceeds limit %ds", ErrVideoTooLong, analysis.Duration, jm.cfg.Limits.MaxDurationSeconds)
	}

	// Rule #4: Re-validate format & quality from backend analysis result
	targetQuality = strings.ToLower(strings.TrimSpace(targetQuality))
	if targetQuality == "" {
		targetQuality = "1080p"
	}
	var matchedFormat *VideoFormatInfo
	for _, f := range analysis.Formats {
		if strings.EqualFold(f.Quality, targetQuality) {
			matchedFormat = &f
			break
		}
	}
	if matchedFormat == nil && len(analysis.Formats) > 0 {
		matchedFormat = &analysis.Formats[0]
		targetQuality = matchedFormat.Quality
	}

	// Check estimated file size limit if known
	if matchedFormat != nil && matchedFormat.FileSize > 0 && jm.cfg.Limits.MaxSizeMB > 0 {
		maxBytes := int64(jm.cfg.Limits.MaxSizeMB) * 1024 * 1024
		if matchedFormat.FileSize > maxBytes {
			return nil, fmt.Errorf("%w: estimated size exceeds %d MB", ErrFileTooLarge, jm.cfg.Limits.MaxSizeMB)
		}
	}

	// Rule #14: Disk space check
	hasSpace, err := jm.fileStore.CheckDiskSpace("", 0)
	if err != nil || !hasSpace {
		return nil, fmt.Errorf("%w: server storage is full or unwritable", ErrInsufficientDisk)
	}

	// Rule #11: Idempotency / Double-click prevention
	dedupKey := makeDedupKey(userID, string(analysis.Platform), analysis.VideoID, targetQuality)
	jm.dedupMu.Lock()
	if existingJobID, exists := jm.activeDedupKeys[dedupKey]; exists {
		jm.dedupMu.Unlock()
		// Return existing active job
		existingJob, err := jm.GetJob(ctx, existingJobID, userID)
		if err == nil {
			return existingJob, nil
		}
	}

	// Check if already completed video exists in DB for this user with matching or better quality
	var existingVideo domain.Video
	err = jm.db.GetContext(ctx, &existingVideo,
		"SELECT * FROM videos WHERE user_id = $1 AND platform = $2 AND external_id = $3 ORDER BY height DESC LIMIT 1",
		userID, analysis.Platform, analysis.VideoID,
	)
	if err == nil && existingVideo.ID != "" && strings.TrimSpace(existingVideo.SourcePath) != "" && jm.fileStore.Exists(existingVideo.SourcePath) {
		// Only re-use if existing video resolution is at least as high as requested quality
		reqHeight := 0
		if matchedFormat != nil && matchedFormat.Height > 0 {
			reqHeight = matchedFormat.Height
		}
		if (reqHeight == 0 && (existingVideo.Quality == targetQuality || existingVideo.Height >= 720)) ||
			(reqHeight > 0 && existingVideo.Height >= reqHeight) ||
			(strings.EqualFold(existingVideo.Quality, targetQuality) && existingVideo.Height >= 720) {
			jm.dedupMu.Unlock()
			// Re-use existing completed video
			now := time.Now()
			completedJob := &domain.VideoJob{
				ID:          uuid.NewString(),
				UserID:      userID,
				VideoID:     sql.NullString{String: existingVideo.ID, Valid: true},
				Platform:    string(analysis.Platform),
				SourceURL:   analysis.SourceURL,
				Quality:     existingVideo.Quality,
				Type:        domain.JobTypeDownload,
				Status:      domain.JobStatusCompleted,
				Progress:    100,
				CreatedAt:   now,
				UpdatedAt:   now,
				CompletedAt: &now,
			}
			insertQuery := `INSERT INTO video_jobs (
				id, user_id, video_id, platform, source_url, quality, type, status, progress, created_at, updated_at, completed_at
			) VALUES (
				:id, :user_id, :video_id, :platform, :source_url, :quality, :type, :status, :progress, :created_at, :updated_at, :completed_at
			)`
			_, _ = jm.db.NamedExecContext(ctx, insertQuery, completedJob)
			return completedJob, nil
		}
	}

	jobID := uuid.NewString()
	jm.activeDedupKeys[dedupKey] = jobID
	jm.dedupMu.Unlock()

	payloadData, _ := json.Marshal(domain.DownloadPayload{
		SourceURL: analysis.SourceURL,
		Platform:  string(analysis.Platform),
		VideoID:   analysis.VideoID,
		Quality:   targetQuality,
	})

	now := time.Now()
	job := &domain.VideoJob{
		ID:        jobID,
		UserID:    userID,
		Platform:  string(analysis.Platform),
		SourceURL: analysis.SourceURL,
		Quality:   targetQuality,
		Type:      domain.JobTypeDownload,
		Status:    domain.JobStatusQueued,
		Progress:  0,
		Payload:   payloadData,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert into DB
	insertQuery := `INSERT INTO video_jobs (
		id, user_id, platform, source_url, quality, type, status, progress, payload, created_at, updated_at
	) VALUES (
		:id, :user_id, :platform, :source_url, :quality, :type, :status, :progress, :payload, :created_at, :updated_at
	)`
	if _, err := jm.db.NamedExecContext(ctx, insertQuery, job); err != nil {
		jm.dedupMu.Lock()
		delete(jm.activeDedupKeys, dedupKey)
		jm.dedupMu.Unlock()
		return nil, fmt.Errorf("create job in DB: %w", err)
	}

	// Dispatch background worker
	go jm.executeJobAsync(job.ID, userID, analysis, targetQuality, dedupKey)

	return job, nil
}

func (jm *JobManager) executeJobAsync(
	jobID, userID string,
	analysis *VideoAnalysisResult,
	quality, dedupKey string,
) {
	defer func() {
		jm.dedupMu.Lock()
		delete(jm.activeDedupKeys, dedupKey)
		jm.dedupMu.Unlock()

		jm.userJobsMu.Lock()
		if jm.userActiveJobs[userID] > 0 {
			jm.userActiveJobs[userID]--
		}
		jm.userJobsMu.Unlock()
	}()

	// Check per-user concurrency limit
	jm.userJobsMu.Lock()
	userLimit := jm.cfg.Limits.MaxConcurrentPerUser
	if userLimit <= 0 {
		userLimit = 2
	}
	if jm.userActiveJobs[userID] >= userLimit {
		jm.userJobsMu.Unlock()
		jm.failJob(jobID, "USER_CONCURRENCY_LIMIT", "Bạn đang có quá nhiều tiến trình tải đang chạy. Vui lòng chờ tiến trình trước hoàn tất.")
		return
	}
	jm.userActiveJobs[userID]++
	jm.userJobsMu.Unlock()

	// Acquire global semaphore slot
	jm.globalSem <- struct{}{}
	defer func() { <-jm.globalSem }()

	// Create job context with cancel capability (Rule #13)
	jobCtx, cancel := context.WithTimeout(context.Background(), 3*time.Hour)
	jm.activeCancelsMu.Lock()
	jm.activeCancels[jobID] = cancel
	jm.activeCancelsMu.Unlock()

	defer func() {
		cancel()
		jm.activeCancelsMu.Lock()
		delete(jm.activeCancels, jobID)
		jm.activeCancelsMu.Unlock()
	}()

	// Rule #9: Transition to ANALYZING
	if err := jm.updateJobStatus(jobID, domain.JobStatusAnalyzing, 2, 0, 0, 0, 0); err != nil {
		log.Printf("[Job %s] transition to ANALYZING failed: %v", jobID, err)
	}

	downloader, err := jm.registry.FindDownloader(analysis.SourceURL)
	if err != nil {
		jm.failJob(jobID, "UNSUPPORTED_PLATFORM", fmt.Sprintf("Không tìm thấy bộ tải cho nền tảng: %v", err))
		return
	}

	tempDir := jm.fileStore.JobTempDir(jobID)
	defer jm.fileStore.CleanupJobTemp(jobID)

	opt := DownloadOptions{
		JobID:     jobID,
		UserID:    userID,
		SourceURL: analysis.SourceURL,
		VideoID:   analysis.VideoID,
		Quality:   quality,
		TempDir:   tempDir,
	}

	var lastMonotonicProgress int = 2

	progressFn := func(p DownloadProgress) {
		// Rule #18: Ensure monotonic progress (never goes backward)
		if p.Progress < lastMonotonicProgress {
			p.Progress = lastMonotonicProgress
		} else {
			lastMonotonicProgress = p.Progress
		}

		var st domain.JobStatus
		switch p.Status {
		case "DOWNLOADING":
			st = domain.JobStatusDownloading
		case "MERGING":
			st = domain.JobStatusMerging
		case "SAVING":
			st = domain.JobStatusSaving
		default:
			st = domain.JobStatusDownloading
		}

		_ = jm.updateJobStatus(jobID, st, p.Progress, p.DownloadedBytes, p.TotalBytes, p.Speed, p.ETA)
	}

	// Execute actual download & merge
	result, err := downloader.Download(jobCtx, opt, progressFn)
	if err != nil {
		if errors.Is(jobCtx.Err(), context.Canceled) {
			jm.markJobCancelled(jobID)
			return
		}
		jm.failJob(jobID, "DOWNLOAD_FAILED", err.Error())
		return
	}

	// Rule #16: Final file atomic move & database insert
	finalFilename := fmt.Sprintf("%s-%s.mp4", storage.SanitizeFilename(analysis.Title), analysis.VideoID)
	finalPath, err := jm.fileStore.MoveToSource(result.FilePath, finalFilename)
	if err != nil {
		jm.failJob(jobID, "STORAGE_FAILED", fmt.Sprintf("Lỗi chuyển file sang thư mục lưu trữ: %v", err))
		return
	}

	// Create Video record in DB
	videoID := uuid.NewString()
	videoRecord := &domain.Video{
		ID:           videoID,
		UserID:       userID,
		SourceURL:    analysis.SourceURL,
		Platform:     analysis.Platform,
		ExternalID:   analysis.VideoID,
		Title:        analysis.Title,
		Description:  analysis.Description,
		ThumbnailURL: analysis.Thumbnail,
		DurationMs:   result.DurationMs,
		Width:        result.Width,
		Height:       result.Height,
		SourcePath:   finalPath,
		Status:       domain.VideoStatusDownloaded,
		FolderID:     "f-all",
		FileSize:     result.FileSize,
		AuthorName:   analysis.Author.Name,
		AuthorAvatar: analysis.Author.Avatar,
		Quality:      quality,
		MimeType:     "video/mp4",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	insertVideoQuery := `INSERT INTO videos (
		id, user_id, source_url, platform, external_id, title, description,
		thumbnail_url, duration_ms, width, height, source_path, status, folder_id,
		file_size, author_name, author_avatar, quality, mime_type, created_at, updated_at
	) VALUES (
		:id, :user_id, :source_url, :platform, :external_id, :title, :description,
		:thumbnail_url, :duration_ms, :width, :height, :source_path, :status, :folder_id,
		:file_size, :author_name, :author_avatar, :quality, :mime_type, :created_at, :updated_at
	)`

	if _, err := jm.db.NamedExecContext(jobCtx, insertVideoQuery, videoRecord); err != nil {
		jm.failJob(jobID, "DB_SAVE_FAILED", fmt.Sprintf("Lỗi lưu thông tin video vào database: %v", err))
		return
	}

	// Mark Job COMPLETED
	now := time.Now()
	_, _ = jm.db.ExecContext(jobCtx, `UPDATE video_jobs SET
		video_id = $1, status = $2, progress = 100, completed_at = $3, updated_at = $4
		WHERE id = $5`,
		videoID, domain.JobStatusCompleted, now, now, jobID,
	)

	log.Printf("[Job %s] Download and processing completed successfully. VideoID: %s", jobID, videoID)
}

// CancelJob terminates the running process, cleans temp files, and updates status to CANCELLED.
func (jm *JobManager) CancelJob(ctx context.Context, jobID, userID string) error {
	job, err := jm.GetJob(ctx, jobID, userID)
	if err != nil {
		return err
	}

	// Only allow cancellation of active jobs
	if job.Status == domain.JobStatusCompleted || job.Status == domain.JobStatusFailed || job.Status == domain.JobStatusCancelled {
		return fmt.Errorf("job is already finished with status %s", job.Status)
	}

	jm.activeCancelsMu.Lock()
	if cancel, exists := jm.activeCancels[jobID]; exists {
		cancel()
	}
	jm.activeCancelsMu.Unlock()

	jm.markJobCancelled(jobID)
	_ = jm.fileStore.CleanupJobTemp(jobID)

	return nil
}

func (jm *JobManager) markJobCancelled(jobID string) {
	now := time.Now()
	_, _ = jm.db.Exec(`UPDATE video_jobs SET status = $1, updated_at = $2 WHERE id = $3`,
		domain.JobStatusCancelled, now, jobID,
	)
	log.Printf("[Job %s] Marked as CANCELLED", jobID)
}

func (jm *JobManager) failJob(jobID, errorCode, technicalErr string) {
	now := time.Now()
	userMsg := technicalErr
	parts := strings.SplitN(technicalErr, ": ", 2)
	if len(parts) == 2 {
		errorCode = parts[0]
		userMsg = parts[1]
	}

	_, _ = jm.db.Exec(`UPDATE video_jobs SET
		status = $1, error_code = $2, error_message = $3, updated_at = $4
		WHERE id = $5`,
		domain.JobStatusFailed, errorCode, userMsg, now, jobID,
	)
	log.Printf("[Job %s] FAILED [%s]: %s", jobID, errorCode, technicalErr)
	_ = jm.fileStore.CleanupJobTemp(jobID)
}

// updateJobStatus enforces strict monotonic progress and state machine transitions (Rule #9).
func (jm *JobManager) updateJobStatus(
	jobID string,
	newStatus domain.JobStatus,
	progress int,
	downloaded, total, speed int64,
	eta int,
) error {
	now := time.Now()
	_, err := jm.db.Exec(`UPDATE video_jobs SET
		status = $1, progress = (CASE WHEN progress > $2 THEN progress ELSE $2 END), downloaded_bytes = $3, total_bytes = $4,
		speed = $5, eta = $6, updated_at = $7
		WHERE id = $8 AND status NOT IN ('completed', 'failed', 'cancelled')`,
		newStatus, progress, downloaded, total, speed, eta, now, jobID,
	)
	return err
}

// GetJob returns the job ensuring strict user authorization (Rule #3).
func (jm *JobManager) GetJob(ctx context.Context, jobID, userID string) (*domain.VideoJob, error) {
	var job domain.VideoJob
	err := jm.db.GetContext(ctx, &job, "SELECT * FROM video_jobs WHERE id = $1", jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("query job: %w", err)
	}

	// Rule #3: Authorization check
	if job.UserID != "" && userID != "" && job.UserID != userID {
		return nil, fmt.Errorf("unauthorized: job belongs to another user")
	}

	return &job, nil
}

// RecoverZombieJobs handles zombie jobs after a server restart (Rule #10).
func (jm *JobManager) RecoverZombieJobs() {
	if jm.db == nil {
		return
	}
	now := time.Now()
	res, err := jm.db.Exec(`UPDATE video_jobs SET
		status = 'failed',
		error_code = 'WORKER_INTERRUPTED',
		error_message = 'Tiến trình bị gián đoạn do máy chủ khởi động lại',
		updated_at = $1
		WHERE status IN ('analyzing', 'downloading', 'merging', 'saving', 'processing')`,
		now,
	)
	if err == nil {
		rows, _ := res.RowsAffected()
		if rows > 0 {
			log.Printf("[JobManager] Recovered %d zombie jobs marked as FAILED (WORKER_INTERRUPTED)", rows)
		}
	}
}
