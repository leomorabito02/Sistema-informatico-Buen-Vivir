package repositories

import (
	"context"
	"encoding/base64"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// NewFirestoreClient crea el cliente de Firestore compartido por los
// repositories de actividades, inscripciones y asistencias.
//
// GOOGLE_APPLICATION_CREDENTIALS puede venir de dos formas en este
// proyecto:
//   - un path a un archivo de credenciales (Application Default
//     Credentials estándar), o
//   - el JSON completo de la service account, crudo, directo en la
//     variable de entorno (mismo criterio que ya usa
//     auth/internal/services/firebase_service.go — pensado para
//     contenedores read_only sin volúmenes montados). Si el valor
//     empieza con '{', se toma como JSON y se usa directo con
//     option.WithCredentialsJSON, sin intentar abrirlo como archivo;
//     además se limpia la variable de entorno para que ninguna capa
//     interna de las librerías de Google (por ejemplo el transporte
//     gRPC de Firestore) intente resolverla de nuevo como path.
//
// firestore.DetectProjectID hace que el SDK infiera el project id a
// partir de esas credenciales, sin necesitar una variable de entorno
// propia para el project id.
//
// GOOGLE_APPLICATION_CREDENTIALS_BASE64 es una tercera forma: el JSON de
// la service account codificado en base64. Es lo que inyecta el pipeline
// de CD a Cloud Run (mismo criterio que auth-service con
// FIREBASE_CREDENTIALS_BASE64), porque un JSON crudo multilínea no entra
// limpio en la lista de env_vars de deploy-cloudrun. Tiene prioridad
// sobre GOOGLE_APPLICATION_CREDENTIALS.
func NewFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	credentialsJSON, err := GoogleCredentialsJSON()
	if err != nil {
		return nil, err
	}
	if credentialsJSON == nil {
		return firestore.NewClient(ctx, firestore.DetectProjectID)
	}
	_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
	return firestore.NewClient(ctx, firestore.DetectProjectID, option.WithCredentialsJSON(credentialsJSON)) //nolint:staticcheck // option.WithCredentialsJSON es necesario para credenciales inline (mismo criterio que auth/firebase_service.go)
}

// GoogleCredentialsJSON devuelve el JSON de la service account cuando
// viene inline (GOOGLE_APPLICATION_CREDENTIALS_BASE64, o
// GOOGLE_APPLICATION_CREDENTIALS con el JSON crudo), o nil cuando
// GOOGLE_APPLICATION_CREDENTIALS es un path / no está (ADC estándar).
// No modifica el entorno: por eso main lo llama antes que
// NewFirestoreClient (que sí limpia la variable) para pasarle las
// mismas credenciales al cliente de Cloud Storage.
func GoogleCredentialsJSON() ([]byte, error) {
	if b64 := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64")); b64 != "" {
		return base64.StdEncoding.DecodeString(b64)
	}
	credentialsEnv := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if strings.HasPrefix(credentialsEnv, "{") {
		return []byte(credentialsEnv), nil
	}
	return nil, nil
}
