package repositories

import (
	"context"
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
func NewFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	credentialsEnv := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if strings.HasPrefix(credentialsEnv, "{") {
		os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		return firestore.NewClient(ctx, firestore.DetectProjectID, option.WithCredentialsJSON([]byte(credentialsEnv)))
	}
	return firestore.NewClient(ctx, firestore.DetectProjectID)
}
