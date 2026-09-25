package services_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAttendanceRepository struct {
	mock.Mock
}

func (m *MockAttendanceRepository) ListByActivity(ctx context.Context, actividadID string) ([]models.Attendance, error) {
	args := m.Called(ctx, actividadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Attendance), args.Error(1)
}

func (m *MockAttendanceRepository) ListByParticipant(ctx context.Context, participanteID string) ([]models.Attendance, error) {
	args := m.Called(ctx, participanteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Attendance), args.Error(1)
}

func (m *MockAttendanceRepository) RegisterBatch(ctx context.Context, actividadID string, inputs []models.AttendanceInput) error {
	args := m.Called(ctx, actividadID, inputs)
	return args.Error(0)
}

func TestAttendanceService_ListByActivity_EmptyID(t *testing.T) {
	svc := services.NewAttendanceService(nil, nil, nil)

	attendances, err := svc.ListByActivity(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, attendances)
	assert.True(t, services.IsValidationError(err))
}

func TestAttendanceService_ListByActivity_Success(t *testing.T) {
	repo := new(MockAttendanceRepository)
	svc := services.NewAttendanceService(repo, nil, nil)

	repo.On("ListByActivity", mock.Anything, "a1").Return([]models.Attendance{{ID: "att1"}}, nil)

	attendances, err := svc.ListByActivity(context.Background(), "a1")
	assert.NoError(t, err)
	assert.Len(t, attendances, 1)
}

func TestAttendanceService_ListByActivity_RepoError(t *testing.T) {
	repo := new(MockAttendanceRepository)
	svc := services.NewAttendanceService(repo, nil, nil)

	repo.On("ListByActivity", mock.Anything, "a1").Return(nil, errors.New("firestore caído"))

	attendances, err := svc.ListByActivity(context.Background(), "a1")
	assert.Error(t, err)
	assert.Nil(t, attendances)
}

func TestAttendanceService_RegisterBatch_EmptyActividadID(t *testing.T) {
	svc := services.NewAttendanceService(nil, nil, nil)

	err := svc.RegisterBatch(context.Background(), "", "c1", []models.AttendanceInput{{ParticipanteID: "p1"}}, "token")
	assert.Error(t, err)
}

func TestAttendanceService_RegisterBatch_EmptyInputs(t *testing.T) {
	svc := services.NewAttendanceService(nil, nil, nil)

	err := svc.RegisterBatch(context.Background(), "a1", "c1", nil, "token")
	assert.Error(t, err)
}

func TestAttendanceService_RegisterBatch_MissingParticipanteID(t *testing.T) {
	svc := services.NewAttendanceService(nil, nil, nil)

	inputs := []models.AttendanceInput{
		{ParticipanteID: "p1", Presente: true},
		{ParticipanteID: "", Presente: false},
	}
	err := svc.RegisterBatch(context.Background(), "a1", "c1", inputs, "token")
	assert.Error(t, err)
}

func TestAttendanceService_RegisterBatch_Success(t *testing.T) {
	// Create a dummy auth service
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Simulate authItemResponse[Participant]
		w.Write([]byte(`{"success":true,"data":{"user_id":"` + "p1" + `","approval_status":"approved"}}`))
	}))
	defer ts.Close()
	
	authClient := repositories.NewAuthServiceClient(ts.URL)

	repo := new(MockAttendanceRepository)
	enrollmentRepo := new(MockEnrollmentRepository)
	svc := services.NewAttendanceService(repo, enrollmentRepo, authClient)

	inputs := []models.AttendanceInput{
		{ParticipanteID: "p1", Presente: true},
	}
	
	enrollmentRepo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{
		{ParticipanteID: "p1", Estado: models.EnrollmentEstadoActiva},
	}, nil)
	repo.On("RegisterBatch", mock.Anything, "a1", inputs).Return(nil)

	err := svc.RegisterBatch(context.Background(), "a1", "c1", inputs, "token")
	assert.NoError(t, err)
}
