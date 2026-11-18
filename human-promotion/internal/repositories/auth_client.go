package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

// AuthDirectory consulta los datos de participantes y usuarios
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

type httpAuthDirectory struct {
	baseURL    string
	httpClient *http.Client
	idToken    func(ctx context.Context) (string, error)
}

type authProfileResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ParticipantID   string `json:"participant_id"`
		UserID          string `json:"user_id"`
		Nombre          string `json:"nombre"`
		Apellido        string `json:"apellido"`
		FechaNacimiento string `json:"fecha_nacimiento"`
	} `json:"data"`
}

type authSearchResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ParticipantID   string `json:"participant_id"`
		UserID          string `json:"user_id"`
		Nombre          string `json:"nombre"`
		Apellido        string `json:"apellido"`
		FechaNacimiento string `json:"fecha_nacimiento"`
	} `json:"data"`
}

type authSummariesResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		UserID   string `json:"user_id"`
		FullName string `json:"full_name"`
	} `json:"data"`
}

func (h *httpAuthDirectory) prepareRequest(ctx context.Context, method, targetURL string, body any) (*http.Request, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if h.idToken != nil {
		if tok, err := h.idToken(ctx); err == nil && tok != "" {
			req.Header.Set("X-Serverless-Authorization", "Bearer "+tok)
		}
	}
	if token := tokenFromContext(ctx); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}

func (h *httpAuthDirectory) GetParticipante(ctx context.Context, participanteID string) (*models.ParticipanteInfo, error) {
	targetURL := fmt.Sprintf("%s/api/v1/directory/participants/%s", h.baseURL, url.PathEscape(participanteID))
	req, err := h.prepareRequest(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error consultando participante en auth-service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service respondió HTTP %d", resp.StatusCode)
	}

	var res authProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("respuesta inválida de auth-service: %w", err)
	}

	return &models.ParticipanteInfo{
		ParticipanteID:  res.Data.ParticipantID,
		UserID:          res.Data.UserID,
		Nombre:          res.Data.Nombre,
		Apellido:        res.Data.Apellido,
		FechaNacimiento: res.Data.FechaNacimiento,
	}, nil
}

func (h *httpAuthDirectory) GetParticipanteByUserID(ctx context.Context, userID string) (*models.ParticipanteInfo, error) {
	targetURL := fmt.Sprintf("%s/api/v1/directory/participants/by-user/%s", h.baseURL, url.PathEscape(userID))
	req, err := h.prepareRequest(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error consultando participante en auth-service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service respondió HTTP %d", resp.StatusCode)
	}

	var res authProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("respuesta inválida de auth-service: %w", err)
	}

	return &models.ParticipanteInfo{
		ParticipanteID:  res.Data.ParticipantID,
		UserID:          res.Data.UserID,
		Nombre:          res.Data.Nombre,
		Apellido:        res.Data.Apellido,
		FechaNacimiento: res.Data.FechaNacimiento,
	}, nil
}

func (h *httpAuthDirectory) SearchParticipantes(ctx context.Context, query string, limit int) ([]models.ParticipanteInfo, error) {
	targetURL := fmt.Sprintf("%s/api/v1/directory/participants/search?q=%s&limit=%s", h.baseURL, url.QueryEscape(query), strconv.Itoa(limit))
	req, err := h.prepareRequest(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error buscando participantes en auth-service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusBadRequest {
		return nil, ErrInvalidSearch
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service respondió HTTP %d", resp.StatusCode)
	}

	var res authSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("respuesta inválida de auth-service: %w", err)
	}

	out := make([]models.ParticipanteInfo, 0, len(res.Data))
	for _, p := range res.Data {
		out = append(out, models.ParticipanteInfo{
			ParticipanteID:  p.ParticipantID,
			UserID:          p.UserID,
			Nombre:          p.Nombre,
			Apellido:        p.Apellido,
			FechaNacimiento: p.FechaNacimiento,
		})
	}
	return out, nil
}

func (h *httpAuthDirectory) GetNombresUsuarios(ctx context.Context, userIDs []string) (map[string]string, error) {
	out := make(map[string]string, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	targetURL := fmt.Sprintf("%s/api/v1/directory/users/summaries", h.baseURL)
	req, err := h.prepareRequest(ctx, http.MethodPost, targetURL, map[string][]string{"user_ids": userIDs})
	if err != nil {
		return nil, err
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error consultando usuarios en auth-service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service respondió HTTP %d", resp.StatusCode)
	}

	var res authSummariesResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("respuesta inválida de auth-service: %w", err)
	}

	for _, u := range res.Data {
		out[u.UserID] = u.FullName
	}
	return out, nil
}

func (h *httpAuthDirectory) Close() error {
	return nil
}

// NewAuthDirectory abre el cliente HTTP o gRPC hacia auth-service.
func NewAuthDirectory(target string, idTokenFn ...func(ctx context.Context) (string, error)) (AuthDirectory, error) {
	target = strings.TrimRight(strings.TrimSpace(target), "/")
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		var idToken func(ctx context.Context) (string, error)
		if len(idTokenFn) > 0 {
			idToken = idTokenFn[0]
		}
		return &httpAuthDirectory{
			baseURL:    target,
			httpClient: &http.Client{Timeout: authCallTimeout},
			idToken:    idToken,
		}, nil
	}

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
