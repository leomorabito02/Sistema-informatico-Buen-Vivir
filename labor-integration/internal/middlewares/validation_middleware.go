package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Validator defines an interface for structs that can validate themselves.
type Validator interface {
	Validate() error
}

// ValidationMiddleware decodifica el cuerpo JSON en T, llama a Validate() y vuelve a
// dejar el cuerpo para que el handler lo lea.
func ValidationMiddleware[T Validator](next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqObj T

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(bodyBytes, &reqObj); err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		if err := reqObj.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Restore the body so next handlers can read it
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(w, r)
	})
}
