package models

import (
	"strings"
	"unicode"
)

// accentFolding lleva las letras con tilde (y la ñ y la ü) a su letra
// base. Se hace a mano, sin golang.org/x/text, para no sumar una
// dependencia por una tabla de 14 letras: el catálogo es en español.
var accentFolding = map[rune]rune{
	'á': 'a', 'é': 'e', 'í': 'i', 'ó': 'o', 'ú': 'u',
	'à': 'a', 'è': 'e', 'ì': 'i', 'ò': 'o', 'ù': 'u',
	'ä': 'a', 'ë': 'e', 'ï': 'i', 'ö': 'o', 'ü': 'u',
	'ñ': 'n', 'ç': 'c',
}

// SkillKey devuelve la clave normalizada de un nombre de habilidad (minúsculas, sin
// tildes ni signos, palabras unidas por guiones). Decide si dos habilidades son la
// misma: es el id del catálogo y filtra las sugerencias de la IA.
func SkillKey(nombre string) string {
	var b strings.Builder
	lastDash := true // evita guiones al principio y repetidos
	for _, r := range strings.ToLower(nombre) {
		if f, ok := accentFolding[r]; ok {
			r = f
		}
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			// cualquier espacio o signo separa palabras
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// CleanSkillName prolija el nombre que se muestra: sin espacios de más y
// con mayúscula inicial ("  trabajo   en equipo " -> "Trabajo en
// equipo"). No toca el resto de las mayúsculas, para respetar siglas o
// nombres propios ("Uso de Excel").
func CleanSkillName(nombre string) string {
	nombre = strings.Join(strings.Fields(nombre), " ")
	if nombre == "" {
		return ""
	}
	runes := []rune(nombre)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
