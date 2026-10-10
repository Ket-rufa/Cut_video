package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"social-video-cutter/api/internal/domain"
)

// JobEnqueuer is the interface for enqueuing jobs.
// Services depend on this interface for testability.
type JobEnqueuer interface {
	EnqueueJob(ctx context.Context, job *domain.VideoJob) error
}

type Producer struct {
	redis *RedisClient
}

func NewProducer(redis *RedisClient) *Producer {
	return &Producer{redis: redis}
}

func (p *Producer) EnqueueJob(ctx context.Context, job *domain.VideoJob) error {
	msg := domain.JobMessage{
		JobID:   job.ID,
		VideoID: job.VideoID.String,
		Type:    string(job.Type),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal job message: %w", err)
	}
	return p.redis.client.RPush(ctx, QueueAll, data).Err()
}

func (p *Producer) SetJobProgress(ctx context.Context, jobID string, progress int) error {
	key := fmt.Sprintf("job:progress:%s", jobID)
	return p.redis.client.Set(ctx, key, progress, 2*time.Hour).Err()
}
