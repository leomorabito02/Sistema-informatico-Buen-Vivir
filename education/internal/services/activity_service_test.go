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

func (m *MockActivityRepository) Update(ctx context.Context, id string, input models.ActivityInput) (*models.Activity, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Activity), args.Error(1)
}

func (m *MockActivityRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestActivityService_ListByCourse(t *testing.T) {
	repo := new(MockActivityRepository)
	svc := services.NewActivityService(repo, nil)

	repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{{ID: "a1"}}, nil)

	activities, err := svc.ListByCourse(context.Background(), "c1")
	assert.NoError(t, err)
	assert.Len(t, activities, 1)
}

func TestActivityService_ListByCourse_RepoError(t *testing.T) {
	repo := new(MockActivityRepository)
	svc := services.NewActivityService(repo, nil)

	repo.On("ListByCourse", mock.Anything, "c1").Return(nil, errors.New("firestore caído"))

	activities, err := svc.ListByCourse(context.Background(), "c1")
	assert.Error(t, err)
	assert.Nil(t, activities)
}

func TestActivityService_GetActivity_EmptyID(t *testing.T) {
	svc := services.NewActivityService(nil, nil)

	activity, err := svc.GetActivity(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, activity)
	assert.True(t, services.IsValidationError(err))
}

func TestActivityService_GetActivity_Success(t *testing.T) {
	repo := new(MockActivityRepository)
	svc := services.NewActivityService(repo, nil)

	repo.On("Get", mock.Anything, "a1").Return(&models.Activity{ID: "a1"}, nil)

	activity, err := svc.GetActivity(context.Background(), "a1")
	assert.NoError(t, err)
	assert.Equal(t, "a1", activity.ID)
}

func TestActivityService_CreateActivity_MissingTitulo(t *testing.T) {
	svc := services.NewActivityService(nil, nil)

	activity, err := svc.CreateActivity(context.Background(), "c1", models.ActivityInput{
		Titulo: "", Consigna: "Traer una receta", Fecha: time.Now(),
	})
	assert.Error(t, err)
	assert.Nil(t, activity)
}

func TestActivityService_CreateActivity_MissingFecha(t *testing.T) {
	svc := services.NewActivityService(nil, nil)

	activity, err := svc.CreateActivity(context.Background(), "c1", models.ActivityInput{
		Titulo: "Práctica 1", Consigna: "Traer una receta",
	})
	assert.Error(t, err)
	assert.Nil(t, activity)
}

func TestActivityService_CreateActivity_Success(t *testing.T) {
	repo := new(MockActivityRepository)
	svc := services.NewActivityService(repo, nil)

	// La fecha llega con hora y huso: el service la guarda como
	// medianoche UTC de ese día, y recorta espacios de titulo/consigna.
	input := models.ActivityInput{
		Titulo:   "  Práctica 1 ",
		Consigna: "Traer una receta",
		Fecha:    time.Date(2026, 10, 5, 15, 30, 0, 0, time.UTC),
	}
	expected := models.ActivityInput{
		Titulo:   "Práctica 1",
		Consigna: "Traer una receta",
		Fecha:    time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
	}
	repo.On("Create", mock.Anything, "c1", expected).Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)

	activity, err := svc.CreateActivity(context.Background(), "c1", input)
	assert.NoError(t, err)
	assert.Equal(t, "a1", activity.ID)
}

func TestActivityService_CreateActivity_MissingConsigna(t *testing.T) {
	svc := services.NewActivityService(nil, nil)

	activity, err := svc.CreateActivity(context.Background(), "c1", models.ActivityInput{
		Titulo: "Práctica 1", Consigna: "   ", Fecha: time.Now(),
	})
	assert.True(t, services.IsValidationError(err))
	assert.Nil(t, activity)
}

func TestActivityService_ListByCourse_SortsByFecha(t *testing.T) {
	repo := new(MockActivityRepository)
	svc := services.NewActivityService(repo, nil)

	d := func(day int) time.Time { return time.Date(2026, 10, day, 0, 0, 0, 0, time.UTC) }
	repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{
		{ID: "tarde", Fecha: d(9)},
		{ID: "temprano", Fecha: d(2)},
		{ID: "medio", Fecha: d(5)},
	}, nil)

	activities, err := svc.ListByCourse(context.Background(), "c1")
	assert.NoError(t, err)
	assert.Equal(t, "temprano", activities[0].ID)
	assert.Equal(t, "medio", activities[1].ID)
	assert.Equal(t, "tarde", activities[2].ID)
}

func TestActivityService_UpdateActivity_Success(t *testing.T) {
	repo := new(MockActivityRepository)
	svc := services.NewActivityService(repo, nil)

	input := models.ActivityInput{Titulo: "Práctica 2", Consigna: "Hornear", Fecha: time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)}
	repo.On("Update", mock.Anything, "a1", input).Return(&models.Activity{ID: "a1", Titulo: "Práctica 2"}, nil)

	activity, err := svc.UpdateActivity(context.Background(), "a1", input)
	assert.NoError(t, err)
	assert.Equal(t, "Práctica 2", activity.Titulo)
}

func TestActivityService_UpdateActivity_Validation(t *testing.T) {
	svc := services.NewActivityService(nil, nil)

	_, err := svc.UpdateActivity(context.Background(), "", models.ActivityInput{})
	assert.True(t, services.IsValidationError(err))

	_, err = svc.UpdateActivity(context.Background(), "a1", models.ActivityInput{Titulo: "x"})
	assert.True(t, services.IsValidationError(err))
}

func TestActivityService_DeleteActivity_CleansUpBucket(t *testing.T) {
	repo := new(MockActivityRepository)
	storage := new(MockFileStorage)
	svc := services.NewActivityService(repo, storage)

	repo.On("Delete", mock.Anything, "a1").Return(nil)
	storage.On("DeletePrefix", mock.Anything, "entregas/a1/").Return(nil)

	assert.NoError(t, svc.DeleteActivity(context.Background(), "a1"))
	storage.AssertExpectations(t)
}

func TestActivityService_DeleteActivity_BucketErrorIsNotFatal(t *testing.T) {
	repo := new(MockActivityRepository)
	storage := new(MockFileStorage)
	svc := services.NewActivityService(repo, storage)

	repo.On("Delete", mock.Anything, "a1").Return(nil)
	storage.On("DeletePrefix", mock.Anything, "entregas/a1/").Return(errors.New("bucket caído"))

	assert.NoError(t, svc.DeleteActivity(context.Background(), "a1"))
}

func TestActivityService_DeleteActivity_RepoError(t *testing.T) {
	repo := new(MockActivityRepository)
	storage := new(MockFileStorage)
	svc := services.NewActivityService(repo, storage)

	repo.On("Delete", mock.Anything, "a1").Return(errors.New("firestore caído"))

	assert.Error(t, svc.DeleteActivity(context.Background(), "a1"))
	storage.AssertNotCalled(t, "DeletePrefix", mock.Anything, mock.Anything)
}

func TestActivityService_DeleteActivity_EmptyID(t *testing.T) {
	svc := services.NewActivityService(nil, nil)
	assert.True(t, services.IsValidationError(svc.DeleteActivity(context.Background(), "")))
}
