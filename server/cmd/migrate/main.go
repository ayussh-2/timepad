package main

import (
	"log"

	"github.com/ayussh-2/timepad/config"
	"github.com/ayussh-2/timepad/internal/database"
	"github.com/ayussh-2/timepad/internal/models"
)

func main() {
	cfg := config.Load()

	db := database.Connect(cfg.DatabaseURL)

	log.Println("Running Database Migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Device{},
		&models.Category{},
		&models.ActivityEvent{},
		&models.UserSetting{},
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations completed successfully!")
}
