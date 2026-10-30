package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"auth/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitDatabase initializes the connection to NeonDB (PostgreSQL) using GORM.
func InitDatabase() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL_AUTH")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL_AUTH environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection pool: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Successfully connected to the database")

	if err := db.AutoMigrate(&models.User{}, &models.EmailVerification{}); err != nil {
		slog.Warn("Failed to auto-migrate models", "error", err)
	}

	return db, nil
}
