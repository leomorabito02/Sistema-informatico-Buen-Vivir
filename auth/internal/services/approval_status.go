package services

// approvalStatusAccepted es el valor canónico que el resto del sistema
// espera para un perfil aprobado (login step 2, education-service vía
// participantApprovalStatusAccepted). Los DTOs de aprobación aceptan
// "approved" como input (así lo manda el frontend), pero se normaliza acá
// antes de persistir para no divergir del valor que realmente se compara
// en las lecturas.
const approvalStatusAccepted = "accepted"
