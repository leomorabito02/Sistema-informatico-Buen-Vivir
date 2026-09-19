package services

// participantApprovalStatusAccepted es el valor real que auth-service
// guarda en `approval_status` para un participante habilitado (Neon,
// tabla `participantes`) — confirmado 2026-09-13 contra datos reales:
// es "accepted" (no "approved", que es lo que decía el validador de
// auth/internal/dto/participant_dto.go y llevó a un falso "El
// participante no existe/no está activo" al probar con datos reales).
// Centralizado acá para que EnrollmentService, AttendanceService y
// GeneralAttendanceService no puedan volver a divergir en el literal.
const participantApprovalStatusAccepted = "accepted"

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
