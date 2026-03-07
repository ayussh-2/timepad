// Package handler is the Vercel serverless entry point.
// Vercel's Go runtime looks for an exported http.HandlerFunc in /api/*.go files.
package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/ayussh-2/timepad/config"
	"github.com/ayussh-2/timepad/internal/database"
	"github.com/ayussh-2/timepad/internal/routes"
	"github.com/ayussh-2/timepad/internal/utils"
)

var (
	once        sync.Once
	httpHandler http.Handler
)

// Handler is the single HTTP entry point exposed to Vercel.
// All connections and the Gin router are initialised exactly once per
// cold-start using sync.Once, then reused across warm invocations.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		cfg := config.Load()

		db := database.Connect(cfg.DatabaseURL)

		jwtUtil, err := utils.NewJWTUtil(cfg)
		if err != nil {
			log.Fatalf("Failed to initialise JWT utility: %v", err)
		}

		// Redis is optional — ConnectRedis returns nil gracefully when unavailable.
		rdb := database.ConnectRedis(cfg.RedisURL)

		httpHandler = routes.SetupRouter(cfg, db, jwtUtil, rdb)
	})

	httpHandler.ServeHTTP(w, r)
}
