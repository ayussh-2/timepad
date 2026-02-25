package main

import (
	"log"

	"github.com/ayussh-2/timepad/config"
	"github.com/ayussh-2/timepad/internal/database"
	"github.com/ayussh-2/timepad/internal/routes"
	"github.com/ayussh-2/timepad/internal/utils"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting server in %s mode...", cfg.Env)
	log.Printf("Server listening on %s", cfg.ServerAddr)

	db := database.Connect(cfg.DatabaseURL)

	jwtUtil, err := utils.NewJWTUtil(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize JWT utility: %v", err)
	}

	router := routes.SetupRouter(cfg, db, jwtUtil)
	if err := router.Run(cfg.ServerAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
