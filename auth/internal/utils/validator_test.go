package utils_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"auth/internal/utils"
)

type DummyStruct struct {
	Name     string `validate:"required"`
	Password string `validate:"password"`
}

func TestGetValidator(t *testing.T) {
	v1 := utils.GetValidator()
	assert.NotNil(t, v1)
	
	// Should be a singleton
	v2 := utils.GetValidator()
	assert.Equal(t, v1, v2)
}

func TestValidator_PasswordCustom(t *testing.T) {
	v := utils.GetValidator()
	
	valid := &DummyStruct{Name: "Test", Password: "RightPassword123!"}
	err := v.Struct(valid)
	assert.NoError(t, err)

	invalid := &DummyStruct{Name: "Test", Password: "weak"}
	err = v.Struct(invalid)
	assert.Error(t, err)
}

func TestFormatValidationError(t *testing.T) {
	v := utils.GetValidator()

	invalid := &DummyStruct{Name: "", Password: "weak"}
	err := v.Struct(invalid)
	assert.Error(t, err)

	formatted := utils.FormatValidationError(err)
	assert.NotNil(t, formatted)
	assert.Equal(t, "required", formatted["Name"])
	assert.Equal(t, "password", formatted["Password"])

	// Test non-validation error
	genericErr := errors.New("generic error")
	formatted2 := utils.FormatValidationError(genericErr)
	assert.Nil(t, formatted2)
}
