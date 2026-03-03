package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ayussh-2/timepad/config"
	"github.com/ayussh-2/timepad/internal/models"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("Seeding database...")

	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 1. Create a Test User
	userID := uuid.New()
	user := models.User{
		ID:           userID,
		Email:        fmt.Sprintf("testuser_%d@timepad.com", time.Now().Unix()),
		PasswordHash: "$2a$10$wKqK.L3Kzq9O24J1U4T.xOW/Jk5GzqK064j/g1H4w.c.s1n.oU/rG", // Hash for 'password123'
		DisplayName:  "Seed User",
		Timezone:     "UTC",
	}
	db.Create(&user)

	// 2. Create a Device
	deviceID := uuid.New()
	device := models.Device{
		ID:        deviceID,
		UserID:    userID,
		Name:      "Test Laptop",
		Platform:  "windows",
		DeviceKey: fmt.Sprintf("seed-win-device-%d", time.Now().Unix()),
	}
	db.Create(&device)

	// 3. Create Categories
	codeCatID := uuid.New()
	browseCatID := uuid.New()
	isProductiveTrue := true
	isProductiveFalse := false
	db.Create([]models.Category{
		{ID: codeCatID, UserID: &userID, Name: "Coding", Color: "#4CAF50", IsProductive: &isProductiveTrue},
		{ID: browseCatID, UserID: &userID, Name: "Browsing", Color: "#2196F3", IsProductive: &isProductiveFalse},
	})

	// 4. Create Apps (upsert-style; associate category at app level)
	vsCodeAppID := uuid.New()
	chromeAppID := uuid.New()
	db.Create([]models.App{
		{
			ID: vsCodeAppID, UserID: userID, Name: "VS Code",
			Platforms: []string{"windows"}, CategoryID: &codeCatID,
			FirstSeenAt: time.Now().AddDate(0, 0, -6),
			LastSeenAt:  time.Now(),
		},
		{
			ID: chromeAppID, UserID: userID, Name: "Google Chrome",
			Platforms: []string{"windows"}, CategoryID: &browseCatID,
			FirstSeenAt: time.Now().AddDate(0, 0, -6),
			LastSeenAt:  time.Now(),
		},
	})

	// 5. Seed Random Events spanning the last 7 days
	var events []models.ActivityEvent
	now := time.Now()

	for d := 6; d >= 0; d-- {
		targetDay := now.AddDate(0, 0, -d).Truncate(24 * time.Hour).Add(10 * time.Hour) // Starting at 10 AM

		// 2 hours coding
		events = append(events, models.ActivityEvent{
			UserID:       userID,
			DeviceID:     deviceID,
			AppID:        &vsCodeAppID,
			AppName:      "VS Code",
			WindowTitle:  "timepad-server - server/main.go",
			StartTime:    targetDay,
			EndTime:      targetDay.Add(2 * time.Hour),
			DurationSecs: 7200,
			IsIdle:       false,
		})

		// 1 hour browsing
		events = append(events, models.ActivityEvent{
			UserID:       userID,
			DeviceID:     deviceID,
			AppID:        &chromeAppID,
			AppName:      "Google Chrome",
			WindowTitle:  "Golang Documentation",
			StartTime:    targetDay.Add(2 * time.Hour),
			EndTime:      targetDay.Add(3 * time.Hour),
			DurationSecs: 3600,
			IsIdle:       false,
		})

		// 30 mins idle
		events = append(events, models.ActivityEvent{
			UserID:       userID,
			DeviceID:     deviceID,
			AppName:      "Unknown",
			WindowTitle:  "Away",
			StartTime:    targetDay.Add(3 * time.Hour),
			EndTime:      targetDay.Add(3*time.Hour + 30*time.Minute),
			DurationSecs: 1800,
			IsIdle:       true,
		})
	}

	result := db.CreateInBatches(events, 50)
	if result.Error != nil {
		log.Fatalf("Failed to seed events: %v", result.Error)
	}

	fmt.Printf("Successfully DB Seeding! Inserted %d events for User: %s\n", len(events), user.Email)
}
