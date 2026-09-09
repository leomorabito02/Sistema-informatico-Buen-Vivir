package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type MockActivityService struct {
	mock.Mock
}

func (m *MockActivityService) ListByCourse(ctx context.Context, cursoID string) ([]models.Activity, error) {
	args := m.Called(ctx, cursoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Activity), args.Error(1)
}

func (m *MockActivityService) GetActivity(ctx context.Context, id string) (*models.Activity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Activity), args.Error(1)
}

func (m *MockActivityService) CreateActivity(ctx context.Context, cursoID string, input models.ActivityInput) (*models.Activity, error) {
	args := m.Called(ctx, cursoID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Activity), args.Error(1)
}

type MockEnrollmentService struct {
	mock.Mock
}

func (m *MockEnrollmentService) ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error) {
	args := m.Called(ctx, cursoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Enrollment), args.Error(1)
}

type MockAttendanceService struct {
	mock.Mock
}

func (m *MockAttendanceService) ListByActivity(ctx context.Context, actividadID string) ([]models.Attendance, error) {
	args := m.Called(ctx, actividadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Attendance), args.Error(1)
}

func (m *MockAttendanceService) RegisterBatch(ctx context.Context, actividadID string, inputs []models.AttendanceInput) error {
	args := m.Called(ctx, actividadID, inputs)
	return args.Error(0)
}

type MockProgressService struct {
	mock.Mock
}

func (m *MockProgressService) CalculateCourseProgress(ctx context.Context, cursoID string) ([]models.ParticipantProgress, error) {
	args := m.Called(ctx, cursoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ParticipantProgress), args.Error(1)
}

func newDocenteController() (*controllers.DocenteController, *MockCourseService, *MockActivityService, *MockEnrollmentService, *MockAttendanceService, *MockProgressService) {
	courseSvc := new(MockCourseService)
	activitySvc := new(MockActivityService)
	enrollmentSvc := new(MockEnrollmentService)
	attendanceSvc := new(MockAttendanceService)
	progressSvc := new(MockProgressService)
	ctrl := controllers.NewDocenteController(courseSvc, activitySvc, enrollmentSvc, attendanceSvc, progressSvc)
	return ctrl, courseSvc, activitySvc, enrollmentSvc, attendanceSvc, progressSvc
}

func withTeacherContext(r *http.Request, teacherID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), middlewares.UserIDKey, teacherID))
}

