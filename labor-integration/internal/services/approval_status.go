package services

import "strings"

// isApprovedStatus: auth guarda "approved", pero hay filas viejas con "accepted". Las
// dos significan aprobado (igual que en auth y education).
func isApprovedStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved", "accepted":
		return true
	}
	return false
}
