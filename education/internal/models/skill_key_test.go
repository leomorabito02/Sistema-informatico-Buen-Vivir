package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkillKey_SameSkillWrittenDifferently(t *testing.T) {
	// Todas estas son "la misma habilidad" para el catálogo.
	variants := []string{
		"Trabajo en equipo",
		"trabajo en equipo",
		"  Trabajo   en  EQUIPO. ",
		"Trabajo en équipo",
		"trabajo-en-equipo",
	}
	for _, v := range variants {
		assert.Equal(t, "trabajo-en-equipo", SkillKey(v), v)
	}
}

func TestSkillKey_AccentsAndEnie(t *testing.T) {
	assert.Equal(t, "diseno-de-huerta", SkillKey("Diseño de huerta"))
	assert.Equal(t, "comunicacion", SkillKey("Comunicación"))
	assert.Equal(t, "linguistica", SkillKey("Lingüística"))
}

func TestSkillKey_EmptyOrOnlySymbols(t *testing.T) {
	assert.Equal(t, "", SkillKey(""))
	assert.Equal(t, "", SkillKey("  ...  "))
}

func TestCleanSkillName(t *testing.T) {
	assert.Equal(t, "Trabajo en equipo", CleanSkillName("  trabajo   en equipo "))
	assert.Equal(t, "Uso de Excel", CleanSkillName("uso de Excel"))
	assert.Equal(t, "Ñandú", CleanSkillName("ñandú"))
	assert.Equal(t, "", CleanSkillName("   "))
}

func TestSkillCategory_IsValid(t *testing.T) {
	for _, c := range SkillCategories {
		assert.True(t, c.IsValid(), c)
	}
	assert.False(t, SkillCategory("Técnica").IsValid())
	assert.False(t, SkillCategory("").IsValid())
}
