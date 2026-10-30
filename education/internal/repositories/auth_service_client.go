package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"buenvivir/education/internal/models"
)

// AuthServiceClient habla con auth-service para resolver datos que
// education no tiene (los participantes viven en la base de auth, no
// en Firestore — ver DER). Es el primer acoplamiento HTTP directo entre
// microservicios backend de este proyecto: se decidió así (en vez de
// que el frontend combine dos llamadas) para que education pueda armar
// en una sola respuesta el listado de participantes + su asistencia, y
// para poder validar en el backend que un participanteId existe y está
// aprobado antes de inscribirlo o registrarle asistencia.
//
// Reenvía el MISMO token del usuario que llamó a education (no usa un
// secreto de servicio propio) — así auth-service sigue siendo la única
// autoridad de RBAC sobre sus propios datos, con las mismas reglas que
// si el usuario le pegara directo (ver middlewares.TokenKey).
type AuthServiceClient struct {
	baseURL    string
	httpClient *http.Client
	idToken    func(ctx context.Context) (string, error)
}

// ErrAuthServiceNotConfigured se devuelve cuando AUTH_SERVICE_URL está
// vacío: el resto de education no depende de auth-service, así que el
// servicio igual arranca — esto solo se ve en el momento de inscribir,
// registrar asistencia, o pedir el listado de participantes.
var ErrAuthServiceNotConfigured = errors.New("AUTH_SERVICE_URL no está configurado")

func NewAuthServiceClient(baseURL string, idTokenFn ...func(ctx context.Context) (string, error)) *AuthServiceClient {
	var idToken func(ctx context.Context) (string, error)
	if len(idTokenFn) > 0 {
		idToken = idTokenFn[0]
	}
	return &AuthServiceClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		idToken:    idToken,
	}
}

// authListResponse replica el sobre {success, data: [...], errors} con
// el que responde auth-service para los listados (ver
// auth/internal/utils/response.go).
type authListResponse[T any] struct {
	Success bool `json:"success"`
	Data    []T  `json:"data"`
	Errors  any  `json:"errors"`
}

// authItemResponse es el mismo sobre, pero para un solo recurso
// (data es un objeto, no un array) — lo devuelven los GET /{id}.
type authItemResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
	Errors  any  `json:"errors"`
}

// ListParticipants pide GET /api/v1/participants a auth-service,
// reenviando bearerToken (el JWT crudo del usuario autenticado en
// education, sin el prefijo "Bearer ").
func (c *AuthServiceClient) ListParticipants(ctx context.Context, bearerToken string) ([]models.AuthParticipant, error) {
	var parsed authListResponse[models.AuthParticipant]
	if err := c.doGet(ctx, "/api/v1/participants", bearerToken, &parsed); err != nil {
		return nil, err
	}
	return parsed.Data, nil
}

// GetParticipant pide GET /api/v1/participants/{id}. Devuelve
// (nil, nil) si auth-service responde 404 (no existe) — no es un error
// de infraestructura, es el resultado esperado de una validación de
// negocio ("el participante no existe"), y así lo puede distinguir el
// caller (EnrollmentService/AttendanceService/GeneralAttendanceService)
// de una falla real de red o del propio auth-service.
func (c *AuthServiceClient) GetParticipant(ctx context.Context, participantID, bearerToken string) (*models.AuthParticipant, error) {
	if c == nil || c.baseURL == "" {
		return nil, ErrAuthServiceNotConfigured
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/participants/"+participantID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	if c.idToken != nil {
		if tok, err := c.idToken(ctx); err == nil && tok != "" {
			req.Header.Set("X-Serverless-Authorization", "Bearer "+tok)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de red hablando con auth-service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("auth-service respondió %d al buscar el participante", resp.StatusCode)
	}

	var parsed authItemResponse[models.AuthParticipant]
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("no se pudo parsear la respuesta de auth-service: %w", err)
	}
	return &parsed.Data, nil
}

// ListUsers pide GET /api/v1/users a auth-service. Participante no
// tiene nombre propio (vive en Usuario, ver DER) — esto es lo que usa
// GeneralAttendanceService para resolver nombre/apellido a partir del
// user_id de cada participante. A diferencia de /participants, este
// endpoint no exige un rol puntual en auth-service (cualquier usuario
// autenticado puede listarlo).
func (c *AuthServiceClient) ListUsers(ctx context.Context, bearerToken string) ([]models.AuthUser, error) {
	var parsed authListResponse[models.AuthUser]
	if err := c.doGet(ctx, "/api/v1/users", bearerToken, &parsed); err != nil {
		return nil, err
	}
	return parsed.Data, nil
}

// GetParticipantName pide GET /api/v1/directory/participants/{id}: el
// nombre y apellido de UN participante, sin DNI ni datos de contacto.
// Es lo único del directorio de auth-service que puede consultar un
// docente (no puede listar /participants), y lo usa para ver los
// nombres de sus alumnos en las entregas.
func (c *AuthServiceClient) GetParticipantName(ctx context.Context, participantID, bearerToken string) (models.AuthUser, error) {
	var parsed authItemResponse[struct {
		Nombre   string `json:"nombre"`
		Apellido string `json:"apellido"`
	}]
	if err := c.doGet(ctx, "/api/v1/directory/participants/"+url.PathEscape(participantID), bearerToken, &parsed); err != nil {
		return models.AuthUser{}, err
	}
	return models.AuthUser{FirstName: parsed.Data.Nombre, LastName: parsed.Data.Apellido}, nil
}

// GetMyParticipantID pide GET /api/v1/profile/me a auth-service con el
// token del propio participante y devuelve su participant_id (el id que
// education usa como participanteId en inscripciones, asistencias y
// entregas — el "sub" del JWT es el user_id, no sirve). Devuelve "" si
// el usuario no tiene perfil de participante.
func (c *AuthServiceClient) GetMyParticipantID(ctx context.Context, bearerToken string) (string, error) {
	var parsed struct {
		RoleData struct {
			ParticipantID string `json:"participant_id"`
		} `json:"role_data"`
	}
	if err := c.doGet(ctx, "/api/v1/profile/me", bearerToken, &parsed); err != nil {
		return "", err
	}
	return parsed.RoleData.ParticipantID, nil
}

func (c *AuthServiceClient) doGet(ctx context.Context, path, bearerToken string, out any) error {
	if c == nil || c.baseURL == "" {
		return ErrAuthServiceNotConfigured
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	if c.idToken != nil {
		if tok, err := c.idToken(ctx); err == nil && tok != "" {
			req.Header.Set("X-Serverless-Authorization", "Bearer "+tok)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de red hablando con auth-service (%s): %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("auth-service respondió %d en %s", resp.StatusCode, path)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("no se pudo parsear la respuesta de auth-service (%s): %w", path, err)
	}
	return nil
}
