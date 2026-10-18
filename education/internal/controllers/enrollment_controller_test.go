package controllers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

func TestEnrollmentController_Enroll(t *testing.T) {
	svc := new(MockEnrollmentService)
	ctrl := controllers.NewEnrollmentController(svc)
	svc.On("Enroll", mock.Anything, "p1", "c1", "tok").Return(&models.Enrollment{ID: "e1"}, nil)
	svc.On("Enroll", mock.Anything, "p2", "c1", "tok").Return(nil, services.NewValidationError("El participante ya se encuentra inscripto en el curso"))

	req := func(body string) *http.Request {
		return withChiURLParams(withAuth(httptest.NewRequest("POST", "/api/cursos/c1/inscripciones", strings.NewReader(body)), "u", "role_volunteer", false), "id", "c1")
	}

	w := httptest.NewRecorder()
	ctrl.Enroll(w, req(`{"participanteId":"p1"}`))
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	ctrl.Enroll(w, req(`{"participanteId":"p2"}`))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ya se encuentra inscripto")

	w = httptest.NewRecorder()
	ctrl.Enroll(w, req(`{`))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnrollmentController_List(t *testing.T) {
	svc := new(MockEnrollmentService)
	ctrl := controllers.NewEnrollmentController(svc)
	svc.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{{ID: "e1"}}, nil)

	w := httptest.NewRecorder()
	ctrl.List(w, withChiURLParams(httptest.NewRequest("GET", "/api/cursos/c1/inscripciones", nil), "id", "c1"))
	assert.Equal(t, http.StatusOK, w.Code)
}

type MockGeneralAttendanceService struct {
	mock.Mock
}

func (m *MockGeneralAttendanceService) ListByDate(ctx context.Context, fecha, bearerToken string) ([]models.GeneralAttendanceView, error) {
	args := m.Called(ctx, fecha, bearerToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GeneralAttendanceView), args.Error(1)
}

func (m *MockGeneralAttendanceService) RegisterBatch(ctx context.Context, fecha string, inputs []models.GeneralAttendanceInput, bearerToken string) error {
	return m.Called(ctx, fecha, inputs, bearerToken).Error(0)
}

