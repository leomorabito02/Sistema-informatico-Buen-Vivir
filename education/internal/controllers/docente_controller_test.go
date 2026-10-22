package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func (m *MockActivityService) UpdateActivity(ctx context.Context, id string, input models.ActivityInput) (*models.Activity, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Activity), args.Error(1)
}

func (m *MockActivityService) DeleteActivity(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
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

func (m *MockEnrollmentService) ListByParticipant(ctx context.Context, participantID string) ([]models.Enrollment, error) {
	args := m.Called(ctx, participantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Enrollment), args.Error(1)
}

func (m *MockEnrollmentService) Enroll(ctx context.Context, participanteID, cursoID, bearerToken string) (*models.Enrollment, error) {
	args := m.Called(ctx, participanteID, cursoID, bearerToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Enrollment), args.Error(1)
}

func (m *MockEnrollmentService) GetBulkUserCourses(ctx context.Context, userIDs []string) ([]models.UserCourseSummary, error) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserCourseSummary), args.Error(1)
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
	ctrl, courseSvc, activitySvc, enrollmentSvc, attendanceSvc, progressSvc, _ := newDocenteControllerWithSubmissions()
	return ctrl, courseSvc, activitySvc, enrollmentSvc, attendanceSvc, progressSvc
}

func newDocenteControllerWithSubmissions() (*controllers.DocenteController, *MockCourseService, *MockActivityService, *MockEnrollmentService, *MockAttendanceService, *MockProgressService, *MockSubmissionService) {
	courseSvc := new(MockCourseService)
	activitySvc := new(MockActivityService)
	enrollmentSvc := new(MockEnrollmentService)
	attendanceSvc := new(MockAttendanceService)
	progressSvc := new(MockProgressService)
	submissionSvc := new(MockSubmissionService)
	ctrl := controllers.NewDocenteController(courseSvc, activitySvc, enrollmentSvc, attendanceSvc, progressSvc, submissionSvc)
	return ctrl, courseSvc, activitySvc, enrollmentSvc, attendanceSvc, progressSvc, submissionSvc
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

// CP-16: un docente que pide la asistencia de una actividad de un curso
// ajeno recibe 403 y no se consulta la asistencia.
func TestDocenteController_ListAttendance_NotOwnerIsForbidden(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, attendanceSvc, _ := newDocenteController()

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-2").Return(services.NewForbiddenError("no sos el docente responsable de este curso"))

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/actividades/a1/asistencia", nil), "teacher-2"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.ListAttendance(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	attendanceSvc.AssertNotCalled(t, "ListByActivity", mock.Anything, mock.Anything)
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

func TestDocenteController_UpdateActivity_Success(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)
	activitySvc.On("UpdateActivity", mock.Anything, "a1", mock.AnythingOfType("models.ActivityInput")).
		Return(&models.Activity{ID: "a1"}, nil)

	body, _ := json.Marshal(models.ActivityInput{Titulo: "Práctica", Consigna: "Hacer"})
	req := withChiURLParam(withTeacherContext(httptest.NewRequest("PUT", "/api/docente/actividades/a1", bytes.NewBuffer(body)), "teacher-1"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.UpdateActivity(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDocenteController_UpdateActivity_NotOwner(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(services.NewForbiddenError("no"))

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("PUT", "/api/docente/actividades/a1", bytes.NewBufferString("{}")), "teacher-1"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.UpdateActivity(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	activitySvc.AssertNotCalled(t, "UpdateActivity", mock.Anything, mock.Anything, mock.Anything)
}

func TestDocenteController_UpdateActivity_BadJSON(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("PUT", "/api/docente/actividades/a1", bytes.NewBufferString("{")), "teacher-1"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.UpdateActivity(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDocenteController_DeleteActivity_Success(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)
	activitySvc.On("DeleteActivity", mock.Anything, "a1").Return(nil)

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("DELETE", "/api/docente/actividades/a1", nil), "teacher-1"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.DeleteActivity(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDocenteController_DeleteActivity_NotFound(t *testing.T) {
	ctrl, _, activitySvc, _, _, _ := newDocenteController()

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(nil, repoErrNotFound())

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("DELETE", "/api/docente/actividades/a1", nil), "teacher-1"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.DeleteActivity(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDocenteController_ListSubmissions_Success(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, _, _, submissionSvc := newDocenteControllerWithSubmissions()

	activity := &models.Activity{ID: "a1", CursoID: "c1"}
	activitySvc.On("GetActivity", mock.Anything, "a1").Return(activity, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)
	submissionSvc.On("ListByActivity", mock.Anything, activity, "").
		Return([]models.SubmissionStatusRow{{ParticipanteID: "p1", Estado: models.SubmissionEstadoSinEntregar}}, nil)

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/actividades/a1/entregas", nil), "teacher-1"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.ListSubmissions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sin_entregar")
}

func TestDocenteController_ListSubmissions_NotOwner(t *testing.T) {
	ctrl, courseSvc, activitySvc, _, _, _, submissionSvc := newDocenteControllerWithSubmissions()

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(services.NewForbiddenError("no"))

	req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/actividades/a1/entregas", nil), "teacher-1"), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.ListSubmissions(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	submissionSvc.AssertNotCalled(t, "ListByActivity", mock.Anything, mock.Anything, mock.Anything)
}

func TestAttendanceController_List(t *testing.T) {
	activitySvc := new(MockActivityService)
	attendanceSvc := new(MockAttendanceService)
	ctrl := controllers.NewAttendanceController(activitySvc, attendanceSvc)

	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1"}, nil)
	activitySvc.On("GetActivity", mock.Anything, "nope").Return(nil, repoErrNotFound())
	attendanceSvc.On("ListByActivity", mock.Anything, "a1").Return([]models.Attendance{{ID: "a1_p1", Presente: true}}, nil)

	w := httptest.NewRecorder()
	ctrl.List(w, withChiURLParam(httptest.NewRequest("GET", "/api/actividades/a1/asistencia", nil), "actividadId", "a1"))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	ctrl.List(w, withChiURLParam(httptest.NewRequest("GET", "/api/actividades/nope/asistencia", nil), "actividadId", "nope"))
	assert.Equal(t, http.StatusNotFound, w.Code)

	activitySvc.On("GetActivity", mock.Anything, "err").Return(&models.Activity{ID: "err"}, nil)
	attendanceSvc.On("ListByActivity", mock.Anything, "err").Return(nil, errors.New("db error"))
	w = httptest.NewRecorder()
	ctrl.List(w, withChiURLParam(httptest.NewRequest("GET", "/api/actividades/err/asistencia", nil), "actividadId", "err"))
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestDocenteController_ErrorBranches(t *testing.T) {
	t.Run("ListMyCourses service error maps to 502", func(t *testing.T) {
		ctrl, courseSvc, _, _, _, _ := newDocenteController()
		courseSvc.On("ListCoursesByTeacher", mock.Anything, "t1").Return(nil, errors.New("err"))

		req := withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos", nil), "t1")
		w := httptest.NewRecorder()
		ctrl.ListMyCourses(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("ListMyDraftCourses service error maps to 502", func(t *testing.T) {
		ctrl, courseSvc, _, _, _, _ := newDocenteController()
		courseSvc.On("ListDraftCoursesByTeacher", mock.Anything, "t1").Return(nil, errors.New("err"))

		req := withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/borradores", nil), "t1")
		w := httptest.NewRecorder()
		ctrl.ListMyDraftCourses(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("CreateMyCourse service error maps to 502", func(t *testing.T) {
		ctrl, courseSvc, _, _, _, _ := newDocenteController()
		courseSvc.On("CreateCourseAsTeacher", mock.Anything, "t1", mock.Anything).Return(nil, errors.New("err"))

		body, _ := json.Marshal(models.CourseInput{Title: "Curso"})
		req := withTeacherContext(httptest.NewRequest("POST", "/api/docente/cursos", bytes.NewBuffer(body)), "t1")
		w := httptest.NewRecorder()
		ctrl.CreateMyCourse(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("UpdateMyCourse service error maps to 502", func(t *testing.T) {
		ctrl, courseSvc, _, _, _, _ := newDocenteController()
		courseSvc.On("UpdateCourseAsTeacher", mock.Anything, "c1", "t1", mock.Anything).Return(nil, errors.New("err"))

		body, _ := json.Marshal(models.CourseInput{Title: "Curso"})
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("PUT", "/api/docente/cursos/c1", bytes.NewBuffer(body)), "t1"), "id", "c1")
		w := httptest.NewRecorder()
		ctrl.UpdateMyCourse(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("ListActivities service error maps to 502", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(nil)
		activitySvc.On("ListByCourse", mock.Anything, "c1").Return(nil, errors.New("err"))

		req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/c1/actividades", nil), "t1"), "cursoId", "c1")
		w := httptest.NewRecorder()
		ctrl.ListActivities(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("ListAttendance service error maps to 502", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, _, attendanceSvc, _ := newDocenteController()
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(nil)
		attendanceSvc.On("ListByActivity", mock.Anything, "a1").Return(nil, errors.New("err"))

		req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/actividades/a1/asistencia", nil), "t1"), "actividadId", "a1")
		w := httptest.NewRecorder()
		ctrl.ListAttendance(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("ListEnrollments service error maps to 502", func(t *testing.T) {
		ctrl, courseSvc, _, enrollmentSvc, _, _ := newDocenteController()
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(nil)
		enrollmentSvc.On("ListByCourse", mock.Anything, "c1").Return(nil, errors.New("err"))

		req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/c1/inscriptos", nil), "t1"), "cursoId", "c1")
		w := httptest.NewRecorder()
		ctrl.ListEnrollments(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("GetProgress service error maps to 502", func(t *testing.T) {
		ctrl, courseSvc, _, _, _, progressSvc := newDocenteController()
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(nil)
		progressSvc.On("CalculateCourseProgress", mock.Anything, "c1").Return(nil, errors.New("err"))

		req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/c1/progreso", nil), "t1"), "cursoId", "c1")
		w := httptest.NewRecorder()
		ctrl.GetProgress(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("UpdateMyCourse bad JSON returns 400", func(t *testing.T) {
		ctrl, _, _, _, _, _ := newDocenteController()
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("PUT", "/api/docente/cursos/c1", bytes.NewBufferString("{bad")), "t1"), "id", "c1")
		w := httptest.NewRecorder()
		ctrl.UpdateMyCourse(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateActivity bad JSON returns 400", func(t *testing.T) {
		ctrl, courseSvc, _, _, _, _ := newDocenteController()
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(nil)
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("POST", "/api/docente/cursos/c1/actividades", bytes.NewBufferString("{bad")), "t1"), "cursoId", "c1")
		w := httptest.NewRecorder()
		ctrl.CreateActivity(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateActivity service error returns 502", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(nil)
		activitySvc.On("CreateActivity", mock.Anything, "c1", mock.Anything).Return(nil, errors.New("err"))
		body, _ := json.Marshal(models.ActivityInput{Titulo: "Clase 1"})
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("POST", "/api/docente/cursos/c1/actividades", bytes.NewBuffer(body)), "t1"), "cursoId", "c1")
		w := httptest.NewRecorder()
		ctrl.CreateActivity(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("UpdateActivity bad JSON returns 400", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(nil)
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("PUT", "/api/docente/actividades/a1", bytes.NewBufferString("{bad")), "t1"), "actividadId", "a1")
		w := httptest.NewRecorder()
		ctrl.UpdateActivity(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateActivity service error returns 502", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(nil)
		activitySvc.On("UpdateActivity", mock.Anything, "a1", mock.Anything).Return(nil, errors.New("err"))
		body, _ := json.Marshal(models.ActivityInput{Titulo: "Clase 1"})
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("PUT", "/api/docente/actividades/a1", bytes.NewBuffer(body)), "t1"), "actividadId", "a1")
		w := httptest.NewRecorder()
		ctrl.UpdateActivity(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("DeleteActivity service error returns 502", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(nil)
		activitySvc.On("DeleteActivity", mock.Anything, "a1").Return(errors.New("err"))
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("DELETE", "/api/docente/actividades/a1", nil), "t1"), "actividadId", "a1")
		w := httptest.NewRecorder()
		ctrl.DeleteActivity(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("ListSubmissions service error returns 502", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, _, _, _, submissionSvc := newDocenteControllerWithSubmissions()
		activity := &models.Activity{ID: "a1", CursoID: "c1"}
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(activity, nil)
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(nil)
		submissionSvc.On("ListByActivity", mock.Anything, activity, mock.Anything).Return(nil, errors.New("err"))

		req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/actividades/a1/entregas", nil), "t1"), "actividadId", "a1")
		w := httptest.NewRecorder()
		ctrl.ListSubmissions(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("CreateActivity ownership error returns 403", func(t *testing.T) {
		// Arrange
		ctrl, courseSvc, _, _, _, _ := newDocenteController()
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(services.NewForbiddenError("not owner"))
		body, _ := json.Marshal(models.ActivityInput{Titulo: "Clase 1"})
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("POST", "/api/docente/cursos/c1/actividades", bytes.NewBuffer(body)), "t1"), "cursoId", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.CreateActivity(w, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("ListAttendance ownership error returns 403", func(t *testing.T) {
		// Arrange
		ctrl, courseSvc, activitySvc, _, _, _ := newDocenteController()
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(services.NewForbiddenError("not owner"))
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/actividades/a1/asistencia", nil), "t1"), "actividadId", "a1")
		w := httptest.NewRecorder()

		// Act
		ctrl.ListAttendance(w, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("ListEnrollments ownership error returns 403", func(t *testing.T) {
		// Arrange
		ctrl, courseSvc, _, _, _, _ := newDocenteController()
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "t1").Return(services.NewForbiddenError("not owner"))
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("GET", "/api/docente/cursos/c1/inscriptos", nil), "t1"), "cursoId", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.ListEnrollments(w, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("UpdateMyCourse success returns 200", func(t *testing.T) {
		// Arrange
		ctrl, courseSvc, _, _, _, _ := newDocenteController()
		expected := &models.Course{ID: "c1", Title: "Updated Title"}
		courseSvc.On("UpdateCourseAsTeacher", mock.Anything, "c1", "t1", mock.Anything).Return(expected, nil)

		body, _ := json.Marshal(models.CourseInput{Title: "Updated Title"})
		req := withChiURLParam(withTeacherContext(httptest.NewRequest("PUT", "/api/docente/cursos/c1", bytes.NewBuffer(body)), "t1"), "id", "c1")
		w := httptest.NewRecorder()

		// Act
		ctrl.UpdateMyCourse(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

