package services

import (
	"context"

	"buenvivir/education/internal/repositories"
)

// participanteDirectorio es lo que education sabe de un participante a
// partir de auth-service: su nombre (vive en Usuario, ver DER), su DNI y
// si está aprobado.
type participanteDirectorio struct {
	Nombre   string
	Apellido string
	DNI      string
	Aprobado bool
}

// directorioParticipantes cruza GET /api/v1/participants con
// GET /api/v1/users (las dos con el token del usuario, así auth-service
// aplica su propio control de acceso) y devuelve un mapa por
// participant_id. Si falla la lista de usuarios, los participantes
// salen igual, sin nombre.
func directorioParticipantes(ctx context.Context, auth *repositories.AuthServiceClient, bearerToken string) (map[string]participanteDirectorio, error) {
	participants, err := auth.ListParticipants(ctx, bearerToken)
	if err != nil {
		return nil, err
	}
	nombres := map[string][2]string{}
	if users, err := auth.ListUsers(ctx, bearerToken); err == nil {
		for _, u := range users {
			nombres[u.UserID] = [2]string{u.FirstName, u.LastName}
		}
	}
	out := make(map[string]participanteDirectorio, len(participants))
	for _, p := range participants {
		n := nombres[p.UserID]
		out[p.ParticipantID] = participanteDirectorio{
			Nombre:   n[0],
			Apellido: n[1],
			DNI:      p.DNI,
			Aprobado: isParticipantApproved(p.ApprovalStatus),
		}
	}
	return out, nil
}

// nombresDeUsuarios devuelve "Nombre Apellido" por user_id. Lo usa la
// auditoría, que guarda el user_id de quien hizo el cambio.
func nombresDeUsuarios(ctx context.Context, auth *repositories.AuthServiceClient, bearerToken string) (map[string]string, error) {
	users, err := auth.ListUsers(ctx, bearerToken)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(users))
	for _, u := range users {
		out[u.UserID] = joinNombre(u.FirstName, u.LastName)
	}
	return out, nil
}

func joinNombre(nombre, apellido string) string {
	switch {
	case nombre == "":
		return apellido
	case apellido == "":
		return nombre
	default:
		return nombre + " " + apellido
	}
}
