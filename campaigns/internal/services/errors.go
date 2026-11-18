package services

// Mensajes de éxito que piden las historias de usuario.
const (
	MsgCampanaCreada    = "Campaña creada correctamente"    // HU-47
	MsgCampanaEditada   = "Campaña editada correctamente"   // HU-48
	MsgCampanaPublicada = "Campaña publicada correctamente" // HU-49 (la HU no fija texto)
	MsgCampanaEliminada = "Campaña eliminada correctamente" // Problema 5 de la auditoría (sin HU)
)

// ValidationError señala que el payload no cumple las reglas de negocio
// (campo obligatorio faltante, formato inválido, etc.). Los controllers
// lo reconocen para responder 400 en vez de 500.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}
