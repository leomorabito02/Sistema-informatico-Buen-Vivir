package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/events"
	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type mockActivityRepo struct {
	activities []models.Activity
	err        error
}

func (m *mockActivityRepo) ListByCourse(ctx context.Context, cursoID string) ([]models.Activity, error) {
	return nil, nil
}
func (m *mockActivityRepo) ListByDateRange(ctx context.Context, start, end time.Time) ([]models.Activity, error) {
	return m.activities, m.err
}
func (m *mockActivityRepo) Get(ctx context.Context, id string) (*models.Activity, error) {
	return nil, nil
}
func (m *mockActivityRepo) Create(ctx context.Context, cursoID string, input models.ActivityInput) (*models.Activity, error) {
	return nil, nil
}
func (m *mockActivityRepo) Update(ctx context.Context, id string, input models.ActivityInput) (*models.Activity, error) {
	return nil, nil
}
func (m *mockActivityRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type mockEnrollmentRepo struct {
	enrollments map[string][]models.Enrollment
	err         error
}

func (m *mockEnrollmentRepo) ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.enrollments[cursoID], nil
}
func (m *mockEnrollmentRepo) ListByParticipant(ctx context.Context, participanteID string) ([]models.Enrollment, error) {
	return nil, nil
}
func (m *mockEnrollmentRepo) ListByParticipants(ctx context.Context, participantIDs []string) ([]models.Enrollment, error) {
	return nil, nil
}
func (m *mockEnrollmentRepo) Create(ctx context.Context, e models.Enrollment) (*models.Enrollment, error) {
	return nil, nil
}
func (m *mockEnrollmentRepo) DeactivateByParticipant(ctx context.Context, participanteID, motivo string) (int, error) {
	return 0, nil
}

func TestReminderService_TriggerReminders(t *testing.T) {
	now := time.Now().UTC()
	actDate := now.AddDate(0, 0, 1)

	acts := []models.Activity{
		{ID: "act-1", CursoID: "c1", Titulo: "Clase 1", Fecha: actDate},
	}
	enrollments := map[string][]models.Enrollment{
		"c1": {
			{ID: "e1", CursoID: "c1", ParticipanteID: "p1", Estado: models.EnrollmentEstadoActiva},
			{ID: "e2", CursoID: "c1", ParticipanteID: "p2", Estado: models.EnrollmentEstadoBaja},
		},
	}

	actRepo := &mockActivityRepo{activities: acts}
	enrRepo := &mockEnrollmentRepo{enrollments: enrollments}

	// producer without writer is disabled and won't crash
	prod := producers.NewEducacion(nil, "educacion-eventos", "contenido-eventos")
	svc := services.NewReminderService(actRepo, enrRepo, prod)

	dispatched, err := svc.TriggerReminders(context.Background(), actDate)
	require.NoError(t, err)
	assert.Equal(t, 1, dispatched)
}

func TestRemindersTriggerHandler(t *testing.T) {
	actRepo := &mockActivityRepo{activities: nil}
	enrRepo := &mockEnrollmentRepo{}
	prod := producers.NewEducacion(nil, "educacion-eventos", "contenido-eventos")
	svc := services.NewReminderService(actRepo, enrRepo, prod)

	h := services.RemindersTriggerHandler(svc, events.NoVerifier{})

	t.Run("Method not allowed on GET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/internal/reminders/trigger", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	})

	t.Run("POST triggers reminders successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/internal/reminders/trigger?date=2026-11-10", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"dispatched":0}`, rec.Body.String())
	})
}
