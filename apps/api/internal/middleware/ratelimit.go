package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Periodic cleanup of stale entries
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			rl.mu.Lock()
			cutoff := time.Now().Add(-rl.window)
			for key, timestamps := range rl.requests {
				var valid []time.Time
				for _, t := range timestamps {
					if t.After(cutoff) {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(rl.requests, key)
				} else {
					rl.requests[key] = valid
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	timestamps := rl.requests[key]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

// RateLimit creates a Gin middleware limiting requests per key (UserID + IP).
func RateLimit(limitPerMinute int, actionName string) gin.HandlerFunc {
	if limitPerMinute <= 0 {
		limitPerMinute = 60
	}
	rl := newRateLimiter(limitPerMinute, time.Minute)

	return func(c *gin.Context) {
		userID := GetUserID(c)
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s:%s:%s", actionName, userID, clientIP)

		if !rl.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": fmt.Sprintf("Bạn đã gửi quá nhiều yêu cầu %s. Vui lòng thử lại sau 1 phút.", actionName),
				},
			})
			return
		}

		c.Next()
	}
}
