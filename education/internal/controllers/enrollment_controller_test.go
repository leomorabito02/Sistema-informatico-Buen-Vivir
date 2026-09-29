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
