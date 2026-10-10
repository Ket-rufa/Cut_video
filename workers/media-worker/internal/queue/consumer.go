package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const QueueAll = "queue:jobs"

type JobMessage struct {
	JobID   string `json:"job_id"`
	VideoID string `json:"video_id"`
	Type    string `json:"type"`
}

type Consumer struct {
	client *redis.Client
}

func NewConsumer(addr string) (*Consumer, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  35 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Consumer{client: rdb}, nil
}

// Next blocks up to 5s for the next job message.
func (c *Consumer) Next(ctx context.Context) (*JobMessage, error) {
	result, err := c.client.BLPop(ctx, 5*time.Second, QueueAll).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // timeout — no message
		}
		return nil, fmt.Errorf("blpop: %w", err)
	}
	if len(result) < 2 {
		return nil, nil
	}
	var msg JobMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return nil, fmt.Errorf("unmarshal job message: %w", err)
	}
	return &msg, nil
}

func (c *Consumer) SetProgress(ctx context.Context, jobID string, progress int) error {
	key := fmt.Sprintf("job:progress:%s", jobID)
	return c.client.Set(ctx, key, progress, 2*time.Hour).Err()
}

func (c *Consumer) Close() error { return c.client.Close() }
