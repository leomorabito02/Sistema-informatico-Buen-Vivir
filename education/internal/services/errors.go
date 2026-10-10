package services

// isParticipantApproved decide si un participante de auth-service está
// habilitado ("participantes activos del centro"). auth-service guarda
// "approved" (valor canónico) pero hay filas reales en Neon con
// "accepted", el valor que se escribía hasta 2026-09-28 — se aceptan
// los dos (ver auth/internal/services/approval_status.go). Vacío se
// tolera (defensivo: no ocultar participantes por un dato mal formado).
// Centralizado acá para que EnrollmentService y
// GeneralAttendanceService no puedan volver a divergir en el literal.
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

// UnavailableError señala que una dependencia externa opcional no pudo
// responder y el pedido no se puede completar ahora, pero el resto del
// sistema sigue funcionando (por ejemplo, ninguna IA pudo sugerir
// habilidades: se pueden cargar a mano). Los controllers responden 503
// con el mensaje tal cual, porque está pensado para mostrarse en pantalla.
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
