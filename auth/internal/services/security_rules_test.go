package services_test

import (
	"context"
	"testing"

	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/models"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Reglas de permisos: nadie se registra ni se crea como administrador sin serlo, y
// nadie se aprueba solo.

func ctxConRol(role, userID string) context.Context {
	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, role)
	return context.WithValue(ctx, middlewares.UserIDKey, userID)
}

func TestAuthService_Register_RechazaRolesNoPublicos(t *testing.T) {
	for _, role := range []string{"admin", "role_admin", "dev", "role_dev", "role_inventado"} {
		t.Run(role, func(t *testing.T) {
			mockUserRepo := new(MockUserRepository)
			mockTransactor := new(MockTransactor)
			svc := services.NewAuthService(mockTransactor, nil, mockUserRepo, nil, nil, nil)
			mockUserRepo.On("FindByEmail", mock.Anything, "nuevo@gmail.com").Return(nil, assert.AnError)

			res, err := svc.Register(context.Background(), &dto.RegisterRequest{
				Email: "nuevo@gmail.com", Password: "Password123!", Role: role,
				FirstName: "Ana", LastName: "Pérez",
			})

			assert.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "no se puede usar para registrarse")
			mockTransactor.AssertNotCalled(t, "RunInTx", mock.Anything, mock.Anything)
		})
	}
}

func TestUserService_CreateUser_SoloAdminOVoluntarioConParticipantes(t *testing.T) {
	casos := []struct {
		quien, rolNuevo string
		permitido       bool
	}{
		{"role_admin", "role_admin", true},
		{"role_dev", "role_teacher", true},
		{"role_volunteer", "role_participant", true},
		{"role_volunteer", "role_admin", false},
		{"role_volunteer", "role_volunteer", false},
		{"role_participant", "role_admin", false},
		{"role_company", "role_participant", false},
		{"role_teacher", "role_teacher", false},
		{"", "role_participant", false},
	}
	for _, c := range casos {
		t.Run(c.quien+"_crea_"+c.rolNuevo, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			svc := services.NewUserService(mockRepo)
			if c.permitido {
				mockRepo.On("CreateWithRoleProfile", mock.Anything, mock.AnythingOfType("*models.User"), c.rolNuevo, "usr-quien").Return(nil)
			}

			res, err := svc.CreateUser(ctxConRol(c.quien, "usr-quien"), &dto.CreateUserRequest{
				FirstName: "Ana", LastName: "Pérez", Email: "ana@gmail.com",
				Password: "Password123!", Phone: "3547000000", RoleID: c.rolNuevo, DNI: "30111222", Address: "San Martín 100",
			})

			if c.permitido {
				assert.NoError(t, err)
				assert.Equal(t, c.rolNuevo, res.RoleID)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "forbidden")
				mockRepo.AssertNotCalled(t, "CreateWithRoleProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestVolunteerService_UpdateVolunteer_NoSeApruebaSolo(t *testing.T) {
	mockRepo := new(MockVolunteerRepository)
	svc := services.NewVolunteerService(mockRepo)
	existing := &models.Volunteer{VolunteerID: "vol-1", UserID: "usr-vol", ApprovalStatus: "pending"}
	mockRepo.On("FindByID", mock.Anything, "vol-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Volunteer")).Return(nil)

	res, err := svc.UpdateVolunteer(ctxConRol("role_volunteer", "usr-vol"), "vol-1",
		&dto.UpdateVolunteerRequest{ApprovalStatus: "approved"})

	assert.NoError(t, err)
	assert.Equal(t, "pending", res.ApprovalStatus)
}

func TestCompanyService_CreateCompany_EmpresaNoSeCreaAprobada(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Company")).Return(nil)

	res, err := svc.CreateCompany(ctxConRol("role_company", "usr-emp"), &dto.CreateCompanyRequest{
		UserID: "usr-emp", BusinessName: "Huerta SA", Cuit: "30-11111111-1", ApprovalStatus: "approved",
	})

	assert.NoError(t, err)
	assert.Equal(t, "pending", res.ApprovalStatus)
}

func TestParticipantService_UpdateParticipant_VoluntarioApruebaParticipantes(t *testing.T) {
	mockRepo := new(MockParticipantRepository)
	svc := services.NewParticipantService(mockRepo, nil, nil)
	existing := &models.Participant{ParticipantID: "part-1", UserID: "usr-part", ApprovalStatus: "pending"}
	mockRepo.On("FindByID", mock.Anything, "part-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Participant")).Return(nil)

	res, err := svc.UpdateParticipant(ctxConRol("role_volunteer", "usr-vol"), "part-1",
		&dto.UpdateParticipantRequest{ApprovalStatus: "approved"})

	assert.NoError(t, err)
	assert.Equal(t, "approved", res.ApprovalStatus)
}

func TestTeacherService_UpdateTeacher_ExternoNoSePasaAInterno(t *testing.T) {
	interno := false
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)
	existing := &models.Teacher{TeacherID: "tech-1", UserID: "usr-doc", IsExternal: true, ApprovalStatus: "approved"}
	mockRepo.On("FindByID", mock.Anything, "tech-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Teacher")).Return(nil)

	res, err := svc.UpdateTeacher(ctxConRol("role_teacher", "usr-doc"), "tech-1",
		&dto.UpdateTeacherRequest{IsExternal: &interno})

	assert.NoError(t, err)
	assert.True(t, res.IsExternal)
}

