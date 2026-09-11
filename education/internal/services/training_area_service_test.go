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

type MockTrainingAreaRepository struct {
	mock.Mock
}

func (m *MockTrainingAreaRepository) List(ctx context.Context) ([]models.TrainingArea, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TrainingArea), args.Error(1)
}

func (m *MockTrainingAreaRepository) Create(ctx context.Context, input models.TrainingAreaInput) (*models.TrainingArea, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TrainingArea), args.Error(1)
}

func (m *MockTrainingAreaRepository) Update(ctx context.Context, id string, input models.TrainingAreaInput) (*models.TrainingArea, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TrainingArea), args.Error(1)
}

func (m *MockTrainingAreaRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTrainingAreaService_ListTrainingAreas(t *testing.T) {
	repo := new(MockTrainingAreaRepository)
	svc := services.NewTrainingAreaService(repo)

	repo.On("List", mock.Anything).Return([]models.TrainingArea{{ID: "ta1"}}, nil)

	areas, err := svc.ListTrainingAreas(context.Background())
	assert.NoError(t, err)
	assert.Len(t, areas, 1)
}

func TestTrainingAreaService_ListTrainingAreas_RepoError(t *testing.T) {
	repo := new(MockTrainingAreaRepository)
	svc := services.NewTrainingAreaService(repo)

	repo.On("List", mock.Anything).Return(nil, errors.New("contentful caído"))

	areas, err := svc.ListTrainingAreas(context.Background())
	assert.Error(t, err)
	assert.Nil(t, areas)
}

func TestTrainingAreaService_CreateTrainingArea_MissingName(t *testing.T) {
	svc := services.NewTrainingAreaService(nil)

	area, err := svc.CreateTrainingArea(context.Background(), models.TrainingAreaInput{Name: "  "})
	assert.Error(t, err)
	assert.Nil(t, area)
	assert.True(t, services.IsValidationError(err))
}

func TestTrainingAreaService_CreateTrainingArea_Success(t *testing.T) {
	repo := new(MockTrainingAreaRepository)
	svc := services.NewTrainingAreaService(repo)

	input := models.TrainingAreaInput{Name: "Oficios"}
	repo.On("Create", mock.Anything, input).Return(&models.TrainingArea{ID: "ta1", Name: "Oficios"}, nil)

	area, err := svc.CreateTrainingArea(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, "ta1", area.ID)
}

func TestTrainingAreaService_UpdateTrainingArea_EmptyID(t *testing.T) {
	svc := services.NewTrainingAreaService(nil)

	area, err := svc.UpdateTrainingArea(context.Background(), "", models.TrainingAreaInput{Name: "Oficios"})
	assert.Error(t, err)
	assert.Nil(t, area)
}

func TestTrainingAreaService_UpdateTrainingArea_MissingName(t *testing.T) {
	svc := services.NewTrainingAreaService(nil)

	area, err := svc.UpdateTrainingArea(context.Background(), "ta1", models.TrainingAreaInput{Name: ""})
	assert.Error(t, err)
	assert.Nil(t, area)
}

func TestTrainingAreaService_UpdateTrainingArea_Success(t *testing.T) {
	repo := new(MockTrainingAreaRepository)
	svc := services.NewTrainingAreaService(repo)

	input := models.TrainingAreaInput{Name: "Oficios 2"}
	repo.On("Update", mock.Anything, "ta1", input).Return(&models.TrainingArea{ID: "ta1", Name: "Oficios 2"}, nil)

	area, err := svc.UpdateTrainingArea(context.Background(), "ta1", input)
	assert.NoError(t, err)
	assert.Equal(t, "Oficios 2", area.Name)
}

func TestTrainingAreaService_DeleteTrainingArea_EmptyID(t *testing.T) {
	svc := services.NewTrainingAreaService(nil)
	err := svc.DeleteTrainingArea(context.Background(), "")
	assert.Error(t, err)
	assert.True(t, services.IsValidationError(err))
}

func TestTrainingAreaService_DeleteTrainingArea_Success(t *testing.T) {
	repo := new(MockTrainingAreaRepository)
	svc := services.NewTrainingAreaService(repo)

	repo.On("Delete", mock.Anything, "ta1").Return(nil)

	err := svc.DeleteTrainingArea(context.Background(), "ta1")
	assert.NoError(t, err)
}
