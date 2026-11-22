package services

import "strings"

// isApprovedStatus: auth guarda "approved", pero hay filas viejas con
// "accepted" (se usaba hasta el 28/09/2026). Las dos significan aprobado,
// igual que en auth (internal/services/approval_status.go) y education.
// Antes labor comparaba solo contra "approved" y rechazaba a esas
// empresas al crear ofertas y a esos participantes al postularse.
func isApprovedStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved", "accepted":
		return true
	}
	return false
}
