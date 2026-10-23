package main

import (
	"log/slog"
	"os"

	"buenvivir/notifications/internal/config"
	"buenvivir/notifications/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	if cfg.DatabaseURL == "" {
		slog.Error("DATABASE_URL_NOTIFICATIONS is required")
		os.Exit(1)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	slog.Info("Migrating notifications models to Neon PostgreSQL...")

	err = db.AutoMigrate(
		&models.Notification{},
		&models.NotificationRecipient{},
		&models.NotificationPreference{},
	)
	if err != nil {
		slog.Error("AutoMigrate failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Migration finished successfully")
}
