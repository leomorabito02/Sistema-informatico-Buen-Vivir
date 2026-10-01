package services

// MsgCamposObligatorios es el mensaje exacto que pide HU-29 cuando falta
// un campo obligatorio de la ficha.
const MsgCamposObligatorios = "Completar todos los campos obligatorios para continuar"

// ValidationError señala que el payload recibido no cumple las reglas
// de negocio (campo requerido faltante, formato inválido, etc.).
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

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}
