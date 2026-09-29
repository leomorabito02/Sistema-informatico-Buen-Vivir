package services_test

import (
	"context"
	"errors"
	"testing"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type enrollFixture struct {
	repo    *MockEnrollmentRepository
	courses *MockCourseRepository
	auth    *fakeAuthServer
	svc     services.EnrollmentService
}

func newEnrollFixture(t *testing.T, participants ...models.AuthParticipant) *enrollFixture {
	f := &enrollFixture{
		repo:    new(MockEnrollmentRepository),
		courses: new(MockCourseRepository),
		auth:    &fakeAuthServer{participants: participants},
	}
	f.svc = services.NewEnrollmentService(f.repo, f.courses, f.auth.start(t))
	return f
}

func intPtr(n int) *int { return &n }

var participanteAprobado = models.AuthParticipant{ParticipantID: "p1", UserID: "u1", ApprovalStatus: "accepted"}

func TestEnroll_Success(t *testing.T) {
	f := newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)
	f.repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{}, nil)
	f.repo.On("Create", mock.Anything, models.Enrollment{
		ParticipanteID: "p1", CursoID: "c1", Estado: models.EnrollmentEstadoActiva,
	}).Return(&models.Enrollment{ID: "e1", ParticipanteID: "p1", CursoID: "c1", Estado: models.EnrollmentEstadoActiva}, nil)

	e, err := f.svc.Enroll(context.Background(), "  p1 ", "c1", "tok-123")

	require.NoError(t, err)
	assert.Equal(t, "e1", e.ID)
	assert.Equal(t, "tok-123", f.auth.receivedToken, "reenvía el JWT del usuario a auth-service")
	f.repo.AssertExpectations(t)
}

func TestEnroll_RequiredFields(t *testing.T) {
	f := newEnrollFixture(t)

	_, err := f.svc.Enroll(context.Background(), "  ", "c1", "tok")
	assert.True(t, services.IsValidationError(err))

	_, err = f.svc.Enroll(context.Background(), "p1", " ", "tok")
	assert.True(t, services.IsValidationError(err))

	f.courses.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
}

func TestEnroll_CourseNotPublished(t *testing.T) {
	f := newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(nil, repositories.ErrNotFound)
	f.courses.On("ExistsInManagement", mock.Anything, "c1").Return(true, nil)

	_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")

	require.True(t, services.IsValidationError(err))
	assert.Contains(t, err.Error(), "no está disponible")
}

func TestEnroll_CourseDoesNotExist(t *testing.T) {
	f := newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(nil, repositories.ErrNotFound)
	f.courses.On("ExistsInManagement", mock.Anything, "c1").Return(false, nil)

	_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")

	require.True(t, services.IsValidationError(err))
	assert.Equal(t, "El curso no existe", err.Error())
}

func TestEnroll_CourseInfraErrors(t *testing.T) {
	boom := errors.New("contentful caído")

	f := newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(nil, boom)
	_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")
	assert.ErrorIs(t, err, boom)
	assert.False(t, services.IsValidationError(err))

	f = newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(nil, repositories.ErrNotFound)
	f.courses.On("ExistsInManagement", mock.Anything, "c1").Return(false, boom)
	_, err = f.svc.Enroll(context.Background(), "p1", "c1", "tok")
	assert.ErrorIs(t, err, boom)
}

func TestEnroll_ParticipantDoesNotExist(t *testing.T) {
	f := newEnrollFixture(t) // auth-service responde 404
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)

	_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")

	require.True(t, services.IsValidationError(err))
	assert.Equal(t, "El participante no existe", err.Error())
}

func TestEnroll_ParticipantNotAccepted(t *testing.T) {
	f := newEnrollFixture(t, models.AuthParticipant{ParticipantID: "p1", ApprovalStatus: "pending"})
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)

	_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")

	require.True(t, services.IsValidationError(err))
	assert.Equal(t, "El participante no está activo", err.Error())
	f.repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestEnroll_AuthServiceError(t *testing.T) {
	f := newEnrollFixture(t, participanteAprobado)
	f.auth.failGet = true
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)

	_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")

	require.Error(t, err)
	assert.False(t, services.IsValidationError(err), "una caída de auth-service no es un 400")
}

func TestEnroll_AuthServiceNotConfigured(t *testing.T) {
	repo, courses := new(MockEnrollmentRepository), new(MockCourseRepository)
	svc := services.NewEnrollmentService(repo, courses, repositories.NewAuthServiceClient(""))
	courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)

	_, err := svc.Enroll(context.Background(), "p1", "c1", "tok")

	assert.ErrorIs(t, err, repositories.ErrAuthServiceNotConfigured)
}

func TestEnroll_AlreadyEnrolled(t *testing.T) {
	f := newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)
	f.repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{
		{ParticipanteID: "p1", Estado: models.EnrollmentEstadoActiva},
	}, nil)

	_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")

	require.True(t, services.IsValidationError(err))
	assert.Contains(t, err.Error(), "ya se encuentra inscripto")
}

func TestEnroll_InactiveEnrollmentAllowsReEnroll(t *testing.T) {
	f := newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", Capacity: intPtr(1)}, nil)
	// una inscripción previa no activa ni cuenta como duplicado ni ocupa cupo
	f.repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{
		{ParticipanteID: "p1", Estado: "baja"},
	}, nil)
	f.repo.On("Create", mock.Anything, mock.Anything).Return(&models.Enrollment{ID: "e2"}, nil)

	e, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")

	require.NoError(t, err)
	assert.Equal(t, "e2", e.ID)
}

