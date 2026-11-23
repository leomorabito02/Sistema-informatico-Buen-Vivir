package utils

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// ValidateURLOrRelative validates that a string is a valid absolute URL or a canonical relative path.
func ValidateURLOrRelative(val string) bool {
	if val == "" {
		return true
	}
	if strings.HasPrefix(val, "/") {
		return true
	}
	parsed, err := url.ParseRequestURI(val)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return true
}

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

		_ = validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
			return ValidatePasswordComplexity(fl.Field().String()) == nil
		})

		_ = validate.RegisterValidation("url_or_relative", func(fl validator.FieldLevel) bool {
			return ValidateURLOrRelative(fl.Field().String())
		})

		// dni: 7 u 8 dígitos sin puntos (HU-07).
		_ = validate.RegisterValidation("dni", func(fl validator.FieldLevel) bool {
			return dniPattern.MatchString(strings.TrimSpace(fl.Field().String()))
		})
	}
	return validate
}

func FormatValidationError(err error) map[string]string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorsMap := make(map[string]string)
		for _, e := range validationErrors {
			errorsMap[e.Field()] = e.Tag()
		}
		return errorsMap
	}
	return nil
}

var dniPattern = regexp.MustCompile(`^\d{7,8}$`)
