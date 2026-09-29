package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type MockSubmissionService struct {
	mock.Mock
}

func (m *MockSubmissionService) ResolveParticipantID(ctx context.Context, bearerToken string) (string, error) {
	args := m.Called(ctx, bearerToken)
	return args.String(0), args.Error(1)
}

func (m *MockSubmissionService) Submit(ctx context.Context, actividadID, participanteID, texto string, uploads []models.SubmissionUpload) (*models.Submission, error) {
	args := m.Called(ctx, actividadID, participanteID, texto, uploads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Submission), args.Error(1)
}

func (m *MockSubmissionService) ListByActivity(ctx context.Context, activity *models.Activity, bearerToken string) ([]models.SubmissionStatusRow, error) {
	args := m.Called(ctx, activity, bearerToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SubmissionStatusRow), args.Error(1)
}

func (m *MockSubmissionService) GetSubmission(ctx context.Context, id string) (*models.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Submission), args.Error(1)
}

func (m *MockSubmissionService) OpenFile(ctx context.Context, sub *models.Submission, index int) (models.SubmissionFile, io.ReadCloser, error) {
	args := m.Called(ctx, sub, index)
	if args.Get(1) == nil {
		return models.SubmissionFile{}, nil, args.Error(2)
	}
	return args.Get(0).(models.SubmissionFile), args.Get(1).(io.ReadCloser), args.Error(2)
}

func (m *MockSubmissionService) ListParticipantCourses(ctx context.Context, participanteID string) ([]models.Course, error) {
	args := m.Called(ctx, participanteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *MockSubmissionService) GetParticipantCourse(ctx context.Context, participanteID, cursoID string) (*models.ParticipantCourseDetail, error) {
	args := m.Called(ctx, participanteID, cursoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ParticipantCourseDetail), args.Error(1)
}

func (m *MockSubmissionService) VerifyParticipantEnrolled(ctx context.Context, participanteID, cursoID string) error {
	return m.Called(ctx, participanteID, cursoID).Error(0)
}

func newSubmissionController() (*controllers.SubmissionController, *MockCourseService, *MockActivityService, *MockSubmissionService) {
	courseSvc := new(MockCourseService)
	activitySvc := new(MockActivityService)
	submissionSvc := new(MockSubmissionService)
	return controllers.NewSubmissionController(courseSvc, activitySvc, submissionSvc), courseSvc, activitySvc, submissionSvc
}

// withAuth simula lo que deja AuthMiddleware en el contexto.
func withAuth(r *http.Request, userID, role string, external bool) *http.Request {
	ctx := context.WithValue(r.Context(), middlewares.UserIDKey, userID)
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, role)
	ctx = context.WithValue(ctx, middlewares.IsExternalKey, external)
	ctx = context.WithValue(ctx, middlewares.TokenKey, "tok")
	return r.WithContext(ctx)
}

