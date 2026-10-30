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

// SkillKey devuelve la clave normalizada de un nombre de habilidad: en
// minúsculas, sin tildes, sin signos de puntuación y con las palabras
// unidas por guiones. "Trabajo en equipo", " trabajo  en Equipo." y
// "Trabajo en Équipo" dan todas "trabajo-en-equipo".
//
// Por qué existe: es la única regla que decide si dos habilidades son
// "la misma". Se usa como id del documento en el catálogo (así Firestore
// impide el duplicado) y para filtrar las sugerencias de la IA contra lo
// que ya existe. No depende de que la IA escriba igual que el catálogo.
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
