package config

import (
	"os"
	"strings"
)

// Config agrupa la configuración del servicio de Promoción Humana.
type Config struct {
	Port string

	CORSAllowedOrigins string

	// EncryptionKeySecret es el nombre completo del secreto en GCP Secret
	// Manager que guarda la clave AES-256 (32 bytes, en base64) con la
	// que se cifran situacion_actual, avances y la descripción de los
	// hitos, ej. "projects/<id>/secrets/human-promotion-enc-key/versions/latest".
	// Es la forma esperada en Cloud Run.
	EncryptionKeySecret string

	// EncryptionKeyBase64 es la alternativa para desarrollo local (Docker,
	// emulador de Firestore), donde no hay Secret Manager: la misma clave
	// de 32 bytes en base64, directo en la variable de entorno. Si
	// EncryptionKeySecret está seteada, tiene prioridad sobre esta.
	EncryptionKeyBase64 string

	// AuthServiceGRPCTarget es la dirección gRPC de auth-service, ej.
	// "auth-service:50051" en Docker (mismo nombre de variable que
	// labor-integration). De ahí salen nombre, apellido y fecha de
	// nacimiento de los participantes y la búsqueda por nombre.
	AuthServiceGRPCTarget string
}

// LoadConfig arma la configuración a partir de variables de entorno.
// La validez de la clave de cifrado no se chequea acá sino en
// crypto.LoadKey, que es quien sabe resolverla (Secret Manager o env).
func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	return &Config{
		Port:                port,
		CORSAllowedOrigins:  os.Getenv("CORS_ALLOWED_ORIGINS"),
		EncryptionKeySecret: strings.TrimSpace(os.Getenv("ENCRYPTION_KEY_SECRET")),
		EncryptionKeyBase64: strings.TrimSpace(os.Getenv("ENCRYPTION_KEY")),

		AuthServiceGRPCTarget: strings.TrimSpace(os.Getenv("AUTH_SERVICE_GRPC_TARGET")),
	}
}