func TestEnroll_CapacityFull(t *testing.T) {
	f := newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", Capacity: intPtr(2)}, nil)
	f.repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{
		{ParticipanteID: "p2", Estado: models.EnrollmentEstadoActiva},
		{ParticipanteID: "p3", Estado: models.EnrollmentEstadoActiva},
	}, nil)

	_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")

	require.True(t, services.IsValidationError(err))
	assert.Equal(t, "El curso no tiene cupo disponible", err.Error())
	f.repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestEnroll_CapacityWithRoomLeft(t *testing.T) {
	f := newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", Capacity: intPtr(2)}, nil)
	f.repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{
		{ParticipanteID: "p2", Estado: models.EnrollmentEstadoActiva},
	}, nil)
	f.repo.On("Create", mock.Anything, mock.Anything).Return(&models.Enrollment{ID: "e3"}, nil)

	_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")

	assert.NoError(t, err)
}

func TestEnroll_RepoErrors(t *testing.T) {
	boom := errors.New("firestore caído")

	f := newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)
	f.repo.On("ListByCourse", mock.Anything, "c1").Return(nil, boom)
	_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")
	assert.ErrorIs(t, err, boom)

	f = newEnrollFixture(t, participanteAprobado)
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)
	f.repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{}, nil)
	f.repo.On("Create", mock.Anything, mock.Anything).Return(nil, boom)
	_, err = f.svc.Enroll(context.Background(), "p1", "c1", "tok")
	assert.ErrorIs(t, err, boom)
}

// auth-service guarda "approved"; "accepted" es el valor que se escribió
// hasta 2026-09-28 y sigue en filas reales. Los dos habilitan a inscribir.
func TestEnroll_AcceptsApprovedAndLegacyAccepted(t *testing.T) {
	for _, status := range []string{"approved", "accepted", ""} {
		f := newEnrollFixture(t, models.AuthParticipant{ParticipantID: "p1", ApprovalStatus: status})
		f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)
		f.repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{}, nil)
		f.repo.On("Create", mock.Anything, mock.Anything).Return(&models.Enrollment{ID: "e1"}, nil)

		_, err := f.svc.Enroll(context.Background(), "p1", "c1", "tok")

		assert.NoError(t, err, "approval_status=%q", status)
	}
}

func TestEnrollmentService_ListByParticipant(t *testing.T) {
	repo := new(MockEnrollmentRepository)
	svc := services.NewEnrollmentService(repo, nil, nil)
	repo.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{{ID: "e1"}}, nil)

	list, err := svc.ListByParticipant(context.Background(), "p1")

	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func summariesByUser(list []models.UserCourseSummary) map[string][]string {
	out := make(map[string][]string, len(list))
	for _, s := range list {
		out[s.UserID] = s.Courses
	}
	return out
}

func TestEnrollmentService_GetBulkUserCourses(t *testing.T) {
	repo, courses := new(MockEnrollmentRepository), new(MockCourseRepository)
	svc := services.NewEnrollmentService(repo, courses, nil)
	ids := []string{"p1", "p2", "t1"}

	repo.On("ListByParticipants", mock.Anything, ids).Return([]models.Enrollment{
		{ParticipanteID: "p1", CursoID: "c1", Estado: models.EnrollmentEstadoActiva},
		{ParticipanteID: "p1", CursoID: "c2", Estado: models.EnrollmentEstadoActiva},
		{ParticipanteID: "p2", CursoID: "c1", Estado: models.EnrollmentEstadoActiva},
		{ParticipanteID: "p2", CursoID: "c3", Estado: "baja"}, // no cuenta
	}, nil)
	courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", Title: "Huerta"}, nil).Once()
	courses.On("Get", mock.Anything, "c2").Return(nil, repositories.ErrNotFound)
	courses.On("ListByResponsibleTeachers", mock.Anything, ids).Return([]models.Course{
		{ID: "c9", Title: "Carpintería", ResponsibleTeacher: "t1"},
	}, nil)

	list, err := svc.GetBulkUserCourses(context.Background(), ids)

	require.NoError(t, err)
	assert.Equal(t, map[string][]string{
		"p1": {"Huerta", "Curso c2"}, // título de c1 cacheado; c2 no se encontró
		"p2": {"Huerta"},
		"t1": {"Carpintería"},
	}, summariesByUser(list))
	courses.AssertNumberOfCalls(t, "Get", 2)
}

func TestEnrollmentService_GetBulkUserCourses_Empty(t *testing.T) {
	svc := services.NewEnrollmentService(new(MockEnrollmentRepository), new(MockCourseRepository), nil)

	list, err := svc.GetBulkUserCourses(context.Background(), nil)

	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestEnrollmentService_GetBulkUserCourses_ToleratesRepoErrors(t *testing.T) {
	repo, courses := new(MockEnrollmentRepository), new(MockCourseRepository)
	svc := services.NewEnrollmentService(repo, courses, nil)
	repo.On("ListByParticipants", mock.Anything, []string{"p1"}).Return(nil, errors.New("firestore caído"))
	courses.On("ListByResponsibleTeachers", mock.Anything, []string{"p1"}).Return(nil, errors.New("contentful caído"))

	list, err := svc.GetBulkUserCourses(context.Background(), []string{"p1"})

	require.NoError(t, err)
	assert.Equal(t, map[string][]string{"p1": {}}, summariesByUser(list))
}
