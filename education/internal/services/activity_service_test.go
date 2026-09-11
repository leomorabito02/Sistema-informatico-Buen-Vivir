package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockActivityRepository struct {
	mock.Mock
}

func (m *MockActivityRepository) ListByCourse(ctx context.Context, cursoID string) ([]models.Activity, error) {
	args := m.Called(ctx, cursoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Activity), args.Error(1)
}

func (m *MockActivityRepository) Get(ctx context.Context, id string) (*models.Activity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Activity), args.Error(1)
}

func (m *MockActivityRepository) Create(ctx context.Context, cursoID string, input models.ActivityInput) (*models.Activity, error) {
	args := m.Called(ctx, cursoID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Activity), args.Error(1)
}

func TestActivityService_ListByCourse(t *testing.T) {
	repo := new(MockActivityRepository)
	svc := services.NewActivityService(repo)

	repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{{ID: "a1"}}, nil)

	activities, err := svc.ListByCourse(context.Background(), "c1")
	assert.NoError(t, err)
	assert.Len(t, activities, 1)
}

func TestActivityService_ListByCourse_RepoError(t *testing.T) {
	repo := new(MockActivityRepository)
	svc := services.NewActivityService(repo)

	repo.On("ListByCourse", mock.Anything, "c1").Return(nil, errors.New("firestore caído"))

	activities, err := svc.ListByCourse(context.Background(), "c1")
	assert.Error(t, err)
	assert.Nil(t, activities)
}

func TestActivityService_GetActivity_EmptyID(t *testing.T) {
	svc := services.NewActivityService(nil)

	activity, err := svc.GetActivity(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, activity)
	assert.True(t, services.IsValidationError(err))
}

func TestActivityService_GetActivity_Success(t *testing.T) {
	repo := new(MockActivityRepository)
	svc := services.NewActivityService(repo)

	repo.On("Get", mock.Anything, "a1").Return(&models.Activity{ID: "a1"}, nil)

	activity, err := svc.GetActivity(context.Background(), "a1")
	assert.NoError(t, err)
	assert.Equal(t, "a1", activity.ID)
}

func TestActivityService_CreateActivity_MissingTitulo(t *testing.T) {
	svc := services.NewActivityService(nil)

	activity, err := svc.CreateActivity(context.Background(), "c1", models.ActivityInput{
		Titulo: "", Fecha: time.Now(),
	})
	assert.Error(t, err)
	assert.Nil(t, activity)
}

func TestActivityService_CreateActivity_MissingFecha(t *testing.T) {
	svc := services.NewActivityService(nil)

	activity, err := svc.CreateActivity(context.Background(), "c1", models.ActivityInput{
		Titulo: "Práctica 1",
	})
	assert.Error(t, err)
	assert.Nil(t, activity)
}

func TestActivityService_CreateActivity_Success(t *testing.T) {
	repo := new(MockActivityRepository)
	svc := services.NewActivityService(repo)

	input := models.ActivityInput{Titulo: "Práctica 1", Fecha: time.Now()}
	repo.On("Create", mock.Anything, "c1", input).Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)

	activity, err := svc.CreateActivity(context.Background(), "c1", input)
	assert.NoError(t, err)
	assert.Equal(t, "a1", activity.ID)
}
