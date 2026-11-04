package controllers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"auth/api/proto/authpb"
	"auth/internal/controllers"
	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/repositories"
)

func TestGrpcGetCompany_LeeElUsuarioConPermisosDeSistema(t *testing.T) {
	companySvc := new(MockCompanyService)
	userSvc := new(MockUserService)
	ctrl := controllers.NewGrpcAuthController(nil, companySvc, userSvc, nil)

	// Inserción Laboral manda el id de usuario de la empresa.
	companySvc.On("GetCompanyByID", mock.Anything, "usr-emp").Return(&dto.CompanyResponse{
		CompanyID: "comp-1", UserID: "usr-emp", BusinessName: "Vivero", Cuit: "30-1", ApprovalStatus: "approved",
	}, nil)
	userSvc.On("GetUserByID", mock.MatchedBy(func(ctx context.Context) bool {
		role, _ := ctx.Value(middlewares.RoleIDKey).(string)
		return role == "role_dev"
	}), "usr-emp").Return(&dto.UserResponse{UserID: "usr-emp", Email: "e@x.test", Phone: "1", Status: "active"}, nil)

	res, err := ctrl.GetCompany(context.Background(), &authpb.GetCompanyRequest{CompanyId: "usr-emp"})
	assert.NoError(t, err)
	assert.Equal(t, "comp-1", res.CompanyId)
	assert.Equal(t, "e@x.test", res.UserEmail)
	assert.Equal(t, "active", res.UserStatus)
	userSvc.AssertExpectations(t)
}

