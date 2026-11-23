package services

import (
	"strings"

	"auth/internal/dto"
)

// MaxBusquedaUsuarios acota el texto de búsqueda de usuarios.
const MaxBusquedaUsuarios = 100

var sinTildes = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")

func normalizar(s string) string {
	return sinTildes.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// FiltrarUsuarios deja los usuarios del rol pedido (vacío = todos) cuyo nombre, apellido,
// email o teléfono contienen q, sin distinguir mayúsculas ni tildes.
func FiltrarUsuarios(users []dto.UserResponse, q, role string) []dto.UserResponse {
	q = normalizar(q)
	role = strings.TrimSpace(role)
	out := []dto.UserResponse{}
	for _, u := range users {
		if role != "" && u.RoleID != role {
			continue
		}
		if q != "" {
			nombre := normalizar(u.FirstName + " " + u.LastName)
			if !strings.Contains(nombre, q) && !strings.Contains(normalizar(u.Email), q) && !strings.Contains(u.Phone, q) {
				continue
			}
		}
		out = append(out, u)
	}
	return out
}
