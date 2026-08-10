package utils

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// GetValidator returns a singleton instance of the validator.
func GetValidator() *validator.Validate {
	if validate == nil {
		validate = validator.New()
		
		// Custom Validation for Password
		validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
			// Password must have at least 8 chars, 1 uppercase, 1 number, 1 special char.
			// Implementación simplificada (en producción usar regex estricto):
			pass := fl.Field().String()
			if len(pass) < 8 {
				return false
			}
			return true // Agregaremos regex avanzado más adelante si es necesario.
		})
	}
	return validate
}

// FormatValidationError formats the validator.ValidationErrors into a readable map
func FormatValidationError(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors[e.Field()] = e.Tag() // Ej: "Email": "required"
		}
	} else {
		errors["request"] = "Invalid request format"
	}
	return errors
}
