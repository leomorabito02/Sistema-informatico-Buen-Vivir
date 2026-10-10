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
