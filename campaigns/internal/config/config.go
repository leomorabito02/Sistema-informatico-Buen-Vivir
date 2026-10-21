package config

import (
	"os"
	"strings"
)

// DefaultPort es el puerto del servicio si no se define PORT (el mismo
// que declaraban el Dockerfile y el .env.example del esqueleto).
const DefaultPort = "8080"

// Config agrupa la configuración del servicio de Campañas.
type Config struct {
	Port string

	// CORSAllowedOrigins: orígenes del frontend separados por coma. Vacío
	// = no se habilita CORS (el acceso normal es por el API Gateway).
	CORSAllowedOrigins string
}

// LoadConfig lee la configuración del entorno. JWT_SECRET y las
// credenciales de Firestore las leen directamente JWTService y
// NewFirestoreClient, con el mismo criterio que human-promotion.
func LoadConfig() *Config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = DefaultPort
	}
	return &Config{
		Port:               port,
		CORSAllowedOrigins: os.Getenv("CORS_ALLOWED_ORIGINS"),
	}
}
