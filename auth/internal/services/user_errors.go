package services

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// camposAltaUsuario traduce los campos de CreateUserRequest a nombres que
// entiende quien usa el panel.
var camposAltaUsuario = map[string]string{
	"FirstName": "el nombre",
	"LastName":  "el apellido",
	"Email":     "el correo electrónico",
	"Password":  "la contraseña",
	"Phone":     "el teléfono",
	"RoleID":    "el rol",
	"DNI":       "el DNI",
	"Address":   "la dirección",
}

// MensajeValidacionAlta arma un mensaje en castellano con los campos que no pasaron la
// validación del alta. Devuelve "" si err no es de validación.
func MensajeValidacionAlta(err error) string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) || len(ve) == 0 {
		return ""
	}
	partes := make([]string, 0, len(ve))
	for _, e := range ve {
		partes = append(partes, mensajeCampo(e))
	}
	return strings.Join(partes, ". ")
}

func mensajeCampo(e validator.FieldError) string {
	campo := camposAltaUsuario[e.Field()]
	if campo == "" {
		campo = strings.ToLower(e.Field())
	}
	switch e.Tag() {
	case "required":
		return "Falta completar " + campo
	case "email":
		return "El correo electrónico no tiene un formato válido"
	case "password":
		return "La contraseña debe tener al menos 8 caracteres, una mayúscula, una minúscula y un carácter especial"
	case "min":
		return "Revisá " + campo + ": es demasiado corto (mínimo " + e.Param() + " caracteres)"
	case "max":
		return "Revisá " + campo + ": es demasiado largo (máximo " + e.Param() + " caracteres)"
	default:
		return "Revisá " + campo
	}
}

// MensajeErrorBaseAlta traduce un error de la base (Postgres vía GORM) al
// crear una cuenta en un mensaje para el panel. El detalle técnico queda
// en el log del servicio; al usuario nunca le llega el texto de la base.
func MensajeErrorBaseAlta(err error) string {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key") || strings.Contains(msg, "duplicated key"):
		switch {
		case strings.Contains(msg, "email"):
			return "Ya existe una cuenta registrada con ese correo electrónico."
		case strings.Contains(msg, "dni"):
			return "Ya hay un participante registrado con ese DNI."
		case strings.Contains(msg, "phone"):
			return "Ya hay una cuenta registrada con ese teléfono."
		default:
			return "Ya existe una cuenta con alguno de esos datos. Revisá el correo, el DNI y el teléfono."
		}
	case strings.Contains(msg, "22001") || strings.Contains(msg, "value too long"):
		return "Alguno de los datos es demasiado largo. Acortalo e intentá de nuevo."
	case strings.Contains(msg, "23502") || strings.Contains(msg, "not-null") || strings.Contains(msg, "null value"):
		return "Falta completar algún dato obligatorio."
	case strings.Contains(msg, "23503") || strings.Contains(msg, "foreign key"):
		return "El rol elegido no es válido."
	default:
		return "No se pudo crear el usuario. Revisá los datos e intentá de nuevo; si el problema sigue, avisá al equipo técnico."
	}
}
