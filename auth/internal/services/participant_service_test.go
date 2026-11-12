package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth/internal/middlewares"

	"auth/internal/dto"
	"auth/internal/models"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockParticipantRepository struct {
	mock.Mock
}

func (m *MockParticipantRepository) Create(ctx context.Context, part *models.Participant) error {
	return m.Called(ctx, part).Error(0)
}

func (m *MockParticipantRepository) mockParticipantReturn(args mock.Arguments) (*models.Participant, error) {
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Participant), args.Error(1)
}

func (m *MockParticipantRepository) FindByID(ctx context.Context, id string) (*models.Participant, error) {
	return m.mockParticipantReturn(m.Called(ctx, id))
}

func (m *MockParticipantRepository) FindByUserID(ctx context.Context, userID string) (*models.Participant, error) {
	return m.mockParticipantReturn(m.Called(ctx, userID))
}

func (m *MockParticipantRepository) FindAll(ctx context.Context) ([]models.Participant, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Participant), args.Error(1)
}

func (m *MockParticipantRepository) Update(ctx context.Context, part *models.Participant) error {
	return m.Called(ctx, part).Error(0)
}

func (m *MockParticipantRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func TestParticipantService_GetAllParticipants_Success(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)

	expected := []models.Participant{{ParticipantID: "part-1", Address: "ESBN"}}
	mockRepo.On("FindAll", mock.Anything).Return(expected, nil)

	res, err := svc.GetAllParticipants(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "ESBN", res[0].Address)
	mockRepo.AssertExpectations(t)
}

func TestParticipantService_GetParticipantByID_NotFound(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)

	mockRepo.On("FindByID", mock.Anything, "invalid").Return(nil, errors.New("not found"))

	res, err := svc.GetParticipantByID(context.Background(), "invalid")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "not found")
}

func TestParticipantService_UpdateParticipant_Success(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)

	existing := &models.Participant{ParticipantID: "part-1", Address: "Old"}
	mockRepo.On("FindByID", mock.Anything, "part-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Participant")).Return(nil)

	req := &dto.UpdateParticipantRequest{Address: "New"}
	res, err := svc.UpdateParticipant(context.Background(), "part-1", req)

	assert.NoError(t, err)
	assert.Equal(t, "New", res.Address)
	mockRepo.AssertExpectations(t)
}

func TestParticipantService_UpdateParticipant_EmptyFields(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)

	existing := &models.Participant{
		ParticipantID:         "part-1",
		DNI:                   "old-dni",
		Address:               "old-addr",
		Gender:                "old-gender",
		EmergencyContactName:  "old-name",
		EmergencyContactPhone: "old-phone",
		ImageURL:              "old-image",
		DocURL:                "old-doc",
	}
	mockRepo.On("FindByID", mock.Anything, "part-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Participant")).Return(nil)

	// Send an empty update request
	req := &dto.UpdateParticipantRequest{}
	res, err := svc.UpdateParticipant(context.Background(), "part-1", req)

	assert.NoError(t, err)
	// Verify fields were NOT updated to empty string
	assert.Equal(t, "old-dni", res.DNI)
	assert.Equal(t, "old-addr", res.Address)
	assert.Equal(t, "old-gender", res.Gender)
	assert.Equal(t, "old-name", res.EmergencyContactName)
	assert.Equal(t, "old-phone", res.EmergencyContactPhone)
	assert.Equal(t, "old-image", res.ImageURL)
	assert.Equal(t, "old-doc", res.DocURL)

	mockRepo.AssertExpectations(t)
}

