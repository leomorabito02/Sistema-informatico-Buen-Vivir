package services

// approvalStatusApproved es el valor que se guarda para un perfil aprobado. Las lecturas
// aceptan también "accepted", un valor viejo que sigue en la base (isApprovedStatus).
const approvalStatusApproved = "approved"

// approvalStatusLegacyAccepted es el valor anterior, que solo se acepta
// al leer, nunca se escribe.
const approvalStatusLegacyAccepted = "accepted"

func isApprovedStatus(status string) bool {
	return status == approvalStatusApproved || status == approvalStatusLegacyAccepted
}
