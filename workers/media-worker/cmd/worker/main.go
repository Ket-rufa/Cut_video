package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"social-video-cutter/worker/internal/ffmpeg"
	"social-video-cutter/worker/internal/jobs"
	workerqueue "social-video-cutter/worker/internal/queue"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	dbDSN := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_DB", "video_cutter"),
		getEnv("POSTGRES_USER", "postgres"),
		getEnv("POSTGRES_PASSWORD", "postgres"),
	)
	redisAddr := fmt.Sprintf("%s:%s",
		getEnv("REDIS_HOST", "localhost"),
		getEnv("REDIS_PORT", "6379"),
	)
	storageRoot := getEnv("STORAGE_ROOT", "./storage")
	ffmpegPath := getEnv("FFMPEG_PATH", "/usr/bin/ffmpeg")
	ffprobePath := getEnv("FFPROBE_PATH", "/usr/bin/ffprobe")

	// Connect DB
	db, err := sqlx.Connect("postgres", dbDSN)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	// Connect Redis consumer
	consumer, err := workerqueue.NewConsumer(redisAddr)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer consumer.Close()

	ffRunner := ffmpeg.NewRunner(ffmpegPath, ffprobePath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Worker shutting down...")
		cancel()
	}()

	log.Println("Media worker started — waiting for jobs...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		default:
		}

		msg, err := consumer.Next(ctx)
		if err != nil {
			log.Printf("consume error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if msg == nil {
			continue // timeout, loop again
		}

		log.Printf("Processing job %s  type=%s  video=%s", msg.JobID, msg.Type, msg.VideoID)
		go processJob(ctx, db, ffRunner, consumer, storageRoot, msg)
	}
}

func processJob(
	ctx context.Context,
	db *sqlx.DB,
	ff *ffmpeg.Runner,
	consumer *workerqueue.Consumer,
	storageRoot string,
	msg *workerqueue.JobMessage,
) {
	jobCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Mark processing
	now := time.Now()
	_, _ = db.ExecContext(jobCtx,
		"UPDATE video_jobs SET status='processing', progress=10, started_at=$1, updated_at=$2 WHERE id=$3",
		now, now, msg.JobID)
	_ = consumer.SetProgress(jobCtx, msg.JobID, 10)

	// Get payload from DB
	var payloadStr string
	if err := db.GetContext(jobCtx, &payloadStr,
		"SELECT payload FROM video_jobs WHERE id=$1", msg.JobID); err != nil {
		markFailed(jobCtx, db, msg.JobID, fmt.Sprintf("get payload: %v", err))
		return
	}

	progressFn := func(pct int) {
		_ = consumer.SetProgress(jobCtx, msg.JobID, pct)
		_, _ = db.ExecContext(jobCtx,
			"UPDATE video_jobs SET progress=$1, updated_at=$2 WHERE id=$3",
			pct, time.Now(), msg.JobID)
		log.Printf("Job %s progress: %d%%", msg.JobID, pct)
	}

	var execErr error
	switch msg.Type {
	case "trim", "export":
		j := jobs.NewTrimJob(db, ff, storageRoot, progressFn)
		execErr = j.Execute(jobCtx, msg.JobID, msg.VideoID, json.RawMessage(payloadStr))
	default:
		execErr = fmt.Errorf("unknown job type: %s", msg.Type)
	}

	completedAt := time.Now()
	if execErr != nil {
		log.Printf("Job %s failed: %v", msg.JobID, execErr)
		markFailed(jobCtx, db, msg.JobID, execErr.Error())
		return
	}

	_ = consumer.SetProgress(jobCtx, msg.JobID, 100)
	_, _ = db.ExecContext(jobCtx,
		"UPDATE video_jobs SET status='completed', progress=100, completed_at=$1, updated_at=$2 WHERE id=$3",
		completedAt, completedAt, msg.JobID)
	log.Printf("Job %s completed successfully", msg.JobID)
}

func markFailed(ctx context.Context, db *sqlx.DB, jobID, errMsg string) {
	now := time.Now()
	_, _ = db.ExecContext(ctx,
		"UPDATE video_jobs SET status='failed', error_message=$1, updated_at=$2 WHERE id=$3",
		errMsg, now, jobID)
}
