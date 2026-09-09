package controllers_test

import (
	"net/http"

	"buenvivir/education/internal/repositories"
)

// repoErrNotFound y repoErrConflict dan errores reconocibles por
// writeError (ver http_helpers.go) sin depender de una respuesta real
// de Contentful/Firestore.
func repoErrNotFound() error {
	return repositories.ErrNotFound
}

func repoErrConflict() error {
	return &repositories.ContentfulAPIError{StatusCode: http.StatusConflict, Body: "version conflict"}
}
