package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"auth/internal/repositories"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
)

// fakeDirectoryRepo guarda lo que recibe para verificar la limpieza que
// hace el servicio antes de llegar a la base.
type fakeDirectoryRepo struct {
	gotQuery string
	gotLimit int
	gotIDs   []string
	err      error
}

func (f *fakeDirectoryRepo) FindProfileByParticipantID(_ context.Context, id string) (*repositories.ParticipantProfileRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &repositories.ParticipantProfileRow{ParticipantID: id}, nil
}

func (f *fakeDirectoryRepo) FindProfileByUserID(_ context.Context, id string) (*repositories.ParticipantProfileRow, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &repositories.ParticipantProfileRow{UserID: id}, nil
}

func (f *fakeDirectoryRepo) SearchProfiles(_ context.Context, q string, limit int) ([]repositories.ParticipantProfileRow, error) {
	f.gotQuery, f.gotLimit = q, limit
	return []repositories.ParticipantProfileRow{{ParticipantID: "part-1"}}, f.err
}

func (f *fakeDirectoryRepo) FindUserSummaries(_ context.Context, ids []string) ([]repositories.UserSummaryRow, error) {
	f.gotIDs = ids
	return []repositories.UserSummaryRow{{UserID: "usr-1", FirstName: "Ana"}}, f.err
}

func TestDirectoryService_ProfileByIDs(t *testing.T) {
	repo := &fakeDirectoryRepo{}
	svc := services.NewParticipantDirectoryService(repo)
	ctx := context.Background()

	p, err := svc.GetProfileByParticipantID(ctx, "part-1")
	assert.NoError(t, err)
	assert.Equal(t, "part-1", p.ParticipantID)

	u, err := svc.GetProfileByUserID(ctx, "usr-1")
	assert.NoError(t, err)
	assert.Equal(t, "usr-1", u.UserID)

	_, err = svc.GetProfileByParticipantID(ctx, "  ")
	assert.ErrorIs(t, err, services.ErrInvalidDirectoryRequest)
	_, err = svc.GetProfileByUserID(ctx, "")
	assert.ErrorIs(t, err, services.ErrInvalidDirectoryRequest)
}

func TestDirectoryService_SearchParticipants(t *testing.T) {
	repo := &fakeDirectoryRepo{}
	svc := services.NewParticipantDirectoryService(repo)
	ctx := context.Background()

	_, err := svc.SearchParticipants(ctx, "  ana   perez ", 10)
	assert.NoError(t, err)
	assert.Equal(t, "ana perez", repo.gotQuery)
	assert.Equal(t, 10, repo.gotLimit)

	// Límite fuera de rango: usa el valor por defecto.
	_, err = svc.SearchParticipants(ctx, "ana", 500)
	assert.NoError(t, err)
	assert.Equal(t, 20, repo.gotLimit)

	_, err = svc.SearchParticipants(ctx, "a", 10)
	assert.ErrorIs(t, err, services.ErrInvalidDirectoryRequest)
	_, err = svc.SearchParticipants(ctx, strings.Repeat("a", 101), 10)
	assert.ErrorIs(t, err, services.ErrInvalidDirectoryRequest)
}

// Sin texto lista a todos los participantes, con un tope propio más alto que el de la búsqueda.
func TestDirectoryService_SearchParticipants_SinTextoListaTodos(t *testing.T) {
	repo := &fakeDirectoryRepo{}
	svc := services.NewParticipantDirectoryService(repo)
	ctx := context.Background()

	_, err := svc.SearchParticipants(ctx, "   ", 0)
	assert.NoError(t, err)
	assert.Equal(t, "", repo.gotQuery)
	assert.Equal(t, 500, repo.gotLimit)

	_, err = svc.SearchParticipants(ctx, "", 200)
	assert.NoError(t, err)
	assert.Equal(t, 200, repo.gotLimit)

	_, err = svc.SearchParticipants(ctx, "", 9999)
	assert.NoError(t, err)
	assert.Equal(t, 500, repo.gotLimit)
}

func TestDirectoryService_GetUserSummaries(t *testing.T) {
	repo := &fakeDirectoryRepo{}
	svc := services.NewParticipantDirectoryService(repo)
	ctx := context.Background()

	_, err := svc.GetUserSummaries(ctx, []string{"usr-1", " usr-1 ", "", "usr-2"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"usr-1", "usr-2"}, repo.gotIDs)

	many := make([]string, 201)
	for i := range many {
		many[i] = "usr-" + strings.Repeat("x", i+1)
	}
	_, err = svc.GetUserSummaries(ctx, many)
	assert.ErrorIs(t, err, services.ErrInvalidDirectoryRequest)

	repo.err = errors.New("db caída")
	_, err = svc.GetUserSummaries(ctx, []string{"usr-1"})
	assert.Error(t, err)
}
