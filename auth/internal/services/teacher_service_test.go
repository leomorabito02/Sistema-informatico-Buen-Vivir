package services_test

import (
	"context"
	"errors"
	"testing"

	"auth/internal/middlewares"

	"auth/internal/dto"
	"auth/internal/models"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTeacherRepository struct {
	mock.Mock
}

func (m *MockTeacherRepository) Create(ctx context.Context, teacher *models.Teacher) error {
	args := m.Called(ctx, teacher)
	return args.Error(0)
}

func (m *MockTeacherRepository) FindByID(ctx context.Context, id string) (*models.Teacher, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Teacher), args.Error(1)
}

func (m *MockTeacherRepository) FindByUserID(ctx context.Context, userID string) (*models.Teacher, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Teacher), args.Error(1)
}

func (m *MockTeacherRepository) FindAll(ctx context.Context) ([]models.Teacher, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Teacher), args.Error(1)
}

func (m *MockTeacherRepository) Update(ctx context.Context, teacher *models.Teacher) error {
	args := m.Called(ctx, teacher)
	return args.Error(0)
}

func (m *MockTeacherRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTeacherService_GetAllTeachers_Success(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)

	expected := []models.Teacher{{TeacherID: "teach-1", Specialization: "Math"}}
	mockRepo.On("FindAll", mock.Anything).Return(expected, nil)

	res, err := svc.GetAllTeachers(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "Math", res[0].Specialization)
	mockRepo.AssertExpectations(t)
}

func TestTeacherService_GetTeacherByID_NotFound(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)

	mockRepo.On("FindByID", mock.Anything, "invalid").Return(nil, errors.New("not found"))

	res, err := svc.GetTeacherByID(context.Background(), "invalid")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "not found")
}

func TestTeacherService_UpdateTeacher_Success(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)

	existing := &models.Teacher{TeacherID: "teach-1", Specialization: "Old"}
	mockRepo.On("FindByID", mock.Anything, "teach-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Teacher")).Return(nil)

	req := &dto.UpdateTeacherRequest{Specialization: "New"}
	res, err := svc.UpdateTeacher(context.Background(), "teach-1", req)

	assert.NoError(t, err)
	assert.Equal(t, "New", res.Specialization)
	mockRepo.AssertExpectations(t)
}

func TestTeacherService_UpdateTeacher_EmptyFields(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)

	existing := &models.Teacher{
		TeacherID:      "teach-1",
		Specialization: "old-spec",
		ApprovalStatus: "old-status",
	}
	mockRepo.On("FindByID", mock.Anything, "teach-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Teacher")).Return(nil)

	// Send an empty update request
	req := &dto.UpdateTeacherRequest{}
	res, err := svc.UpdateTeacher(context.Background(), "teach-1", req)

	assert.NoError(t, err)
	// Verify fields were NOT updated to empty string
	assert.Equal(t, "old-spec", res.Specialization)
	assert.Equal(t, "old-status", res.ApprovalStatus)

	mockRepo.AssertExpectations(t)
}

func TestTeacherService_CreateTeacher_ValidationError(t *testing.T) {
	svc := services.NewTeacherService(nil)
	req := &dto.CreateTeacherRequest{
		Specialization: "", // Assuming empty specialization fails validation
	}
	res, err := svc.CreateTeacher(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestTeacherService_CreateTeacher_DBError(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)
	
	req := &dto.CreateTeacherRequest{
		Specialization: "Math",
		UserID: "usr-1",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Teacher")).Return(errors.New("db error"))

	res, err := svc.CreateTeacher(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestTeacherService_GetTeacherByID_Forbidden(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_teacher")
	ctx = context.WithValue(ctx, middlewares.IsExternalKey, true)
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	teach := &models.Teacher{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "teach-1").Return(teach, nil)

	res, err := svc.GetTeacherByID(ctx, "teach-1")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTeacherService_GetAllTeachers_Forbidden(t *testing.T) {
	svc := services.NewTeacherService(nil)
	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_teacher")
	ctx = context.WithValue(ctx, middlewares.IsExternalKey, true)

	res, err := svc.GetAllTeachers(ctx)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTeacherService_GetAllTeachers_DBError(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)
	
	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	mockRepo.On("FindAll", mock.Anything).Return(nil, errors.New("db error"))

	res, err := svc.GetAllTeachers(ctx)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestTeacherService_UpdateTeacher_ValidationError(t *testing.T) {
	// Create request that will fail validation
	mockRepo := new(MockTeacherRepository)
	svc2 := services.NewTeacherService(mockRepo)
	mockRepo.On("FindByID", mock.Anything, "teach-1").Return(nil, errors.New("db error"))

	res, err := svc2.UpdateTeacher(context.Background(), "teach-1", &dto.UpdateTeacherRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestTeacherService_UpdateTeacher_Forbidden(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_teacher")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	teach := &models.Teacher{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "teach-1").Return(teach, nil)

	res, err := svc.UpdateTeacher(ctx, "teach-1", &dto.UpdateTeacherRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTeacherService_UpdateTeacher_DBError(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	teach := &models.Teacher{UserID: "usr-123"}
	
	mockRepo.On("FindByID", mock.Anything, "teach-1").Return(teach, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Teacher")).Return(errors.New("db error"))

	res, err := svc.UpdateTeacher(ctx, "teach-1", &dto.UpdateTeacherRequest{ApprovalStatus: "approved", ApprovedBy: "admin"})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestTeacherService_DeleteTeacher_Forbidden(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_teacher")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	teach := &models.Teacher{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "teach-1").Return(teach, nil)

	err := svc.DeleteTeacher(ctx, "teach-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTeacherService_DeleteTeacher_DBError(t *testing.T) {
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)

	mockRepo.On("FindByID", mock.Anything, "teach-1").Return(nil, errors.New("db error"))

	err := svc.DeleteTeacher(context.Background(), "teach-1")
	assert.Error(t, err)
}