func TestTeacherService_UpdateTeacher_AdminCambiaExterno(t *testing.T) {
	interno := false
	mockRepo := new(MockTeacherRepository)
	svc := services.NewTeacherService(mockRepo)
	existing := &models.Teacher{TeacherID: "tech-1", UserID: "usr-doc", IsExternal: true}
	mockRepo.On("FindByID", mock.Anything, "tech-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Teacher")).Return(nil)

	res, err := svc.UpdateTeacher(ctxConRol("role_admin", "usr-admin"), "tech-1",
		&dto.UpdateTeacherRequest{IsExternal: &interno})

	assert.NoError(t, err)
	assert.False(t, res.IsExternal)
}

func TestFiltrarUsuarios(t *testing.T) {
	users := []dto.UserResponse{
		{UserID: "1", FirstName: "Ana", LastName: "Gómez", Email: "ana@gmail.com", RoleID: "role_teacher"},
		{UserID: "2", FirstName: "Luis", LastName: "Núñez", Email: "lnunez@gmail.com", Phone: "3547444", RoleID: "role_participant"},
		{UserID: "3", FirstName: "Marta", LastName: "Gomez", Email: "marta@gmail.com", RoleID: "role_participant"},
	}
	ids := func(us []dto.UserResponse) []string {
		out := []string{}
		for _, u := range us {
			out = append(out, u.UserID)
		}
		return out
	}

	assert.Equal(t, []string{"1", "2", "3"}, ids(services.FiltrarUsuarios(users, "", "")))
	assert.Equal(t, []string{"1", "3"}, ids(services.FiltrarUsuarios(users, "  GOMEZ ", "")), "sin tildes ni mayúsculas")
	assert.Equal(t, []string{"3"}, ids(services.FiltrarUsuarios(users, "gómez", "role_participant")), "combina texto y rol")
	assert.Equal(t, []string{"2"}, ids(services.FiltrarUsuarios(users, "nunez", "")), "ñ como n")
	assert.Equal(t, []string{"1"}, ids(services.FiltrarUsuarios(users, "", "role_teacher")))
	assert.Equal(t, []string{"2"}, ids(services.FiltrarUsuarios(users, "lnunez@", "")), "por email")
	assert.Equal(t, []string{"2"}, ids(services.FiltrarUsuarios(users, "444", "")), "por teléfono")
	assert.Empty(t, services.FiltrarUsuarios(users, "zzz", ""))
	assert.NotNil(t, services.FiltrarUsuarios(nil, "", ""), "lista vacía, no null")
}
