// Package clients consulta a los servicios de origen en nombre del
// administrador que pidió el dashboard.
package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Credentials es la identidad de quien pidió el dashboard. Autenticación
// y Educación validan el JWT por su cuenta (cookie o Authorization);
// Inserción Laboral confía en X-User-ID / X-User-Role, que pone el API
// Gateway después de validar el token.
type Credentials struct {
	Cookie        string
	Authorization string
	UserID        string
	UserRole      string
}

// FromRequest toma las credenciales del pedido que llegó del gateway.
func FromRequest(r *http.Request) Credentials {
	return Credentials{
		Cookie:        r.Header.Get("Cookie"),
		Authorization: r.Header.Get("Authorization"),
		UserID:        r.Header.Get("X-User-ID"),
		UserRole:      r.Header.Get("X-User-Role"),
	}
}

// UpstreamError: un servicio de origen respondió mal o no respondió.
type UpstreamError struct {
	Service string
	Status  int // 0 si no hubo respuesta
	Err     error
}

func (e *UpstreamError) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("%s respondió %d", e.Service, e.Status)
	}
	return fmt.Sprintf("%s no respondió: %v", e.Service, e.Err)
}

func (e *UpstreamError) Unwrap() error { return e.Err }

// ErrNoConfigurado: no hay dirección para ese servicio. Solo falla ese indicador.
var ErrNoConfigurado = errors.New("servicio no configurado")

// Upstream es un servicio de origen.
type Upstream struct {
	Name    string
	BaseURL string
	HTTP    *http.Client
	// IDToken da un token de identidad de Google para llamar a servicios privados de Cloud
	// Run (va en X-Serverless-Authorization; la sesión sigue en Cookie y Authorization).
	IDToken func(ctx context.Context) (string, error)
}

func NewUpstream(name, baseURL string, timeout time.Duration) *Upstream {
	return &Upstream{Name: name, BaseURL: baseURL, HTTP: &http.Client{Timeout: timeout}}
}

// GetJSON hace GET a path y decodifica la respuesta JSON en out.
func (u *Upstream) GetJSON(ctx context.Context, path string, cred Credentials, out any) error {
	if u.BaseURL == "" {
		return &UpstreamError{Service: u.Name, Err: ErrNoConfigurado}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.BaseURL+path, nil)
	if err != nil {
		return &UpstreamError{Service: u.Name, Err: err}
	}
	if u.IDToken != nil {
		tok, err := u.IDToken(ctx)
		if err != nil {
			return &UpstreamError{Service: u.Name, Err: fmt.Errorf("token de identidad: %w", err)}
		}
		req.Header.Set("X-Serverless-Authorization", "Bearer "+tok)
	}
	req.Header.Set("Accept", "application/json")
	if cred.Cookie != "" {
		req.Header.Set("Cookie", cred.Cookie)
	}
	if cred.Authorization != "" {
		req.Header.Set("Authorization", cred.Authorization)
	}
	if cred.UserID != "" {
		req.Header.Set("X-User-ID", cred.UserID)
	}
	if cred.UserRole != "" {
		req.Header.Set("X-User-Role", cred.UserRole)
	}

	res, err := u.HTTP.Do(req)
	if err != nil {
		return &UpstreamError{Service: u.Name, Err: err}
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return &UpstreamError{Service: u.Name, Status: res.StatusCode}
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return &UpstreamError{Service: u.Name, Status: res.StatusCode, Err: fmt.Errorf("respuesta inválida: %w", err)}
	}
	return nil
}
