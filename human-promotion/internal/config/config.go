package config

import "os"

type Config struct {
	Port string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")

	return &Config{Port: port}
}
