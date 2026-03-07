package app

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
	once    sync.Once
	handler http.Handler
)

func Handler() http.Handler {
	once.Do(func() {
		cfg := config.Load()

		db := database.Connect(cfg.DatabaseURL)

		jwtUtil, err := utils.NewJWTUtil(cfg)
		if err != nil {
			log.Fatalf("Failed to initialise JWT utility: %v", err)
		}

		rdb := database.ConnectRedis(cfg.RedisURL)

		router := routes.SetupRouter(cfg, db, jwtUtil, rdb)

		handler = router
	})
	return handler
}