func withChiURLParams(r *http.Request, kv ...string) *http.Request {
	rctx := chi.NewRouteContext()
	for i := 0; i+1 < len(kv); i += 2 {
		rctx.URLParams.Add(kv[i], kv[i+1])
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestSubmissionController_ListMyCourses(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("ResolveParticipantID", mock.Anything, "tok").Return("p1", nil)
	svc.On("ListParticipantCourses", mock.Anything, "p1").Return([]models.Course{{ID: "c1"}}, nil)

	w := httptest.NewRecorder()
	ctrl.ListMyCourses(w, withAuth(httptest.NewRequest("GET", "/api/participante/cursos", nil), "u1", "role_participant", false))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSubmissionController_ListMyCourses_NoParticipantProfile(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("ResolveParticipantID", mock.Anything, "tok").Return("", services.NewForbiddenError("sin perfil"))

	w := httptest.NewRecorder()
	ctrl.ListMyCourses(w, withAuth(httptest.NewRequest("GET", "/api/participante/cursos", nil), "u1", "role_participant", false))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSubmissionController_GetMyCourse(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("ResolveParticipantID", mock.Anything, "tok").Return("p1", nil)
	svc.On("GetParticipantCourse", mock.Anything, "p1", "c1").Return(&models.ParticipantCourseDetail{Curso: models.Course{ID: "c1"}}, nil)

	req := withChiURLParams(withAuth(httptest.NewRequest("GET", "/api/participante/cursos/c1", nil), "u1", "role_participant", false), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.GetMyCourse(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSubmissionController_GetMyCourse_NotEnrolled(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("ResolveParticipantID", mock.Anything, "tok").Return("p1", nil)
	svc.On("GetParticipantCourse", mock.Anything, "p1", "c1").Return(nil, services.NewForbiddenError("no estás inscripto en este curso"))

	req := withChiURLParams(withAuth(httptest.NewRequest("GET", "/api/participante/cursos/c1", nil), "u1", "role_participant", false), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.GetMyCourse(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func multipartSubmission(t *testing.T, texto string, files map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	require.NoError(t, mw.WriteField("texto", texto))
	for name, content := range files {
		fw, err := mw.CreateFormFile("archivos", name)
		require.NoError(t, err)
		_, _ = fw.Write([]byte(content))
	}
	require.NoError(t, mw.Close())
	return body, mw.FormDataContentType()
}

func TestSubmissionController_Submit_ParsesMultipart(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("ResolveParticipantID", mock.Anything, "tok").Return("p1", nil)
	svc.On("Submit", mock.Anything, "a1", "p1", "mi respuesta", mock.MatchedBy(func(u []models.SubmissionUpload) bool {
		// CreateFormFile manda application/octet-stream: se completa
		// por extensión.
		return len(u) == 1 && u[0].Nombre == "tp.docx" &&
			u[0].Tipo == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" &&
			string(u[0].Data) == "contenido"
	})).Return(&models.Submission{ID: "a1_p1"}, nil)

	body, contentType := multipartSubmission(t, "mi respuesta", map[string]string{"tp.docx": "contenido"})
	req := httptest.NewRequest("POST", "/api/participante/actividades/a1/entrega", body)
	req.Header.Set("Content-Type", contentType)
	req = withChiURLParams(withAuth(req, "u1", "role_participant", false), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.Submit(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestSubmissionController_Submit_NotMultipart(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("ResolveParticipantID", mock.Anything, "tok").Return("p1", nil)

	req := withChiURLParams(withAuth(httptest.NewRequest("POST", "/api/participante/actividades/a1/entrega", strings.NewReader("{}")), "u1", "role_participant", false), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.Submit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmissionController_Submit_AlreadySubmitted(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("ResolveParticipantID", mock.Anything, "tok").Return("p1", nil)
	svc.On("Submit", mock.Anything, "a1", "p1", "", mock.Anything).Return(nil, services.NewValidationError("Ya entregaste esta actividad"))

	body, contentType := multipartSubmission(t, "", nil)
	req := httptest.NewRequest("POST", "/api/participante/actividades/a1/entrega", body)
	req.Header.Set("Content-Type", contentType)
	req = withChiURLParams(withAuth(req, "u1", "role_participant", false), "actividadId", "a1")
	w := httptest.NewRecorder()
	ctrl.Submit(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Ya entregaste esta actividad", resp["error"])
}

var entregaConArchivo = &models.Submission{
	ID: "a1_p1", ParticipanteID: "p1", ActividadID: "a1",
	Archivos: []models.SubmissionFile{{Nombre: "tp final.pdf", Tipo: "application/pdf", Ruta: "entregas/a1/p1/x"}},
}

func downloadRequest(userID, role string, external bool) *http.Request {
	req := httptest.NewRequest("GET", "/api/entregas/a1_p1/archivos/0", nil)
	return withChiURLParams(withAuth(req, userID, role, external), "entregaId", "a1_p1", "indice", "0")
}

func expectDownload(svc *MockSubmissionService) {
	svc.On("OpenFile", mock.Anything, entregaConArchivo, 0).
		Return(entregaConArchivo.Archivos[0], io.NopCloser(strings.NewReader("%PDF")), nil)
}

func TestSubmissionController_DownloadFile_Volunteer(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("GetSubmission", mock.Anything, "a1_p1").Return(entregaConArchivo, nil)
	expectDownload(svc)

	w := httptest.NewRecorder()
	ctrl.DownloadFile(w, downloadRequest("u9", "role_volunteer", false))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "%PDF", w.Body.String())
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	assert.Contains(t, w.Header().Get("Content-Disposition"), "tp final.pdf")
}

func TestSubmissionController_DownloadFile_OwnerParticipant(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("GetSubmission", mock.Anything, "a1_p1").Return(entregaConArchivo, nil)
	svc.On("ResolveParticipantID", mock.Anything, "tok").Return("p1", nil)
	expectDownload(svc)

	w := httptest.NewRecorder()
	ctrl.DownloadFile(w, downloadRequest("u1", "role_participant", false))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSubmissionController_DownloadFile_OtherParticipantForbidden(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("GetSubmission", mock.Anything, "a1_p1").Return(entregaConArchivo, nil)
	svc.On("ResolveParticipantID", mock.Anything, "tok").Return("p2", nil)

	w := httptest.NewRecorder()
	ctrl.DownloadFile(w, downloadRequest("u2", "role_participant", false))

	assert.Equal(t, http.StatusForbidden, w.Code)
	svc.AssertNotCalled(t, "OpenFile", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubmissionController_DownloadFile_ResponsibleTeacher(t *testing.T) {
	ctrl, courseSvc, activitySvc, svc := newSubmissionController()
	svc.On("GetSubmission", mock.Anything, "a1_p1").Return(entregaConArchivo, nil)
	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-1").Return(nil)
	expectDownload(svc)

	w := httptest.NewRecorder()
	ctrl.DownloadFile(w, downloadRequest("teacher-1", "role_teacher", false))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSubmissionController_DownloadFile_OtherTeacherForbidden(t *testing.T) {
	ctrl, courseSvc, activitySvc, svc := newSubmissionController()
	svc.On("GetSubmission", mock.Anything, "a1_p1").Return(entregaConArchivo, nil)
	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "teacher-2").Return(services.NewForbiddenError("no"))

	w := httptest.NewRecorder()
	ctrl.DownloadFile(w, downloadRequest("teacher-2", "role_teacher", false))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSubmissionController_DownloadFile_ExternalTeacherAndCompanyForbidden(t *testing.T) {
	for _, c := range []struct {
		role     string
		external bool
	}{{"role_teacher", true}, {"role_company", false}} {
		ctrl, _, _, svc := newSubmissionController()
		svc.On("GetSubmission", mock.Anything, "a1_p1").Return(entregaConArchivo, nil)

		w := httptest.NewRecorder()
		ctrl.DownloadFile(w, downloadRequest("x", c.role, c.external))

		assert.Equal(t, http.StatusForbidden, w.Code, c.role)
	}
}

func TestSubmissionController_DownloadFile_BadIndex(t *testing.T) {
	ctrl, _, _, _ := newSubmissionController()
	req := withChiURLParams(withAuth(httptest.NewRequest("GET", "/", nil), "u", "role_admin", false), "entregaId", "a1_p1", "indice", "x")
	w := httptest.NewRecorder()
	ctrl.DownloadFile(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmissionController_DownloadFile_SubmissionNotFound(t *testing.T) {
	ctrl, _, _, svc := newSubmissionController()
	svc.On("GetSubmission", mock.Anything, "a1_p1").Return(nil, repoErrNotFound())

	w := httptest.NewRecorder()
	ctrl.DownloadFile(w, downloadRequest("u", "role_admin", false))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func newActivityController() (*controllers.ActivityController, *MockCourseService, *MockActivityService, *MockSubmissionService) {
	courseSvc := new(MockCourseService)
	activitySvc := new(MockActivityService)
	submissionSvc := new(MockSubmissionService)
	return controllers.NewActivityController(courseSvc, activitySvc, submissionSvc), courseSvc, activitySvc, submissionSvc
}

func TestActivityController_Create_AcceptsDraftCourse(t *testing.T) {
	ctrl, courseSvc, activitySvc, _ := newActivityController()
	courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
	activitySvc.On("CreateActivity", mock.Anything, "c1", mock.AnythingOfType("models.ActivityInput")).Return(&models.Activity{ID: "a1"}, nil)

	body, _ := json.Marshal(models.ActivityInput{Titulo: "t", Consigna: "c"})
	req := withChiURLParams(httptest.NewRequest("POST", "/api/cursos/c1/actividades", bytes.NewBuffer(body)), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestActivityController_Create_CourseNotFound(t *testing.T) {
	ctrl, courseSvc, activitySvc, _ := newActivityController()
	courseSvc.On("VerifyExists", mock.Anything, "c1").Return(repoErrNotFound())

	req := withChiURLParams(httptest.NewRequest("POST", "/api/cursos/c1/actividades", strings.NewReader("{}")), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Create(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	activitySvc.AssertNotCalled(t, "CreateActivity", mock.Anything, mock.Anything, mock.Anything)
}

func TestActivityController_List(t *testing.T) {
	ctrl, _, activitySvc, _ := newActivityController()
	activitySvc.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{{ID: "a1"}}, nil)

	w := httptest.NewRecorder()
	ctrl.List(w, withChiURLParams(httptest.NewRequest("GET", "/", nil), "id", "c1"))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestActivityController_UpdateAndDelete(t *testing.T) {
	ctrl, _, activitySvc, _ := newActivityController()
	activitySvc.On("UpdateActivity", mock.Anything, "a1", mock.AnythingOfType("models.ActivityInput")).Return(&models.Activity{ID: "a1"}, nil)
	activitySvc.On("DeleteActivity", mock.Anything, "a1").Return(nil)

	w := httptest.NewRecorder()
	ctrl.Update(w, withChiURLParams(httptest.NewRequest("PUT", "/", strings.NewReader(`{"titulo":"t"}`)), "actividadId", "a1"))
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	ctrl.Update(w, withChiURLParams(httptest.NewRequest("PUT", "/", strings.NewReader(`{`)), "actividadId", "a1"))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	ctrl.Delete(w, withChiURLParams(httptest.NewRequest("DELETE", "/", nil), "actividadId", "a1"))
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestActivityController_ListSubmissions(t *testing.T) {
	ctrl, _, activitySvc, submissionSvc := newActivityController()
	activity := &models.Activity{ID: "a1", CursoID: "c1"}
	activitySvc.On("GetActivity", mock.Anything, "a1").Return(activity, nil)
	submissionSvc.On("ListByActivity", mock.Anything, activity, "tok").Return([]models.SubmissionStatusRow{}, nil)

	w := httptest.NewRecorder()
	ctrl.ListSubmissions(w, withChiURLParams(withAuth(httptest.NewRequest("GET", "/", nil), "u", "role_volunteer", false), "actividadId", "a1"))

	assert.Equal(t, http.StatusOK, w.Code)
}
