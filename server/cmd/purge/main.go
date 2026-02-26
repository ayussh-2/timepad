package main

import (
	"log"

	"github.com/ayussh-2/timepad/config"
	"github.com/ayussh-2/timepad/internal/database"
	"github.com/ayussh-2/timepad/internal/services"
)

func main() {
	cfg := config.Load()

	log.Println("Starting data retention purge job...")

	db := database.Connect(cfg.DatabaseURL)

	purgeService := services.NewPurgeService(db)

	totalPurged, err := purgeService.PurgeExpiredEvents()
	if err != nil {
		log.Fatalf("Purge job failed: %v", err)
	}

	log.Printf("Purge job completed. Total events purged: %d", totalPurged)
}
