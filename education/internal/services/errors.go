package services

// isParticipantApproved decide si un participante está habilitado. Acepta "approved"
// y "accepted" (valor viejo que sigue en la base) y tolera el vacío.
func isParticipantApproved(status string) bool {
	return status == "" || status == "approved" || status == "accepted"
}

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

// ForbiddenError señala falta de permiso sobre un recurso puntual (ownership), por
// ejemplo un docente editando un curso ajeno.
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

// UnavailableError señala que una dependencia opcional no respondió (por ejemplo, la
// IA). Se responde 503 con el mensaje tal cual.
type UnavailableError struct {
	Message string
}

func (e *UnavailableError) Error() string {
	return e.Message
}

func NewUnavailableError(message string) error {
	return &UnavailableError{Message: message}
}

func IsUnavailableError(err error) bool {
	_, ok := err.(*UnavailableError)
	return ok
}
