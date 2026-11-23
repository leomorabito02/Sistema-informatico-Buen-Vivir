package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Port                      string
	GRPCPort                  string
	AuthGRPCURL               string
	AuthServiceURL            string
	UpstreamIDToken           bool
	JWTSecret                 string
	ResendAPIKey              string
	ResendFrom                string
	ResendFromNotifications   string
	ResendFromAlerts          string
	DatabaseURL               string
	InternalHMACSecret        string
	PubSubPushAudience        string
	PubSubPushServiceAccounts []string
	PubSubPushAuth            string
	FrontendURL               string
	Environment               string
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

	env := strings.TrimSpace(os.Getenv("ENV"))
	if env == "" {
		env = strings.TrimSpace(os.Getenv("APP_ENV"))
	}

	authServiceURL := strings.TrimRight(strings.TrimSpace(os.Getenv("AUTH_SERVICE_URL")), "/")
	authGRPCURL := strings.TrimSpace(os.Getenv("AUTH_GRPC_URL"))
	if authGRPCURL == "" && authServiceURL == "" && env != "production" && env != "testing" {
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

	frontendURL := resolveFrontendURL(os.Getenv("FRONTEND_URL"), env)

	return &Config{
		Port:                      port,
		GRPCPort:                  grpcPort,
		AuthGRPCURL:               authGRPCURL,
		AuthServiceURL:            authServiceURL,
		UpstreamIDToken:           strings.EqualFold(strings.TrimSpace(os.Getenv("UPSTREAM_ID_TOKEN")), "true"),
		JWTSecret:                 strings.TrimSpace(os.Getenv("JWT_SECRET")),
		ResendAPIKey:              os.Getenv("RESEND_API_KEY"),
		ResendFrom:                resendFrom,
		ResendFromNotifications:   resendFromNotifications,
		ResendFromAlerts:          resendFromAlerts,
		DatabaseURL:               os.Getenv("DATABASE_URL_NOTIFICATIONS"),
		InternalHMACSecret:        os.Getenv("INTERNAL_HMAC_SECRET"),
		PubSubPushAudience:        strings.TrimSpace(os.Getenv("PUBSUB_PUSH_AUDIENCE")),
		PubSubPushServiceAccounts: splitCommaList(os.Getenv("PUBSUB_PUSH_SERVICE_ACCOUNTS")),
		PubSubPushAuth:            strings.TrimSpace(os.Getenv("PUBSUB_PUSH_AUTH")),
		FrontendURL:               frontendURL,
		Environment:               env,
	}
}

func resolveFrontendURL(raw, env string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed != "" && !strings.Contains(trimmed, "localhost") && !strings.Contains(trimmed, "127.0.0.1") {
		return trimmed
	}
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "production", "prod":
		return "https://ceibuenvivir.org"
	default:
		return "https://testing.ceibuenvivir.pages.dev"
	}
}

func splitCommaList(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var out []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
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
