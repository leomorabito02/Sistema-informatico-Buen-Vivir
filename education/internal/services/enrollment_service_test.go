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

type MockEnrollmentRepository struct {
	mock.Mock
}

func (m *MockEnrollmentRepository) ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error) {
	args := m.Called(ctx, cursoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Enrollment), args.Error(1)
}

func TestEnrollmentService_ListByCourse_Success(t *testing.T) {
	repo := new(MockEnrollmentRepository)
	svc := services.NewEnrollmentService(repo)

	repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{{ID: "e1"}}, nil)

	enrollments, err := svc.ListByCourse(context.Background(), "c1")
	assert.NoError(t, err)
	assert.Len(t, enrollments, 1)
}

func TestEnrollmentService_ListByCourse_RepoError(t *testing.T) {
	repo := new(MockEnrollmentRepository)
	svc := services.NewEnrollmentService(repo)

	repo.On("ListByCourse", mock.Anything, "c1").Return(nil, errors.New("firestore caído"))

	enrollments, err := svc.ListByCourse(context.Background(), "c1")
	assert.Error(t, err)
	assert.Nil(t, enrollments)
}
