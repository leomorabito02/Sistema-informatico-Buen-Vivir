package config

import (
	"log/slog"
	"os"
)

// InitLogger configures the default slog logger to output JSON.
// Google Cloud Logging automatically parses JSON logs and integrates them into its system.
func InitLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}