func TestDocenteController_ListMyCourses_Success(t *testing.T) {
	ctrl, courseSvc, _, _, _, _ := newDocenteController()

	courseSvc.On("ListCoursesByTeacher", mock.Anything, "teacher-1").Return([]models.Course{{ID: "c1"}}, nil)

	req := withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos", nil), "teacher-1")
	w := httptest.NewRecorder()
	ctrl.ListMyCourses(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocenteController_ListMyDraftCourses_Success(t *testing.T) {
	ctrl, courseSvc, _, _, _, _ := newDocenteController()

	courseSvc.On("ListDraftCoursesByTeacher", mock.Anything, "teacher-1").Return([]models.Course{{ID: "draft-1"}}, nil)

	req := withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/borradores", nil), "teacher-1")
	w := httptest.NewRecorder()
	ctrl.ListMyDraftCourses(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocenteController_CreateMyCourse_BadJSON(t *testing.T) {
	ctrl, _, _, _, _, _ := newDocenteController()

	req := withTeacherContext(httptest.NewRequest("POST", "/api/docente/cursos", bytes.NewBufferString("{invalid")), "teacher-1")
	w := httptest.NewRecorder()
	ctrl.CreateMyCourse(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDocenteController_CreateMyCourse_Success(t *testing.T) {
	ctrl, courseSvc, _, _, _, _ := newDocenteController()

	courseSvc.On("CreateCourseAsTeacher", mock.Anything, "teacher-1", mock.AnythingOfType("models.CourseInput")).
		Return(&models.Course{ID: "c1", ResponsibleTeacher: "teacher-1"}, nil)

	body, _ := json.Marshal(models.CourseInput{Title: "Curso"})
	req := withTeacherContext(httptest.NewRequest("POST", "/api/docente/cursos", bytes.NewBuffer(body)), "teacher-1")
	w := httptest.NewRecorder()
	ctrl.CreateMyCourse(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestDocenteController_UpdateMyCourse_ServiceError(t *testing.T) {
	ctrl, courseSvc, _, _, _, _ := newDocenteController()

	courseSvc.On("UpdateCourseAsTeacher", mock.Anything, "c1", "teacher-1", mock.AnythingOfType("models.CourseInput")).
		Return(nil, services.NewForbiddenError("no sos el docente responsable de este curso"))

	body, _ := json.Marshal(models.CourseInput{Title: "Curso"})
	req := withChiURLParam(withTeacherContext(httptest.NewRequest("PUT", "/api/docente/cursos/c1", bytes.NewBuffer(body)), "teacher-1"), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.UpdateMyCourse(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDocenteController_ListActivities_NotOwner(t *testing.T) {
	ctrl, courseSvc, _, _, _, _ := newDocenteController()

	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").
		Return(services.NewForbiddenError("no sos el docente responsable de este curso"))

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/c1/actividades", nil), "teacher-1"), "cursoId", "c1")
	w := httptest.NewRecorder()
	ctrl.ListActivities(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDocenteController_ListActivities_Success(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()

	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)
	activitySvc.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{{ID: "a1"}}, nil)

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/c1/actividades", nil), "teacher-1"), "cursoId", "c1")
	w := httptest.NewRecorder()
	ctrl.ListActivities(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocenteController_CreateActivity_BadJSON(t *testing.T) {
	ctrl, courseSvc, _, _, _, _ := newDocenteController()

	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("POST", "/api/docente/cursos/c1/actividades", bytes.NewBufferString("{invalid")), "teacher-1"), "cursoId", "c1")
	w := httptest.NewRecorder()
	ctrl.CreateActivity(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDocenteController_CreateActivity_Success(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()

	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)
	activitySvc.On("CreateActivity", mock.Anything, "c1", mock.AnythingOfType("models.ActivityInput")).
		Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)

	body, _ := json.Marshal(models.ActivityInput{Titulo: "Práctica 1"})
	req := withChiURLParam(withTeacherContext(httptest.NewRequest("POST", "/api/docente/cursos/c1/actividades", bytes.NewBuffer(body)), "teacher-1"), "cursoId", "c1")
	w := httptest.NewRecorder()
	ctrl.CreateActivity(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestDocenteController_RegisterAttendance_ActivityNotFound(t *testing.T) {
	ctrl, _, activitySvc, _, _, _ := newDocenteController()

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(nil, repoErrNotFound())

	body, _ := json.Marshal([]models.AttendanceInput{{ParticipanteID: "p1", Presente: true}})
	req := withChiURLParam(withTeacherContext(httptest.NewRequest("POST", "/api/docente/actividades/a1/asistencia", bytes.NewBuffer(body)), "teacher-1"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.RegisterAttendance(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDocenteController_RegisterAttendance_Success(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, attendanceSvc, _ := newDocenteController()

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)
	attendanceSvc.On("RegisterBatch", mock.Anything, "a1", mock.AnythingOfType("[]models.AttendanceInput")).Return(nil)

	body, _ := json.Marshal([]models.AttendanceInput{{ParticipanteID: "p1", Presente: true}})
	req := withChiURLParam(withTeacherContext(httptest.NewRequest("POST", "/api/docente/actividades/a1/asistencia", bytes.NewBuffer(body)), "teacher-1"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.RegisterAttendance(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocenteController_ListAttendance_Success(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, attendanceSvc, _ := newDocenteController()

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)
	attendanceSvc.On("ListByActivity", mock.Anything, "a1").Return([]models.Attendance{{ID: "att1"}}, nil)

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/actividades/a1/asistencia", nil), "teacher-1"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.ListAttendance(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocenteController_ListEnrollments_Success(t *testing.T) {
	ctrl, courseSvc, _, enrollmentSvc, _, _ := newDocenteController()

	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)
	enrollmentSvc.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{{ID: "e1"}}, nil)

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/c1/inscriptos", nil), "teacher-1"), "cursoId", "c1")
	w := httptest.NewRecorder()
	ctrl.ListEnrollments(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocenteController_GetProgress_Success(t *testing.T) {
	ctrl, courseSvc, _, _, _, progressSvc := newDocenteController()

	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)
	progressSvc.On("CalculateCourseProgress", mock.Anything, "c1").
		Return([]models.ParticipantProgress{{ParticipanteID: "p1"}}, nil)

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/c1/progreso", nil), "teacher-1"), "cursoId", "c1")
	w := httptest.NewRecorder()
	ctrl.GetProgress(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocenteController_GetProgress_NotOwner(t *testing.T) {
	ctrl, courseSvc, _, _, _, _ := newDocenteController()

	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").
		Return(services.NewForbiddenError("no sos el docente responsable de este curso"))

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/c1/progreso", nil), "teacher-1"), "cursoId", "c1")
	w := httptest.NewRecorder()
	ctrl.GetProgress(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
