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
	"auth/internal/services"
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

func TestGrpcGetParticipant(t *testing.T) {
	partSvc := new(MockParticipantService)
	ctrl := controllers.NewGrpcAuthController(partSvc, nil, nil, nil)

	// Missing participant_id
	_, err := ctrl.GetParticipant(context.Background(), &authpb.GetParticipantRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// Internal error
	partSvc.On("GetParticipantByID", mock.Anything, "p-err").Return(nil, errors.New("db fail"))
	_, err = ctrl.GetParticipant(context.Background(), &authpb.GetParticipantRequest{ParticipantId: "p-err"})
	assert.Equal(t, codes.Internal, status.Code(err))

	// Not found (nil)
	partSvc.On("GetParticipantByID", mock.Anything, "p-nil").Return(nil, nil)
	_, err = ctrl.GetParticipant(context.Background(), &authpb.GetParticipantRequest{ParticipantId: "p-nil"})
	assert.Equal(t, codes.NotFound, status.Code(err))

	// Success
	partSvc.On("GetParticipantByID", mock.Anything, "p-ok").Return(&dto.ParticipantResponse{
		ParticipantID:  "p-ok",
		UserID:         "u-1",
		ApprovalStatus: "approved",
	}, nil)
	resp, err := ctrl.GetParticipant(context.Background(), &authpb.GetParticipantRequest{ParticipantId: "p-ok"})
	assert.NoError(t, err)
	assert.Equal(t, "p-ok", resp.Id)
	assert.Equal(t, "u-1", resp.UserId)
	assert.Equal(t, "approved", resp.ApprovalStatus)
}

func TestGrpcGetParticipantProfile(t *testing.T) {
	fakeDir := &fakeDirectoryService{}
	ctrl := controllers.NewGrpcAuthController(nil, nil, nil, fakeDir)

	// Missing lookup
	_, err := ctrl.GetParticipantProfile(context.Background(), &authpb.GetParticipantProfileRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// Success with ParticipantId
	resp, err := ctrl.GetParticipantProfile(context.Background(), &authpb.GetParticipantProfileRequest{
		Lookup: &authpb.GetParticipantProfileRequest_ParticipantId{ParticipantId: "part-123"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "part-123", resp.ParticipantId)
	assert.Equal(t, "Ana", resp.FirstName)
	assert.Equal(t, "2000-05-03", resp.Birthdate)

	// Success with UserId
	resp, err = ctrl.GetParticipantProfile(context.Background(), &authpb.GetParticipantProfileRequest{
		Lookup: &authpb.GetParticipantProfileRequest_UserId{UserId: "usr-123"},
	})
	assert.NoError(t, err)
	assert.Equal(t, "usr-123", resp.UserId)

	// Directory errors
	fakeDir.err = services.ErrInvalidDirectoryRequest
	_, err = ctrl.GetParticipantProfile(context.Background(), &authpb.GetParticipantProfileRequest{
		Lookup: &authpb.GetParticipantProfileRequest_ParticipantId{ParticipantId: "err"},
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	fakeDir.err = repositories.ErrParticipantProfileNotFound
	_, err = ctrl.GetParticipantProfile(context.Background(), &authpb.GetParticipantProfileRequest{
		Lookup: &authpb.GetParticipantProfileRequest_ParticipantId{ParticipantId: "err"},
	})
	assert.Equal(t, codes.NotFound, status.Code(err))

	fakeDir.err = errors.New("unexpected error")
	_, err = ctrl.GetParticipantProfile(context.Background(), &authpb.GetParticipantProfileRequest{
		Lookup: &authpb.GetParticipantProfileRequest_ParticipantId{ParticipantId: "err"},
	})
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestGrpcSearchParticipants(t *testing.T) {
	fakeDir := &fakeDirectoryService{}
	ctrl := controllers.NewGrpcAuthController(nil, nil, nil, fakeDir)

	// Success
	resp, err := ctrl.SearchParticipants(context.Background(), &authpb.SearchParticipantsRequest{Query: "test", Limit: 10})
	assert.NoError(t, err)
	assert.Len(t, resp.Participants, 2)

	// Error
	fakeDir.err = errors.New("fail")
	_, err = ctrl.SearchParticipants(context.Background(), &authpb.SearchParticipantsRequest{})
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestGrpcGetUsersByIDs(t *testing.T) {
	fakeDir := &fakeDirectoryService{}
	ctrl := controllers.NewGrpcAuthController(nil, nil, nil, fakeDir)

	// Success
	resp, err := ctrl.GetUsersByIDs(context.Background(), &authpb.GetUsersByIDsRequest{UserIds: []string{"usr-1", "usr-2"}})
	assert.NoError(t, err)
	assert.Len(t, resp.Users, 2)
	assert.Equal(t, "Ana", resp.Users[0].FirstName)

	// Error
	fakeDir.err = errors.New("fail")
	_, err = ctrl.GetUsersByIDs(context.Background(), &authpb.GetUsersByIDsRequest{})
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestGrpcIssueActionLink(t *testing.T) {
	userSvc := new(MockUserService)
	verifSvc := new(MockVerificationService)
	recovSvc := new(MockPasswordRecoveryService)
	twofaSvc := new(MockTwofaService)

	ctrl := controllers.NewGrpcAuthController(nil, nil, userSvc, nil)

	// Nil request
	_, err := ctrl.IssueActionLink(context.Background(), nil)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// Unsupported type
	_, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{Type: "unknown"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	// cuenta_aprobada
	resp, err := ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{Type: "cuenta_aprobada"})
	assert.NoError(t, err)
	assert.Contains(t, resp.ActionUrl, "/login")

	// verificacion before SetLinkIssuers -> Unimplemented
	_, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{Type: "verificacion"})
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// recuperacion_password before SetLinkIssuers -> Unimplemented
	_, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{Type: "recuperacion_password"})
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// two_factor_code before SetLinkIssuers -> Unimplemented
	_, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{Type: "two_factor_code"})
	assert.Equal(t, codes.Unimplemented, status.Code(err))

	// Now set link issuers
	ctrl.SetLinkIssuers(verifSvc, recovSvc, twofaSvc)

	// verificacion error
	verifSvc.On("IssueActionLink", mock.Anything, "req-1", "u-1").Return("", errors.New("err"))
	_, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{
		Type: "verificacion", RequestId: "req-1", UserId: "u-1",
	})
	assert.Equal(t, codes.Internal, status.Code(err))

	// verificacion success
	verifSvc.On("IssueActionLink", mock.Anything, "req-2", "u-2").Return("https://app/verify", nil)
	resp, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{
		Type: "verificacion", RequestId: "req-2", UserId: "u-2",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://app/verify", resp.ActionUrl)

	// recuperacion_password error
	recovSvc.On("IssueActionLink", mock.Anything, "req-3", "u-3").Return("", errors.New("err"))
	_, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{
		Type: "recuperacion_password", RequestId: "req-3", UserId: "u-3",
	})
	assert.Equal(t, codes.Internal, status.Code(err))

	// recuperacion_password success
	recovSvc.On("IssueActionLink", mock.Anything, "req-4", "u-4").Return("https://app/recover", nil)
	resp, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{
		Type: "recuperacion_password", RequestId: "req-4", UserId: "u-4",
	})
	assert.NoError(t, err)
	assert.Equal(t, "https://app/recover", resp.ActionUrl)

	// two_factor_code user not found
	userSvc.On("GetUserByID", mock.Anything, "u-404").Return(nil, errors.New("not found"))
	_, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{
		Type: "two_factor_code", UserId: "u-404",
	})
	assert.Equal(t, codes.NotFound, status.Code(err))

	// two_factor_code empty code
	userSvc.On("GetUserByID", mock.Anything, "u-nocode").Return(&dto.UserResponse{Email: "nocode@example.com"}, nil)
	twofaSvc.On("GetActiveCodeForEmail", "nocode@example.com").Return("")
	_, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{
		Type: "two_factor_code", UserId: "u-nocode",
	})
	assert.Equal(t, codes.NotFound, status.Code(err))

	// two_factor_code success
	userSvc.On("GetUserByID", mock.Anything, "u-good").Return(&dto.UserResponse{Email: "good@example.com"}, nil)
	twofaSvc.On("GetActiveCodeForEmail", "good@example.com").Return("123456")
	resp, err = ctrl.IssueActionLink(context.Background(), &authpb.IssueActionLinkRequest{
		Type: "two_factor_code", UserId: "u-good",
	})
	assert.NoError(t, err)
	assert.Equal(t, "123456", resp.ActionUrl)
}