func TestGrpcGetCompany_Errores(t *testing.T) {
	ctrl := controllers.NewGrpcAuthController(nil, new(MockCompanyService), new(MockUserService), nil)
	_, err := ctrl.GetCompany(context.Background(), &authpb.GetCompanyRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	companySvc := new(MockCompanyService)
	companySvc.On("GetCompanyByID", mock.Anything, "x").Return(nil, errors.New("empresa no encontrada"))
	ctrl = controllers.NewGrpcAuthController(nil, companySvc, new(MockUserService), nil)
	_, err = ctrl.GetCompany(context.Background(), &authpb.GetCompanyRequest{CompanyId: "x"})
	assert.Equal(t, codes.Internal, status.Code(err))

	companySvc = new(MockCompanyService)
	companySvc.On("GetCompanyByID", mock.Anything, "y").Return(&dto.CompanyResponse{UserID: "usr-y"}, nil)
	userSvc := new(MockUserService)
	userSvc.On("GetUserByID", mock.Anything, "usr-y").Return(nil, errors.New("usuario no encontrado"))
	ctrl = controllers.NewGrpcAuthController(nil, companySvc, userSvc, nil)
	_, err = ctrl.GetCompany(context.Background(), &authpb.GetCompanyRequest{CompanyId: "y"})
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestGrpcGetUserDeviceTokens_Success(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	mockRepo.On("GetUserTokens", mock.Anything, "user-123").Return([]string{"tok-a", "tok-b"}, nil)

	ctrl := controllers.NewGrpcAuthController(nil, nil, nil, nil, mockRepo)

	res, err := ctrl.GetUserDeviceTokens(context.Background(), &authpb.GetUserDeviceTokensRequest{UserId: "user-123"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"tok-a", "tok-b"}, res.Tokens)
	mockRepo.AssertExpectations(t)
}

func TestGrpcGetUserDeviceTokens_MissingUserID(t *testing.T) {
	ctrl := controllers.NewGrpcAuthController(nil, nil, nil, nil)
	_, err := ctrl.GetUserDeviceTokens(context.Background(), &authpb.GetUserDeviceTokensRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGrpcRevokeDeviceTokens_Success(t *testing.T) {
	mockRepo := new(MockDeviceRepository)
	mockRepo.On("DeleteDeviceTokens", mock.Anything, "user-123", []string{"tok-a"}).Return(nil)

	ctrl := controllers.NewGrpcAuthController(nil, nil, nil, nil, mockRepo)

	res, err := ctrl.RevokeDeviceTokens(context.Background(), &authpb.RevokeDeviceTokensRequest{
		UserId: "user-123",
		Tokens: []string{"tok-a"},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(1), res.RevokedCount)
	mockRepo.AssertExpectations(t)
}

func TestGrpcRevokeDeviceTokens_MissingUserID(t *testing.T) {
	ctrl := controllers.NewGrpcAuthController(nil, nil, nil, nil)
	_, err := ctrl.RevokeDeviceTokens(context.Background(), &authpb.RevokeDeviceTokensRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGrpcGetCampaignRecipients_Success(t *testing.T) {
	userSvc := new(MockUserService)
	ctrl := controllers.NewGrpcAuthController(nil, nil, userSvc, nil)

	userSvc.On("GetCampaignEligibleUsers", mock.Anything).Return([]repositories.CampaignEligibleUser{
		{UserID: "usr-1", Email: "u1@test.com", FirstName: "Ana", LastName: "Gomez"},
		{UserID: "usr-2", Email: "u2@test.com", FirstName: "Carlos", LastName: "Perez"},
	}, nil)

	res, err := ctrl.GetCampaignRecipients(context.Background(), &authpb.GetCampaignRecipientsRequest{})
	assert.NoError(t, err)
	assert.Len(t, res.Recipients, 2)
	assert.Equal(t, "usr-1", res.Recipients[0].UserId)
	assert.Equal(t, "u1@test.com", res.Recipients[0].Email)
	assert.Equal(t, "Ana", res.Recipients[0].FirstName)
	userSvc.AssertExpectations(t)
}

func TestGrpcGetCampaignRecipients_Error(t *testing.T) {
	userSvc := new(MockUserService)
	ctrl := controllers.NewGrpcAuthController(nil, nil, userSvc, nil)

	userSvc.On("GetCampaignEligibleUsers", mock.Anything).Return(nil, errors.New("db error"))

	_, err := ctrl.GetCampaignRecipients(context.Background(), &authpb.GetCampaignRecipientsRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestGrpcCheckUserEmailVerified_Success(t *testing.T) {
	userSvc := new(MockUserService)
	ctrl := controllers.NewGrpcAuthController(nil, nil, userSvc, nil)

	userSvc.On("IsEmailVerified", mock.Anything, "usr-verified").Return(true, nil)
	userSvc.On("IsEmailVerified", mock.Anything, "usr-unverified").Return(false, nil)

	res1, err := ctrl.CheckUserEmailVerified(context.Background(), &authpb.CheckUserEmailVerifiedRequest{UserId: "usr-verified"})
	assert.NoError(t, err)
	assert.True(t, res1.IsVerified)

	res2, err := ctrl.CheckUserEmailVerified(context.Background(), &authpb.CheckUserEmailVerifiedRequest{UserId: "usr-unverified"})
	assert.NoError(t, err)
	assert.False(t, res2.IsVerified)

	userSvc.AssertExpectations(t)
}

func TestGrpcCheckUserEmailVerified_MissingUserID(t *testing.T) {
	ctrl := controllers.NewGrpcAuthController(nil, nil, nil, nil)
	_, err := ctrl.CheckUserEmailVerified(context.Background(), &authpb.CheckUserEmailVerifiedRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGrpcGetUserContact_Success(t *testing.T) {
	userSvc := new(MockUserService)
	ctrl := controllers.NewGrpcAuthController(nil, nil, userSvc, nil)

	userSvc.On("GetUserByID", mock.MatchedBy(func(ctx context.Context) bool {
		role, _ := ctx.Value(middlewares.RoleIDKey).(string)
		return role == "role_dev"
	}), "usr-123").Return(&dto.UserResponse{
		UserID:          "usr-123",
		Email:           "usr123@test.com",
		FirstName:       "Juan",
		LastName:        "Perez",
		IsEmailVerified: true,
	}, nil)

	res, err := ctrl.GetUserContact(context.Background(), &authpb.GetUserContactRequest{UserId: "usr-123"})
	assert.NoError(t, err)
	assert.Equal(t, "usr-123", res.UserId)
	assert.Equal(t, "usr123@test.com", res.Email)
	assert.Equal(t, "Juan", res.FirstName)
	assert.Equal(t, "Perez", res.LastName)
	assert.True(t, res.IsEmailVerified)
	userSvc.AssertExpectations(t)
}

func TestGrpcGetUserContact_NotFound(t *testing.T) {
	userSvc := new(MockUserService)
	ctrl := controllers.NewGrpcAuthController(nil, nil, userSvc, nil)

	userSvc.On("GetUserByID", mock.Anything, "usr-404").Return(nil, errors.New("user not found"))

	_, err := ctrl.GetUserContact(context.Background(), &authpb.GetUserContactRequest{UserId: "usr-404"})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
	userSvc.AssertExpectations(t)
}

func TestGrpcGetUserContact_MissingUserID(t *testing.T) {
	ctrl := controllers.NewGrpcAuthController(nil, nil, nil, nil)
	_, err := ctrl.GetUserContact(context.Background(), &authpb.GetUserContactRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGrpcGetActiveUserIDsByRoles_Success(t *testing.T) {
	userSvc := new(MockUserService)
	ctrl := controllers.NewGrpcAuthController(nil, nil, userSvc, nil)

	userSvc.On("GetActiveUserIDsByRoles", mock.MatchedBy(func(ctx context.Context) bool {
		role, _ := ctx.Value(middlewares.RoleIDKey).(string)
		return role == "role_dev"
	}), []string{"role_admin", "role_volunteer"}).Return([]string{"admin-1", "vol-1"}, nil)

	res, err := ctrl.GetActiveUserIDsByRoles(context.Background(), &authpb.GetActiveUserIDsByRolesRequest{
		RoleIds: []string{"role_admin", "role_volunteer"},
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"admin-1", "vol-1"}, res.UserIds)
	userSvc.AssertExpectations(t)
}

func TestGrpcGetActiveUserIDsByRoles_Empty(t *testing.T) {
	ctrl := controllers.NewGrpcAuthController(nil, nil, nil, nil)
	_, err := ctrl.GetActiveUserIDsByRoles(context.Background(), &authpb.GetActiveUserIDsByRolesRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGrpcGetActiveUserIDsByRoles_ServiceError(t *testing.T) {
	userSvc := new(MockUserService)
	ctrl := controllers.NewGrpcAuthController(nil, nil, userSvc, nil)

	userSvc.On("GetActiveUserIDsByRoles", mock.Anything, []string{"role_admin"}).Return(nil, errors.New("db error"))

	_, err := ctrl.GetActiveUserIDsByRoles(context.Background(), &authpb.GetActiveUserIDsByRolesRequest{
		RoleIds: []string{"role_admin"},
	})
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	userSvc.AssertExpectations(t)
}


