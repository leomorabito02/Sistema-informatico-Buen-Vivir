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

// ValidationMiddleware creates a middleware that decodes the JSON request body
// into the provided generic struct T, calls Validate() if it implements Validator,
// and sets the decoded object in the request context or allows the controller to re-read it.
// To keep it simple and standard, we decode and re-encode the body so handlers can read it.
func ValidationMiddleware[T Validator](next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqObj T

		// Read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Decode JSON
		if err := json.Unmarshal(bodyBytes, &reqObj); err != nil {
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
			return
		}

		// Validate
		if err := reqObj.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Restore the body so next handlers can read it
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(w, r)
	})
}
