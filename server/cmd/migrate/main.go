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

	// Phase 1: create/update all tables (apps table added, app_id added to activity_events).
	err := db.AutoMigrate(
		&models.User{},
		&models.Device{},
		&models.Category{},
		&models.App{},
		&models.ActivityEvent{},
		&models.UserSetting{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Phase 2: back-fill apps from existing activity_events data.
	// This is idempotent — ON CONFLICT does nothing for already-migrated rows.
	log.Println("Back-filling apps table from existing events...")
	db.Exec(`
		INSERT INTO apps (id, user_id, name, platforms, icon, first_seen_at, last_seen_at)
		SELECT
			gen_random_uuid(),
			ae.user_id,
			ae.app_name,
			array_agg(DISTINCT d.platform),
			'',
			MIN(ae.start_time),
			MAX(ae.end_time)
		FROM activity_events ae
		JOIN devices d ON d.id = ae.device_id
		GROUP BY ae.user_id, ae.app_name
		ON CONFLICT (user_id, name) DO NOTHING
	`)

	// Phase 3: migrate category from events to apps (use most-recent event's category).
	log.Println("Migrating category assignments to apps table...")
	db.Exec(`
		UPDATE apps a
		SET category_id = latest.category_id
		FROM (
			SELECT DISTINCT ON (user_id, app_name)
				user_id, app_name, category_id
			FROM activity_events
			WHERE category_id IS NOT NULL
			ORDER BY user_id, app_name, start_time DESC
		) latest
		WHERE a.user_id = latest.user_id
		  AND a.name    = latest.app_name
		  AND a.category_id IS NULL
	`)

	// Phase 4: set app_id on existing events.
	log.Println("Linking events to apps via app_id...")
	db.Exec(`
		UPDATE activity_events ae
		SET app_id = a.id
		FROM apps a
		WHERE a.user_id = ae.user_id
		  AND a.name    = ae.app_name
		  AND ae.app_id IS NULL
	`)

	log.Println("Migrations completed successfully!")
}
