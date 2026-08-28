package services

// ValidationError señala que el payload recibido no cumple las reglas
// de negocio (campo requerido faltante, valor fuera del enum, etc.).
// Los controllers lo reconocen para responder 400 en vez de 500.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

// IsValidationError permite a la capa de controllers distinguir un
// error de validación de negocio de un error de infraestructura
// (por ejemplo, Contentful caído o devolviendo 5xx).
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// ForbiddenError señala que el usuario autenticado no tiene permiso
// sobre el recurso puntual solicitado. Es distinto de un rol
// insuficiente (eso lo corta RequireRole antes de llegar al service):
// esto es una verificación de ownership, por ejemplo un docente
// intentando editar un curso del que no es responsable.
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

func NewForbiddenError(message string) error {
	return &ForbiddenError{Message: message}
}

func IsForbiddenError(err error) bool {
	_, ok := err.(*ForbiddenError)
	return ok
}
