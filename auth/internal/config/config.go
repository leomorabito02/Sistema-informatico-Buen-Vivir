package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DATABASE_URL_AUTH")
	jwtSecret := os.Getenv("JWT_SECRET")
	

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
	}
}
