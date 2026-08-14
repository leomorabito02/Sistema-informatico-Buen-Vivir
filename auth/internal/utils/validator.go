package utils

import (
	"errors"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// ValidatePasswordComplexity checks that a password has >=8 chars, 1 uppercase, 1 lowercase, 1 special character.
func ValidatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return errors.New("la contraseña debe tener al menos 8 caracteres")
	}

	var hasUpper, hasLower, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case !unicode.IsLetter(ch) && !unicode.IsNumber(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("la contraseña debe incluir al menos una letra mayúscula")
	}
	if !hasLower {
		return errors.New("la contraseña debe incluir al menos una letra minúscula")
	}
	if !hasSpecial {
		return errors.New("la contraseña debe incluir al menos un carácter especial (ej: !@#$%^&*)")
	}
	return nil
}

// GetValidator returns a singleton instance of the validator.
func GetValidator() *validator.Validate {
	if validate == nil {
		validate = validator.New()

		// Custom Validation for Password
		_ = validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
			return ValidatePasswordComplexity(fl.Field().String()) == nil
		})
	}
	return validate
}

// FormatValidationError formats the validator.ValidationErrors into a readable map
func FormatValidationError(err error) map[string]string {
	errorsMap := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errorsMap[e.Field()] = e.Tag()
		}
	} else {
		errorsMap["request"] = "Invalid request format"
	}
	return errorsMap
}
