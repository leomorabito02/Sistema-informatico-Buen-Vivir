package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auth/internal/controllers"
	"auth/internal/repositories"
	"auth/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type fakeDirectoryService struct {
	err      error
	gotLimit int
}

func (f *fakeDirectoryService) GetProfileByParticipantID(_ context.Context, id string) (*repositories.ParticipantProfileRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &repositories.ParticipantProfileRow{
		ParticipantID: id, UserID: "usr-1", FirstName: "Ana", LastName: "Pérez",
		Birthdate: time.Date(2000, 5, 3, 0, 0, 0, 0, time.UTC), ApprovalStatus: "approved",
	}, nil
}

func (f *fakeDirectoryService) GetProfileByUserID(_ context.Context, id string) (*repositories.ParticipantProfileRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &repositories.ParticipantProfileRow{ParticipantID: "part-1", UserID: id}, nil
}

func (f *fakeDirectoryService) SearchParticipants(_ context.Context, _ string, limit int) ([]repositories.ParticipantProfileRow, error) {
	f.gotLimit = limit
	if f.err != nil {
		return nil, f.err
	}
	return []repositories.ParticipantProfileRow{{ParticipantID: "part-1"}, {ParticipantID: "part-2"}}, nil
}

func (f *fakeDirectoryService) GetUserSummaries(_ context.Context, _ []string) ([]repositories.UserSummaryRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []repositories.UserSummaryRow{
		{UserID: "usr-1", FirstName: "Ana", LastName: "Pérez"},
		{UserID: "usr-2"},
	}, nil
}

func directoryRouter(svc services.ParticipantDirectoryService) http.Handler {
	c := controllers.NewDirectoryController(svc)
	r := chi.NewRouter()
	r.Get("/participants/{id}", c.GetParticipantProfile)
	r.Get("/users/{userId}", c.GetProfileByUserID)
	r.Get("/search", c.SearchParticipants)
	r.Post("/summaries", c.GetUserSummaries)
	return r
}

func doDirectory(h http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestDirectoryController_GetParticipantProfile(t *testing.T) {
	rec := doDirectory(directoryRouter(&fakeDirectoryService{}), http.MethodGet, "/participants/part-1", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Data controllers.ParticipantProfileDTO `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "part-1", resp.Data.ParticipantID)
	assert.Equal(t, "Ana", resp.Data.Nombre)
	assert.Equal(t, "2000-05-03", resp.Data.FechaNacimiento)

	rec = doDirectory(directoryRouter(&fakeDirectoryService{err: repositories.ErrParticipantProfileNotFound}), http.MethodGet, "/participants/part-9", nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	rec = doDirectory(directoryRouter(&fakeDirectoryService{err: errors.New("db")}), http.MethodGet, "/participants/part-9", nil)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	rec = doDirectory(directoryRouter(&fakeDirectoryService{}), http.MethodGet, "/participants/%20", nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDirectoryController_GetProfileByUserID(t *testing.T) {
	rec := doDirectory(directoryRouter(&fakeDirectoryService{}), http.MethodGet, "/users/usr-1", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, rec.Body.String(), "fecha_nacimiento")

	rec = doDirectory(directoryRouter(&fakeDirectoryService{err: repositories.ErrParticipantProfileNotFound}), http.MethodGet, "/users/usr-9", nil)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	rec = doDirectory(directoryRouter(&fakeDirectoryService{err: errors.New("db")}), http.MethodGet, "/users/usr-9", nil)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	rec = doDirectory(directoryRouter(&fakeDirectoryService{}), http.MethodGet, "/users/%20", nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDirectoryController_SearchParticipants(t *testing.T) {
	svc := &fakeDirectoryService{}
	rec := doDirectory(directoryRouter(svc), http.MethodGet, "/search?q=ana&limit=5", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 5, svc.gotLimit)
	assert.Contains(t, rec.Body.String(), "part-2")

	// Límite inválido: queda el valor por defecto.
	doDirectory(directoryRouter(svc), http.MethodGet, "/search?q=ana&limit=abc", nil)
	assert.Equal(t, 20, svc.gotLimit)

	rec = doDirectory(directoryRouter(&fakeDirectoryService{err: services.ErrInvalidDirectoryRequest}), http.MethodGet, "/search?q=a", nil)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	rec = doDirectory(directoryRouter(&fakeDirectoryService{err: errors.New("db")}), http.MethodGet, "/search?q=ana", nil)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDirectoryController_GetUserSummaries(t *testing.T) {
	body := []byte(`{"user_ids":["usr-1","usr-2"]}`)
	rec := doDirectory(directoryRouter(&fakeDirectoryService{}), http.MethodPost, "/summaries", body)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Data []controllers.UserSummaryDTO `json:"data"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "Ana Pérez", resp.Data[0].FullName)

	rec = doDirectory(directoryRouter(&fakeDirectoryService{}), http.MethodPost, "/summaries", []byte("{"))
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	rec = doDirectory(directoryRouter(&fakeDirectoryService{err: services.ErrInvalidDirectoryRequest}), http.MethodPost, "/summaries", body)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	rec = doDirectory(directoryRouter(&fakeDirectoryService{err: errors.New("db")}), http.MethodPost, "/summaries", body)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