func TestGeneralAttendanceController_ListAndRegister(t *testing.T) {
	svc := new(MockGeneralAttendanceService)
	ctrl := controllers.NewGeneralAttendanceController(svc)
	svc.On("ListByDate", mock.Anything, "2026-10-05", "tok").Return([]models.GeneralAttendanceView{{ParticipanteID: "p1"}}, nil)
	svc.On("ListByDate", mock.Anything, "mal", "tok").Return(nil, services.NewValidationError("fecha inválida"))
	svc.On("RegisterBatch", mock.Anything, "2026-10-05", mock.Anything, "tok").Return(nil)

	w := httptest.NewRecorder()
	ctrl.List(w, withAuth(httptest.NewRequest("GET", "/api/asistencia-general?fecha=2026-10-05", nil), "u", "role_admin", false))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	ctrl.List(w, withAuth(httptest.NewRequest("GET", "/api/asistencia-general?fecha=mal", nil), "u", "role_admin", false))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	ctrl.Register(w, withAuth(httptest.NewRequest("POST", "/api/asistencia-general?fecha=2026-10-05", strings.NewReader(`[{"participanteId":"p1","presente":true}]`)), "u", "role_admin", false))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	ctrl.Register(w, withAuth(httptest.NewRequest("POST", "/api/asistencia-general?fecha=2026-10-05", strings.NewReader(`[`)), "u", "role_admin", false))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnrollmentController_DecisionTable(t *testing.T) {
	t.Run("Enroll unexpected service error returns 500", func(t *testing.T) {
		// Arrange
		svc := new(MockEnrollmentService)
		ctrl := controllers.NewEnrollmentController(svc)
		svc.On("Enroll", mock.Anything, "p1", "c1", "tok").Return(nil, assert.AnError)
		req := withChiURLParams(withAuth(httptest.NewRequest("POST", "/api/cursos/c1/inscripciones", strings.NewReader(`{"participanteId":"p1"}`)), "u", "role_volunteer", false), "id", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.Enroll(w, req)

		// Assert
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("List with service error returns 500", func(t *testing.T) {
		// Arrange
		svc := new(MockEnrollmentService)
		ctrl := controllers.NewEnrollmentController(svc)
		svc.On("ListByCourse", mock.Anything, "c1").Return(nil, assert.AnError)
		req := withChiURLParams(httptest.NewRequest("GET", "/api/cursos/c1/inscripciones", nil), "id", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.List(w, req)

		// Assert
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("ListByParticipant success returns 200", func(t *testing.T) {
		// Arrange
		svc := new(MockEnrollmentService)
		ctrl := controllers.NewEnrollmentController(svc)
		svc.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{{ID: "e1"}}, nil)
		req := withChiURLParams(httptest.NewRequest("GET", "/api/participantes/p1/inscripciones", nil), "id", "p1")
		w := httptest.NewRecorder()

		// Act
		ctrl.ListByParticipant(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ListByParticipant service error returns 500", func(t *testing.T) {
		// Arrange
		svc := new(MockEnrollmentService)
		ctrl := controllers.NewEnrollmentController(svc)
		svc.On("ListByParticipant", mock.Anything, "p1").Return(nil, assert.AnError)
		req := withChiURLParams(httptest.NewRequest("GET", "/api/participantes/p1/inscripciones", nil), "id", "p1")
		w := httptest.NewRecorder()

		// Act
		ctrl.ListByParticipant(w, req)

		// Assert
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("GetBulkUserCourses invalid JSON returns 400", func(t *testing.T) {
		// Arrange
		svc := new(MockEnrollmentService)
		ctrl := controllers.NewEnrollmentController(svc)
		req := httptest.NewRequest("POST", "/api/bulk-cursos-usuarios", strings.NewReader("{bad"))
		w := httptest.NewRecorder()

		// Act
		ctrl.GetBulkUserCourses(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetBulkUserCourses success returns 200", func(t *testing.T) {
		// Arrange
		svc := new(MockEnrollmentService)
		ctrl := controllers.NewEnrollmentController(svc)
		svc.On("GetBulkUserCourses", mock.Anything, []string{"u1"}).Return([]models.UserCourseSummary{{UserID: "u1"}}, nil)
		req := httptest.NewRequest("POST", "/api/bulk-cursos-usuarios", strings.NewReader(`{"user_ids":["u1"]}`))
		w := httptest.NewRecorder()

		// Act
		ctrl.GetBulkUserCourses(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetBulkUserCourses service error returns 500", func(t *testing.T) {
		// Arrange
		svc := new(MockEnrollmentService)
		ctrl := controllers.NewEnrollmentController(svc)
		svc.On("GetBulkUserCourses", mock.Anything, []string{"u1"}).Return(nil, assert.AnError)
		req := httptest.NewRequest("POST", "/api/bulk-cursos-usuarios", strings.NewReader(`{"user_ids":["u1"]}`))
		w := httptest.NewRecorder()

		// Act
		ctrl.GetBulkUserCourses(w, req)

		// Assert
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})
}

func TestGeneralAttendanceController_Register_ServiceError(t *testing.T) {
	// Arrange
	svc := new(MockGeneralAttendanceService)
	ctrl := controllers.NewGeneralAttendanceController(svc)
	svc.On("RegisterBatch", mock.Anything, "2026-10-05", mock.Anything, "tok").Return(assert.AnError)
	req := withAuth(httptest.NewRequest("POST", "/api/asistencia-general?fecha=2026-10-05", strings.NewReader(`[{"participanteId":"p1","presente":true}]`)), "u", "role_admin", false)
	w := httptest.NewRecorder()

	// Act
	ctrl.Register(w, req)

	// Assert
	assert.Equal(t, http.StatusBadGateway, w.Code)
}
