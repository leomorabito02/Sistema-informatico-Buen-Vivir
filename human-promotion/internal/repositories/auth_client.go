package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"buenvivir/human-promotion/api/proto/authpb"
	"buenvivir/human-promotion/internal/models"
)

// authCallTimeout acota cada llamada a Auth para que una caída del
// servicio no deje colgadas las requests de este.
const authCallTimeout = 5 * time.Second

// ErrInvalidSearch señala que Auth rechazó el texto de búsqueda
// (menos de 2 o más de 100 caracteres).
var ErrInvalidSearch = errors.New("búsqueda inválida")

// AuthDirectory consulta por gRPC los datos de participantes y usuarios
// que viven en auth-service (nombre, apellido, fecha de nacimiento).
type AuthDirectory interface {
	GetParticipante(ctx context.Context, participanteID string) (*models.ParticipanteInfo, error)
	GetParticipanteByUserID(ctx context.Context, userID string) (*models.ParticipanteInfo, error)
	SearchParticipantes(ctx context.Context, query string, limit int) ([]models.ParticipanteInfo, error)
	// GetNombresUsuarios devuelve "Nombre Apellido" por user_id; los ids
	// que Auth no encuentra no aparecen en el mapa.
	GetNombresUsuarios(ctx context.Context, userIDs []string) (map[string]string, error)
	Close() error
}

type authDirectory struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

// NewAuthDirectory abre el cliente gRPC hacia auth-service, con el mismo
// criterio que labor-integration (AUTH_SERVICE_GRPC_TARGET).
func NewAuthDirectory(target string) (AuthDirectory, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar con auth-service: %w", err)
	}
	return &authDirectory{conn: conn, client: authpb.NewAuthServiceClient(conn)}, nil
}

func (a *authDirectory) GetParticipante(ctx context.Context, participanteID string) (*models.ParticipanteInfo, error) {
	return a.getProfile(ctx, &authpb.GetParticipantProfileRequest{
		Lookup: &authpb.GetParticipantProfileRequest_ParticipantId{ParticipantId: participanteID},
	})
}

func (a *authDirectory) GetParticipanteByUserID(ctx context.Context, userID string) (*models.ParticipanteInfo, error) {
	return a.getProfile(ctx, &authpb.GetParticipantProfileRequest{
		Lookup: &authpb.GetParticipantProfileRequest_UserId{UserId: userID},
	})
}

func (a *authDirectory) getProfile(ctx context.Context, req *authpb.GetParticipantProfileRequest) (*models.ParticipanteInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, authCallTimeout)
	defer cancel()

	profile, err := a.client.GetParticipantProfile(ctx, req)
	if status.Code(err) == codes.NotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error consultando el participante en auth-service: %w", err)
	}
	info := toParticipanteInfo(profile)
	return &info, nil
}

func (a *authDirectory) SearchParticipantes(ctx context.Context, query string, limit int) ([]models.ParticipanteInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, authCallTimeout)
	defer cancel()

	resp, err := a.client.SearchParticipants(ctx, &authpb.SearchParticipantsRequest{Query: query, Limit: int32(limit)}) //nolint:gosec // limit lo acota el servicio (≤ 50)
	if status.Code(err) == codes.InvalidArgument {
		return nil, ErrInvalidSearch
	}
	if err != nil {
		return nil, fmt.Errorf("error buscando participantes en auth-service: %w", err)
	}
	out := make([]models.ParticipanteInfo, 0, len(resp.GetParticipants()))
	for _, p := range resp.GetParticipants() {
		out = append(out, toParticipanteInfo(p))
	}
	return out, nil
}

func (a *authDirectory) GetNombresUsuarios(ctx context.Context, userIDs []string) (map[string]string, error) {
	out := make(map[string]string, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	ctx, cancel := context.WithTimeout(ctx, authCallTimeout)
	defer cancel()

	resp, err := a.client.GetUsersByIDs(ctx, &authpb.GetUsersByIDsRequest{UserIds: userIDs})
	if err != nil {
		return nil, fmt.Errorf("error consultando usuarios en auth-service: %w", err)
	}
	for _, u := range resp.GetUsers() {
		out[u.GetUserId()] = joinNombre(u.GetFirstName(), u.GetLastName())
	}
	return out, nil
}

func (a *authDirectory) Close() error {
	return a.conn.Close()
}

func toParticipanteInfo(p *authpb.ParticipantProfile) models.ParticipanteInfo {
	return models.ParticipanteInfo{
		ParticipanteID:  p.GetParticipantId(),
		UserID:          p.GetUserId(),
		Nombre:          p.GetFirstName(),
		Apellido:        p.GetLastName(),
		FechaNacimiento: p.GetBirthdate(),
	}
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
