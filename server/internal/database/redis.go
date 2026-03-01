package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis parses the given URL and verifies connectivity.
// Returns nil if the URL is empty, invalid, or the server is unreachable —
// in that case the application degrades gracefully to synchronous ingestion.
func ConnectRedis(redisURL string) *redis.Client {
	if redisURL == "" {
		log.Println("REDIS_URL not set, running without async queue")
		return nil
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Failed to parse REDIS_URL: %v — running without async queue", err)
		return nil
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("Redis unreachable (%v) — running without async queue", err)
		return nil
	}

	log.Println("Redis connected successfully")
	return rdb
}
