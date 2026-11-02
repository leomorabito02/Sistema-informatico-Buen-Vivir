package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Port                    string
	GRPCPort                string
	AuthGRPCURL             string
	ResendAPIKey            string
	ResendFrom              string
	ResendFromNotifications string
	ResendFromAlerts        string
	DatabaseURL             string
	InternalHMACSecret      string
}

func LoadConfig() *Config {
	_ = loadEnvFile(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	authGRPCURL := os.Getenv("AUTH_GRPC_URL")
	if authGRPCURL == "" {
		authGRPCURL = "localhost:50051"
	}

	resendFrom := os.Getenv("RESEND_FROM_EMAIL")
	if resendFrom == "" {
		resendFrom = "onboarding@resend.dev"
	}

	resendFromNotifications := os.Getenv("RESEND_FROM_EMAIL_NOTIFICATIONS")
	if resendFromNotifications == "" {
		resendFromNotifications = resendFrom
	}

	resendFromAlerts := os.Getenv("RESEND_FROM_EMAIL_ALERTS")
	if resendFromAlerts == "" {
		resendFromAlerts = resendFromNotifications
	}

	return &Config{
		Port:                    port,
		GRPCPort:                grpcPort,
		AuthGRPCURL:             authGRPCURL,
		ResendAPIKey:            os.Getenv("RESEND_API_KEY"),
		ResendFrom:              resendFrom,
		ResendFromNotifications: resendFromNotifications,
		ResendFromAlerts:        resendFromAlerts,
		DatabaseURL:             os.Getenv("DATABASE_URL_NOTIFICATIONS"),
		InternalHMACSecret:      os.Getenv("INTERNAL_HMAC_SECRET"),
	}
}

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, value, ok := parseEnvLine(scanner.Text())
		if !ok {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func parseEnvLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}

	key, value, found := strings.Cut(line, "=")
	if !found {
		return "", "", false
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return "", "", false
	}

	return key, cleanEnvValue(strings.TrimSpace(value)), true
}

func cleanEnvValue(value string) string {
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}
