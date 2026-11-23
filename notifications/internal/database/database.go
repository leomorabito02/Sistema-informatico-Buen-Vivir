package database

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"buenvivir/notifications/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDatabase initializes the connection pool to PostgreSQL using GORM and runs schema auto-migrations.
func InitDatabase(dsn string) (*gorm.DB, error) {
	dsn = strings.TrimSpace(strings.Trim(strings.TrimSpace(dsn), `"'`))
	if dsn == "" {
		return nil, fmt.Errorf("database connection string is empty")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB handle: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Notification{},
		&models.NotificationRecipient{},
		&models.NotificationPreference{},
		&models.ProcessedEvent{},
	); err != nil {
		slog.Warn("Failed to auto-migrate notifications models", "error", err)
	}

	return db, nil
}
