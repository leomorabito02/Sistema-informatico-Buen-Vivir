package services

import (
	"errors"
	"regexp"
	"strings"
)

// ErrParticipantGmail: decisión del 15/11/2026 sobre HU-01, el email de
// un participante tiene que ser de Gmail (el resto de los roles puede
// usar cualquier dominio).
var ErrParticipantGmail error = &validacionError{"el correo de un participante tiene que ser una cuenta de Gmail (@gmail.com)"}

// ErrValidacion marca los errores de datos inválidos del alta: el
// controlador responde 400 (antes cualquier error del servicio era 500).
var ErrValidacion = errors.New("datos inválidos")

type validacionError struct{ msg string }

func (e *validacionError) Error() string        { return e.msg }
func (e *validacionError) Is(target error) bool { return target == ErrValidacion }

var dniRe = regexp.MustCompile(`^\d{7,8}$`)

// IsGmail dice si el email es de dominio gmail.com.
func IsGmail(email string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(email)), "@gmail.com")
}

// validateParticipantAccount valida el alta de un participante desde
// Usuarios (HU-07): Gmail, DNI de 7 u 8 dígitos y dirección. El
// teléfono ya es obligatorio para cualquier cuenta.
func validateParticipantAccount(email, dni, address string) error {
	if !IsGmail(email) {
		return ErrParticipantGmail
	}
	if !dniRe.MatchString(strings.TrimSpace(dni)) {
		return &validacionError{"el DNI del participante es obligatorio y debe tener 7 u 8 números, sin puntos"}
	}
	if len(strings.TrimSpace(address)) < 3 {
		return &validacionError{"la dirección del participante es obligatoria"}
	}
	return nil
}
