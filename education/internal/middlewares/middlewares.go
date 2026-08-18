package middlewares

import "net/http"

// Middleware envuelve un http.Handler con comportamiento adicional
// (logging, seguridad, CORS, etc.). Se usa net/http puro, sin un
// router externo, siguiendo el mismo criterio de no sumar
// dependencias que no sean estrictamente necesarias.
type Middleware func(http.Handler) http.Handler

// Chain aplica los middlewares en el orden en que se los pasa: el
// primero de la lista es el que más "afuera" queda (se ejecuta
// primero en la request y último en la response).
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
