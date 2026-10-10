package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
	"social-video-cutter/api/internal/ai"
	"social-video-cutter/api/internal/config"
	"social-video-cutter/api/internal/downloader"
	"social-video-cutter/api/internal/handler"
	"social-video-cutter/api/internal/middleware"
	"social-video-cutter/api/internal/platform"
	"social-video-cutter/api/internal/queue"
	"social-video-cutter/api/internal/repository"
	"social-video-cutter/api/internal/service"
	"social-video-cutter/api/internal/storage"
	"social-video-cutter/api/migrations"
)

func main() {
	cfg := config.Load()

	log.Printf("==================================================")
	log.Printf("Starting Social Video Cutter API (Environment: %s)", cfg.Env)
	log.Printf("==================================================")

	// Step 1: Startup Diagnostics / Dependency Check (Rule #21)
	diagnostics(cfg)

	// Step 2: Connect DB (Postgres with automatic SQLite fallback for dev)
	var db *sqlx.DB
	pgDB, err := sqlx.Connect("postgres", cfg.DB.DSN())
	if err != nil {
		if cfg.Env == "production" {
			log.Fatalf("CRITICAL [PRODUCTION]: connect postgres failed: %v", err)
		} else {
			_ = os.MkdirAll(cfg.Storage.Root, 0755)
			sqlitePath := filepath.Join(cfg.Storage.Root, "dev.db")
			log.Printf("[DEV FALLBACK] PostgreSQL not reachable (%v). Using local SQLite database (%s).", err, sqlitePath)
			sqliteDB, sErr := sqlx.Connect("sqlite", sqlitePath)
			if sErr != nil {
				log.Fatalf("connect sqlite fallback failed: %v", sErr)
			}
			db = sqliteDB
		}
	} else {
		db = pgDB
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
		log.Printf("Connected to PostgreSQL at %s:%s", cfg.DB.Host, cfg.DB.Port)
	}
	defer db.Close()

	// Step 3: Run migrations from dedicated migration runner (Rule #2)
	if err := migrations.RunMigrations(db); err != nil {
		log.Fatalf("database migrations failed: %v", err)
	}

	// Step 4: Connect Redis (Rule #1)
	var rdb *queue.RedisClient
	rdbClient, err := queue.NewRedisClient(cfg.Redis.Addr())
	if err != nil {
		if cfg.Env == "production" {
			log.Fatalf("CRITICAL [PRODUCTION]: connect redis failed: %v", err)
		} else {
			log.Printf("[DEV WARNING] Redis not reachable at %s (%v). Using local in-memory processing.", cfg.Redis.Addr(), err)
		}
	} else {
		rdb = rdbClient
		defer rdb.Close()
		log.Printf("Connected to Redis at %s", cfg.Redis.Addr())
	}

	// Step 5: Wire Core Storage & Downloader Adapters (Open-Closed Principle, Rule #23)
	localStorage := storage.NewLocalStorage(cfg.Storage.Root)
	analysisCache := downloader.NewAnalysisCache(30 * time.Minute)

	ytDownloader := downloader.NewYouTubeDownloader(cfg.Downloader.YTDLPPath, cfg.Downloader.FFmpegPath, cfg.Downloader.FFprobePath)
	tkDownloader := downloader.NewTikTokDownloader(cfg.Downloader.YTDLPPath, cfg.Downloader.FFmpegPath, cfg.Downloader.FFprobePath)
	dyDownloader := downloader.NewDouyinDownloader(cfg.Downloader.YTDLPPath, cfg.Downloader.FFmpegPath, cfg.Downloader.FFprobePath)
	directDownloader := downloader.NewDirectDownloader(cfg.Downloader.FFmpegPath, cfg.Downloader.FFprobePath)

	registry := downloader.NewRegistry(ytDownloader, tkDownloader, dyDownloader, directDownloader)
	jobManager := downloader.NewJobManager(db, localStorage, registry, analysisCache, cfg)

	// Wire legacy & auxiliary repositories and services
	videoRepo := repository.NewVideoRepository(db)
	jobRepo := repository.NewJobRepository(db)
	clipRepo := repository.NewClipRepository(db)
	libRepo := repository.NewLibraryRepository(db)
	subRepo := repository.NewSubtitleRepository(db)

	var producer queue.JobEnqueuer
	if rdb != nil {
		producer = queue.NewProducer(rdb)
	}
	resolver := platform.NewResolver()

	videoSvc := service.NewVideoService(videoRepo, resolver)
	exportSvc := service.NewExportService(jobRepo, producer)
	clipSvc := service.NewClipService(clipRepo)
	libSvc := service.NewLibraryService(libRepo)
	geminiSvc := ai.NewGeminiService(cfg, subRepo)
	dubbingSvc := ai.NewDubbingService(cfg, subRepo)

	videoHandler := handler.NewVideoHandler(videoSvc, cfg.Storage.Root, jobManager, registry, analysisCache)
	jobHandler := handler.NewJobHandler(exportSvc, cfg.Storage.Root)
	clipHandler := handler.NewClipHandler(clipSvc, libSvc)
	aiHandler := handler.NewAIHandler(geminiSvc, dubbingSvc, videoSvc, subRepo, cfg.Storage.Root)

	// Step 6: Gin Router Setup
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	corsOrigins := []string{"http://localhost:5173", "http://localhost"}
	for _, o := range strings.Split(cfg.CORS.Origins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			corsOrigins = append(corsOrigins, o)
		}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-User-Id"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(middleware.AuthMiddleware())

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"env":    cfg.Env,
			"time":   time.Now(),
			"dependencies": gin.H{
				"yt-dlp":  cfg.Downloader.YTDLPPath,
				"ffmpeg":  cfg.Downloader.FFmpegPath,
				"ffprobe": cfg.Downloader.FFprobePath,
			},
		})
	})

	// Rule #20: Strict API v1 route grouping
	v1 := r.Group("/api/v1")
	{
		// Real Video Downloader APIs (YouTube, TikTok & Douyin)
		v1.POST("/videos/analyze",
			middleware.RateLimit(cfg.Limits.AnalyzeRateLimit, "phân tích video"),
			videoHandler.Analyze,
		)
		v1.POST("/videos/download",
			middleware.RateLimit(cfg.Limits.CreateJobRateLimit, "tải video"),
			videoHandler.Download,
		)
		v1.GET("/videos/jobs/:jobId", videoHandler.GetJob)
		v1.POST("/videos/jobs/:jobId/cancel", videoHandler.CancelJob)
		v1.GET("/videos/:id/file", videoHandler.GetVideoFile)
		v1.HEAD("/videos/:id/file", videoHandler.GetVideoFile)

		// Video Library APIs
		v1.GET("/videos", videoHandler.List)
		v1.GET("/videos/:id", videoHandler.GetByID)
		v1.PATCH("/videos/:id", videoHandler.Update)
		v1.POST("/videos/:id", videoHandler.Update)
		v1.DELETE("/videos/:id", videoHandler.Delete)
		v1.POST("/videos/upload", videoHandler.Upload)
		v1.POST("/videos/resolve", videoHandler.Resolve)
		v1.POST("/videos/import", videoHandler.Resolve)

		// Clips
		v1.POST("/videos/:id/clips", clipHandler.CreateClip)
		v1.GET("/videos/:id/clips", clipHandler.ListClips)
		v1.DELETE("/clips/:id", clipHandler.DeleteClip)

		// Folders
		v1.GET("/folders", clipHandler.ListFolders)
		v1.POST("/folders", clipHandler.CreateFolder)
		v1.DELETE("/folders/:id", clipHandler.DeleteFolder)

		// Tags
		v1.GET("/tags", clipHandler.ListTags)
		v1.POST("/tags", clipHandler.CreateTag)
		v1.DELETE("/tags/:id", clipHandler.DeleteTag)

		// History
		v1.GET("/history", clipHandler.ListHistory)
		v1.POST("/history", clipHandler.CreateHistory)

		// Trim & Export Jobs
		v1.POST("/videos/:id/trim", jobHandler.CreateTrimJob)
		v1.POST("/videos/:id/export", jobHandler.CreateTrimJob)
		v1.GET("/jobs", jobHandler.ListJobs)
		v1.GET("/jobs/:id", jobHandler.GetJob)
		v1.GET("/jobs/:id/progress", jobHandler.GetJobProgress)

		v1.GET("/exports/:id", jobHandler.GetExport)
		v1.GET("/exports/:id/download", jobHandler.DownloadExport)

		// AI Translation & Subtitles
		v1.POST("/videos/:id/translate-meta", aiHandler.TranslateMeta)
		v1.POST("/videos/:id/subtitles/generate", aiHandler.GenerateSubtitles)
		v1.GET("/videos/:id/subtitles", aiHandler.GetSubtitles)
		v1.PUT("/videos/:id/subtitles", aiHandler.UpdateSubtitles)
		v1.POST("/videos/:id/subtitles/translate-cue", aiHandler.TranslateCue)
		v1.GET("/videos/:id/subtitles/download", aiHandler.DownloadSubtitles)
		v1.GET("/videos/:id/subtitles/cues/:index/audio", aiHandler.StreamCueAudio)

		// AI Dubbing (TTS & Audio Sync)
		v1.GET("/dubbing/voices", aiHandler.GetDubbingVoices)
		v1.GET("/dubbing/preview", aiHandler.PreviewVoice)
		v1.HEAD("/dubbing/preview", aiHandler.PreviewVoice)
		v1.POST("/videos/:id/dubbing/generate", aiHandler.GenerateDubbing)
		v1.GET("/videos/:id/dubbing", aiHandler.GetDubbing)
		v1.GET("/videos/:id/dubbing/audio", aiHandler.StreamDubbedAudio)
		v1.HEAD("/videos/:id/dubbing/audio", aiHandler.StreamDubbedAudio)
		v1.GET("/videos/:id/dubbing/video", aiHandler.StreamDubbedVideo)
		v1.HEAD("/videos/:id/dubbing/video", aiHandler.StreamDubbedVideo)

		// AI Settings & Diagnostics
		v1.GET("/settings/ai", aiHandler.GetAIConfig)
		v1.POST("/settings/ai", aiHandler.UpdateAIConfig)
		v1.POST("/settings/ai/test", aiHandler.TestAIConnection)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Minute,
		WriteTimeout: 15 * time.Minute,
		IdleTimeout:  15 * time.Minute,
	}

	go func() {
		log.Printf("API server listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server listen failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("Server exited cleanly")
}

func diagnostics(cfg *config.Config) {
	// Check yt-dlp
	out, err := exec.Command(cfg.Downloader.YTDLPPath, "--version").Output()
	if err != nil {
		msg := fmt.Sprintf("yt-dlp not found or error at %s: %v", cfg.Downloader.YTDLPPath, err)
		if cfg.Env == "production" {
			log.Fatalf("CRITICAL [PRODUCTION]: %s", msg)
		} else {
			log.Printf("[DEV WARNING] %s", msg)
		}
	} else {
		log.Printf("[DIAGNOSTICS] yt-dlp available: version %s (path: %s)", strings.TrimSpace(string(out)), cfg.Downloader.YTDLPPath)
	}

	// Check ffmpeg
	out, err = exec.Command(cfg.Downloader.FFmpegPath, "-version").Output()
	if err != nil {
		msg := fmt.Sprintf("ffmpeg not found or error at %s: %v", cfg.Downloader.FFmpegPath, err)
		if cfg.Env == "production" {
			log.Fatalf("CRITICAL [PRODUCTION]: %s", msg)
		} else {
			log.Printf("[DEV WARNING] %s", msg)
		}
	} else {
		firstLine := strings.Split(string(out), "\n")[0]
		log.Printf("[DIAGNOSTICS] ffmpeg available: %s (path: %s)", strings.TrimSpace(firstLine), cfg.Downloader.FFmpegPath)
	}

	// Check ffprobe
	out, err = exec.Command(cfg.Downloader.FFprobePath, "-version").Output()
	if err != nil {
		msg := fmt.Sprintf("ffprobe not found or error at %s: %v", cfg.Downloader.FFprobePath, err)
		if cfg.Env == "production" {
			log.Fatalf("CRITICAL [PRODUCTION]: %s", msg)
		} else {
			log.Printf("[DEV WARNING] %s", msg)
		}
	} else {
		firstLine := strings.Split(string(out), "\n")[0]
		log.Printf("[DIAGNOSTICS] ffprobe available: %s (path: %s)", strings.TrimSpace(firstLine), cfg.Downloader.FFprobePath)
	}
}
