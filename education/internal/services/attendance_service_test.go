package services_test

import (
	"context"
	"errors"
	"testing"

	"buenvivir/education/internal/models"
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

func TestAttendanceService_ListByActivity_EmptyID(t *testing.T) {
	svc := services.NewAttendanceService(nil)

	attendances, err := svc.ListByActivity(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, attendances)
	assert.True(t, services.IsValidationError(err))
}

func TestAttendanceService_ListByActivity_Success(t *testing.T) {
	repo := new(MockAttendanceRepository)
	svc := services.NewAttendanceService(repo)

	repo.On("ListByActivity", mock.Anything, "a1").Return([]models.Attendance{{ID: "att1"}}, nil)

	attendances, err := svc.ListByActivity(context.Background(), "a1")
	assert.NoError(t, err)
	assert.Len(t, attendances, 1)
}

func TestAttendanceService_ListByActivity_RepoError(t *testing.T) {
	repo := new(MockAttendanceRepository)
	svc := services.NewAttendanceService(repo)

	repo.On("ListByActivity", mock.Anything, "a1").Return(nil, errors.New("firestore caído"))

	attendances, err := svc.ListByActivity(context.Background(), "a1")
	assert.Error(t, err)
	assert.Nil(t, attendances)
}
