package repositories

import (
	"errors"
	"net/http"
)

// ErrNotFound es el sentinel que usan los repositories que no hablan
// con Contentful (Firestore) para señalar "no existe", equivalente a
// un ContentfulAPIError con StatusCode 404. IsNotFound reconoce ambos
// casos para que la capa de controllers no tenga que distinguir de
// qué fuente de datos vino el error.
var ErrNotFound = errors.New("recurso no encontrado")

func IsNotFound(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	apiErr, ok := err.(*ContentfulAPIError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}
