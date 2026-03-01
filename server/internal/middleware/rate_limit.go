package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	mu      sync.Mutex
	count   int
	resetAt time.Time
}

var rateLimitVisitors sync.Map

// RateLimit returns a per-IP fixed-window rate limiter middleware.
// rpm is the maximum number of requests allowed per IP per minute.
func RateLimit(rpm int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		val, _ := rateLimitVisitors.LoadOrStore(ip, &rateLimitEntry{resetAt: now.Add(time.Minute)})
		entry := val.(*rateLimitEntry)

		entry.mu.Lock()
		if now.After(entry.resetAt) {
			entry.count = 0
			entry.resetAt = now.Add(time.Minute)
		}
		entry.count++
		count := entry.count
		entry.mu.Unlock()

		if count > rpm {
			utils.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests", "Please slow down and try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