func TestParticipantService_CreateParticipant_ValidationError(t *testing.T) {
	svc := services.NewParticipantService(nil, nil, nil)
	req := &dto.CreateParticipantRequest{
		DNI: "", // Assuming empty DNI fails validation
	}
	res, err := svc.CreateParticipant(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestParticipantService_CreateParticipant_DBError(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)
	
	req := &dto.CreateParticipantRequest{
		DNI: "12345678",
		UserID: "usr-1",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Participant")).Return(errors.New("db error"))

	res, err := svc.CreateParticipant(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestParticipantService_GetParticipantByID_Forbidden(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	part := &models.Participant{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "part-1").Return(part, nil)

	res, err := svc.GetParticipantByID(ctx, "part-1")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestParticipantService_GetAllParticipants_Forbidden(t *testing.T) {
	svc := services.NewParticipantService(nil, nil, nil)
	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")

	res, err := svc.GetAllParticipants(ctx)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestParticipantService_GetAllParticipants_DBError(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)
	
	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	mockRepo.On("FindAll", mock.Anything).Return(nil, errors.New("db error"))

	res, err := svc.GetAllParticipants(ctx)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestParticipantService_UpdateParticipant_ValidationError(t *testing.T) {
	// Create request that will fail validation
	mockRepo := new(MockParticipantRepository)
	svc2 := services.NewParticipantService(mockRepo, nil, nil)
	mockRepo.On("FindByID", mock.Anything, "part-1").Return(nil, errors.New("db error"))

	res, err := svc2.UpdateParticipant(context.Background(), "part-1", &dto.UpdateParticipantRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestParticipantService_UpdateParticipant_Forbidden(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	part := &models.Participant{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "part-1").Return(part, nil)

	res, err := svc.UpdateParticipant(ctx, "part-1", &dto.UpdateParticipantRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestParticipantService_UpdateParticipant_DBError(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	part := &models.Participant{UserID: "usr-123"}
	
	mockRepo.On("FindByID", mock.Anything, "part-1").Return(part, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Participant")).Return(errors.New("db error"))

	res, err := svc.UpdateParticipant(ctx, "part-1", &dto.UpdateParticipantRequest{ApprovalStatus: "approved", ApprovedBy: "admin"})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestParticipantService_DeleteParticipant_Forbidden(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	part := &models.Participant{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "part-1").Return(part, nil)

	err := svc.DeleteParticipant(ctx, "part-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestParticipantService_DeleteParticipant_DBError(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)

	mockRepo.On("FindByID", mock.Anything, "part-1").Return(nil, errors.New("db error"))

	err := svc.DeleteParticipant(context.Background(), "part-1")
	assert.Error(t, err)
}

func TestParticipantService_EdgeCases(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	mockUserRepo := new(MockUserRepository)
	producer := &mockUsuariosProducer{}
	svc := services.NewParticipantService(mockRepo, nil, nil, mockUserRepo, producer)

	// CreateParticipant with producer & status != approved
	mockUserRepo.On("CountActiveParticipants", mock.Anything).Return(int64(5), nil).Maybe()
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Participant) bool {
		return p.ApprovalStatus == "pending"
	})).Return(nil)

	req := &dto.CreateParticipantRequest{
		UserID:    "u-1",
		Birthdate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		DNI:       "12345678",
		Gender:    "otro",
	}
	// Lo crea el propio usuario (los pedidos siempre llegan autenticados).
	ownerCtx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ownerCtx = context.WithValue(ownerCtx, middlewares.UserIDKey, "u-1")
	res, err := svc.CreateParticipant(ownerCtx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "pending", res.ApprovalStatus)

	// GetParticipantByID - forbidden for participant
	part := &models.Participant{ParticipantID: "part-1", UserID: "u-owner"}
	mockRepo.On("FindByID", mock.Anything, "part-1").Return(part, nil)
	ctxOther := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ctxOther = context.WithValue(ctxOther, middlewares.UserIDKey, "u-someone-else")
	_, err = svc.GetParticipantByID(ctxOther, "part-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")

	// GetParticipantByID - success for owner
	ctxOwner := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ctxOwner = context.WithValue(ctxOwner, middlewares.UserIDKey, "u-owner")
	res, err = svc.GetParticipantByID(ctxOwner, "part-1")
	assert.NoError(t, err)
	assert.Equal(t, "part-1", res.ParticipantID)

	// GetAllParticipants - forbidden for participant
	_, err = svc.GetAllParticipants(ctxOther)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")

	// DeleteParticipant - success with producer
	mockRepo.On("Delete", mock.Anything, "part-1").Return(nil)
	ctxAdmin := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	err = svc.DeleteParticipant(ctxAdmin, "part-1")
	assert.NoError(t, err)
}

// Auditoría 2026-11-11 (Problema 4): nadie sin rol de gestión crea un
// perfil a nombre de otra cuenta, y el participante no da de baja perfiles
// de participante (HU-09).
func TestParticipantService_PermisosDeAltaYBaja(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	mockUserRepo := new(MockUserRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil, mockUserRepo, nil)
	mockUserRepo.On("CountActiveParticipants", mock.Anything).Return(int64(1), nil).Maybe()
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	req := &dto.CreateParticipantRequest{UserID: "usr-victima", Birthdate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), Gender: "otro"}
	ctxDe := func(rol, uid string) context.Context {
		ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, rol)
		return context.WithValue(ctx, middlewares.UserIDKey, uid)
	}

	for _, rol := range []string{"role_participant", "role_company", "role_teacher"} {
		_, err := svc.CreateParticipant(ctxDe(rol, "usr-atacante"), req)
		assert.ErrorContains(t, err, "forbidden", rol)
	}
	for _, rol := range []string{"role_admin", "role_dev", "role_volunteer"} {
		_, err := svc.CreateParticipant(ctxDe(rol, "usr-staff"), req)
		assert.NoError(t, err, rol)
	}
	_, err := svc.CreateParticipant(context.Background(), req)
	assert.ErrorContains(t, err, "forbidden", "sin usuario autenticado")

	part := &models.Participant{ParticipantID: "part-propio", UserID: "usr-1"}
	mockRepo.On("FindByID", mock.Anything, "part-propio").Return(part, nil)
	err = svc.DeleteParticipant(ctxDe("role_participant", "usr-1"), "part-propio")
	assert.ErrorContains(t, err, "forbidden")
	mockRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}
