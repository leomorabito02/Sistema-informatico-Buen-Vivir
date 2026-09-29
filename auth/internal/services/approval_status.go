package services

// approvalStatusApproved es el valor canónico que se persiste para un
// perfil aprobado ("approved", así lo manda el frontend y así lo filtran
// el login y los contadores de impacto).
//
// Hasta 2026-09-28 esta rama normalizaba a "accepted", y hay filas
// reales en Neon con ese valor (aprobadas en ese período), así que las
// LECTURAS aceptan los dos (isApprovedStatus). Education hace lo mismo
// del otro lado (services.isParticipantApproved).
const approvalStatusApproved = "approved"

// approvalStatusLegacyAccepted es el valor anterior, que solo se acepta
// al leer, nunca se escribe.
const approvalStatusLegacyAccepted = "accepted"

func isApprovedStatus(status string) bool {
	return status == approvalStatusApproved || status == approvalStatusLegacyAccepted
}
