package repositories

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Errores de "no existe" que los controllers traducen a 404.
var (
	ErrApplicationNotFound = errors.New("application not found")
	ErrOfferNotFound       = errors.New("offer not found")
)

// notFoundAs devuelve target si err es el NotFound de Firestore; si no,
// devuelve err sin cambios.
func notFoundAs(err, target error) error {
	if status.Code(err) == codes.NotFound {
		return target
	}
	return err
}
