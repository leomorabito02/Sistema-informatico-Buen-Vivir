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

type MockVolunteerRepository struct {
	mock.Mock
}

func (m *MockVolunteerRepository) Create(ctx context.Context, vol *models.Volunteer) error {
	return m.Called(ctx, vol).Error(0)
}

func (m *MockVolunteerRepository) mockVolunteerReturn(args mock.Arguments) (*models.Volunteer, error) {
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Volunteer), args.Error(1)
}

func (m *MockVolunteerRepository) FindByID(ctx context.Context, id string) (*models.Volunteer, error) {
	return m.mockVolunteerReturn(m.Called(ctx, id))
}

func (m *MockVolunteerRepository) FindByUserID(ctx context.Context, userID string) (*models.Volunteer, error) {
	return m.mockVolunteerReturn(m.Called(ctx, userID))
}

func (m *MockVolunteerRepository) FindAll(ctx context.Context) ([]models.Volunteer, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Volunteer), args.Error(1)
}

func (m *MockVolunteerRepository) Update(ctx context.Context, vol *models.Volunteer) error {
	return m.Called(ctx, vol).Error(0)
}

func (m *MockVolunteerRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func TestVolunteerService_GetAllVolunteers_Success(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)

	expected := []models.Volunteer{{VolunteerID: "vol-1", ApprovalStatus: "pending"}}
	mockRepo.On("FindAll", mock.Anything).Return(expected, nil)

	res, err := svc.GetAllVolunteers(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "pending", res[0].ApprovalStatus)
	mockRepo.AssertExpectations(t)
}

func TestVolunteerService_GetVolunteerByID_NotFound(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)

	mockRepo.On("FindByID", mock.Anything, "invalid").Return(nil, errors.New("not found"))

	res, err := svc.GetVolunteerByID(context.Background(), "invalid")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "not found")
}

func TestVolunteerService_UpdateVolunteer_Success(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)

	existing := &models.Volunteer{VolunteerID: "vol-1", ApprovalStatus: "pending"}
	mockRepo.On("FindByID", mock.Anything, "vol-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Volunteer")).Return(nil)

	req := &dto.UpdateVolunteerRequest{ApprovalStatus: "approved", ApprovedBy: "admin"}
	res, err := svc.UpdateVolunteer(context.Background(), "vol-1", req)

	assert.NoError(t, err)
	assert.Equal(t, "approved", res.ApprovalStatus)
	mockRepo.AssertExpectations(t)
}

func TestVolunteerService_UpdateVolunteer_EmptyFields(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)

	existing := &models.Volunteer{
		VolunteerID:    "vol-1",
		ApprovalStatus: "old-status",
	}
	mockRepo.On("FindByID", mock.Anything, "vol-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Volunteer")).Return(nil)

	// Send an empty update request
	req := &dto.UpdateVolunteerRequest{}
	res, err := svc.UpdateVolunteer(context.Background(), "vol-1", req)

	assert.NoError(t, err)
	// Verify fields were NOT updated to empty string
	assert.Equal(t, "old-status", res.ApprovalStatus)

	mockRepo.AssertExpectations(t)
}

func TestVolunteerService_CreateVolunteer_ValidationError(t *testing.T) {
	svc := services.NewVolunteerService(nil)
	req := &dto.CreateVolunteerRequest{
		UserID: "", // Assuming empty UserID fails validation
	}
	res, err := svc.CreateVolunteer(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestVolunteerService_CreateVolunteer_DBError(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)

	req := &dto.CreateVolunteerRequest{
		UserID: "usr-1",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Volunteer")).Return(errors.New("db error"))

	res, err := svc.CreateVolunteer(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestVolunteerService_GetAllVolunteers_DBError(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)

	mockRepo.On("FindAll", mock.Anything).Return(nil, errors.New("db error"))

	res, err := svc.GetAllVolunteers(context.Background())
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestVolunteerService_UpdateVolunteer_ValidationError(t *testing.T) {
	// Create request that will fail validation
	mockRepo := new(MockVolunteerRepository)
	svc2 := services.NewVolunteerService(mockRepo)
	mockRepo.On("FindByID", mock.Anything, "vol-1").Return(nil, errors.New("db error"))

	res, err := svc2.UpdateVolunteer(context.Background(), "vol-1", &dto.UpdateVolunteerRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestVolunteerService_UpdateVolunteer_Forbidden(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_volunteer")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	vol := &models.Volunteer{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "vol-1").Return(vol, nil)

	res, err := svc.UpdateVolunteer(ctx, "vol-1", &dto.UpdateVolunteerRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestVolunteerService_UpdateVolunteer_DBError(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	vol := &models.Volunteer{UserID: "usr-123"}

	mockRepo.On("FindByID", mock.Anything, "vol-1").Return(vol, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Volunteer")).Return(errors.New("db error"))

	res, err := svc.UpdateVolunteer(ctx, "vol-1", &dto.UpdateVolunteerRequest{ApprovalStatus: "approved", ApprovedBy: "admin"})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestVolunteerService_DeleteVolunteer_Forbidden(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_volunteer")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	vol := &models.Volunteer{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "vol-1").Return(vol, nil)

	err := svc.DeleteVolunteer(ctx, "vol-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestVolunteerService_DeleteVolunteer_DBError(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)

	mockRepo.On("FindByID", mock.Anything, "vol-1").Return(nil, errors.New("db error"))

	err := svc.DeleteVolunteer(context.Background(), "vol-1")
	assert.Error(t, err)
}
