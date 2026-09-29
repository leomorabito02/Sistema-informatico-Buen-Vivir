package services_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSubmissionRepository struct {
	mock.Mock
}

func (m *MockSubmissionRepository) CreateWithAttendance(ctx context.Context, activity *models.Activity, sub models.Submission) (*models.Submission, error) {
	args := m.Called(ctx, activity, sub)
	if fn, ok := args.Get(0).(func(context.Context, *models.Activity, models.Submission) *models.Submission); ok {
		return fn(ctx, activity, sub), args.Error(1)
	}
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Submission), args.Error(1)
}

func (m *MockSubmissionRepository) Get(ctx context.Context, id string) (*models.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Submission), args.Error(1)
}

func (m *MockSubmissionRepository) ListByActivity(ctx context.Context, actividadID string) ([]models.Submission, error) {
	args := m.Called(ctx, actividadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Submission), args.Error(1)
}

func (m *MockSubmissionRepository) GetForParticipant(ctx context.Context, participanteID string, actividadIDs []string) (map[string]*models.Submission, error) {
	args := m.Called(ctx, participanteID, actividadIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*models.Submission), args.Error(1)
}

type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) Upload(ctx context.Context, path, contentType string, data []byte) error {
	return m.Called(ctx, path, contentType, data).Error(0)
}

func (m *MockFileStorage) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockFileStorage) DeletePrefix(ctx context.Context, prefix string) error {
	return m.Called(ctx, prefix).Error(0)
}

type MockParticipantDirectory struct {
	mock.Mock
}

func (m *MockParticipantDirectory) GetMyParticipantID(ctx context.Context, bearerToken string) (string, error) {
	args := m.Called(ctx, bearerToken)
	return args.String(0), args.Error(1)
}

func (m *MockParticipantDirectory) ListParticipants(ctx context.Context, bearerToken string) ([]models.AuthParticipant, error) {
	args := m.Called(ctx, bearerToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AuthParticipant), args.Error(1)
}

func (m *MockParticipantDirectory) ListUsers(ctx context.Context, bearerToken string) ([]models.AuthUser, error) {
	args := m.Called(ctx, bearerToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AuthUser), args.Error(1)
}

type submissionFixture struct {
	repo       *MockSubmissionRepository
	activities *MockActivityRepository
	enrollment *MockEnrollmentRepository
	courses    *MockCourseRepository
	storage    *MockFileStorage
	directory  *MockParticipantDirectory
	svc        services.SubmissionService
}

func newSubmissionFixture(withStorage bool) *submissionFixture {
	f := &submissionFixture{
		repo:       new(MockSubmissionRepository),
		activities: new(MockActivityRepository),
		enrollment: new(MockEnrollmentRepository),
		courses:    new(MockCourseRepository),
		storage:    new(MockFileStorage),
		directory:  new(MockParticipantDirectory),
	}
	var storage repositories.FileStorage
	if withStorage {
		storage = f.storage
	}
	f.svc = services.NewSubmissionService(f.repo, f.activities, f.enrollment, f.courses, storage, f.directory)
	return f
}

var fechaActividad = time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)

func (f *submissionFixture) givenActivityAndEnrollment(estado string) *models.Activity {
	activity := &models.Activity{ID: "a1", CursoID: "c1", Titulo: "Práctica", Fecha: fechaActividad}
	f.activities.On("Get", mock.Anything, "a1").Return(activity, nil)
	f.enrollment.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{
		{ParticipanteID: "p1", CursoID: "otro", Estado: models.EnrollmentEstadoActiva},
		{ParticipanteID: "p1", CursoID: "c1", Estado: estado},
	}, nil)
	return activity
}

func TestSubmissionService_ResolveParticipantID(t *testing.T) {
	f := newSubmissionFixture(false)
	f.directory.On("GetMyParticipantID", mock.Anything, "tok").Return("p1", nil)

	id, err := f.svc.ResolveParticipantID(context.Background(), "tok")
	assert.NoError(t, err)
	assert.Equal(t, "p1", id)
}

func TestSubmissionService_ResolveParticipantID_NoProfileIsForbidden(t *testing.T) {
	f := newSubmissionFixture(false)
	f.directory.On("GetMyParticipantID", mock.Anything, "tok").Return("", nil)

	_, err := f.svc.ResolveParticipantID(context.Background(), "tok")
	assert.True(t, services.IsForbiddenError(err))
}

func TestSubmissionService_ResolveParticipantID_AuthError(t *testing.T) {
	f := newSubmissionFixture(false)
	f.directory.On("GetMyParticipantID", mock.Anything, "tok").Return("", errors.New("auth caído"))

	_, err := f.svc.ResolveParticipantID(context.Background(), "tok")
	assert.Error(t, err)
	assert.False(t, services.IsForbiddenError(err))
}

func TestSubmissionService_Submit_TextOnly_CreatesSubmissionWithAttendance(t *testing.T) {
	f := newSubmissionFixture(false)
	activity := f.givenActivityAndEnrollment(models.EnrollmentEstadoActiva)
	f.repo.On("GetForParticipant", mock.Anything, "p1", []string{"a1"}).Return(map[string]*models.Submission{}, nil)
	f.repo.On("CreateWithAttendance", mock.Anything, activity, mock.MatchedBy(func(s models.Submission) bool {
		return s.ParticipanteID == "p1" && s.Texto == "mi respuesta" && len(s.Archivos) == 0 && !s.FechaEntrega.IsZero()
	})).Return(&models.Submission{ID: "a1_p1"}, nil)

	sub, err := f.svc.Submit(context.Background(), "a1", "p1", "  mi respuesta  ", nil)
	require.NoError(t, err)
	assert.Equal(t, "a1_p1", sub.ID)
}

func TestSubmissionService_Submit_NotActiveEnrollmentIsForbidden(t *testing.T) {
	for _, estado := range []string{models.EnrollmentEstadoFinalizada, models.EnrollmentEstadoBaja} {
		f := newSubmissionFixture(false)
		f.givenActivityAndEnrollment(estado)

		_, err := f.svc.Submit(context.Background(), "a1", "p1", "x", nil)
		assert.True(t, services.IsForbiddenError(err), estado)
		f.repo.AssertNotCalled(t, "CreateWithAttendance", mock.Anything, mock.Anything, mock.Anything)
	}
}

func TestSubmissionService_Submit_NotEnrolledIsForbidden(t *testing.T) {
	f := newSubmissionFixture(false)
	f.activities.On("Get", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	f.enrollment.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{}, nil)

	_, err := f.svc.Submit(context.Background(), "a1", "p1", "x", nil)
	assert.True(t, services.IsForbiddenError(err))
}

func TestSubmissionService_Submit_AlreadySubmitted(t *testing.T) {
	f := newSubmissionFixture(false)
	f.givenActivityAndEnrollment(models.EnrollmentEstadoActiva)
	f.repo.On("GetForParticipant", mock.Anything, "p1", []string{"a1"}).Return(map[string]*models.Submission{"a1": {ID: "a1_p1"}}, nil)

	_, err := f.svc.Submit(context.Background(), "a1", "p1", "x", nil)
	assert.True(t, services.IsValidationError(err))
	assert.Equal(t, "Ya entregaste esta actividad", err.Error())
}

func TestSubmissionService_Submit_RaceLostIsValidationAndCleansUp(t *testing.T) {
	f := newSubmissionFixture(true)
	activity := f.givenActivityAndEnrollment(models.EnrollmentEstadoActiva)
	f.repo.On("GetForParticipant", mock.Anything, "p1", []string{"a1"}).Return(map[string]*models.Submission{}, nil)
	f.storage.On("Upload", mock.Anything, mock.Anything, "application/pdf", mock.Anything).Return(nil)
	f.storage.On("DeletePrefix", mock.Anything, mock.MatchedBy(func(p string) bool { return strings.HasPrefix(p, "entregas/a1/p1/") })).Return(nil)
	f.repo.On("CreateWithAttendance", mock.Anything, activity, mock.Anything).Return(nil, repositories.ErrAlreadySubmitted)

	_, err := f.svc.Submit(context.Background(), "a1", "p1", "", []models.SubmissionUpload{{Nombre: "tp.pdf", Tipo: "application/pdf", Data: []byte("%PDF")}})
	assert.True(t, services.IsValidationError(err))
	f.storage.AssertCalled(t, "DeletePrefix", mock.Anything, mock.Anything)
}

func TestSubmissionService_Submit_WithFiles_UploadsAndStoresMetadata(t *testing.T) {
	f := newSubmissionFixture(true)
	activity := f.givenActivityAndEnrollment(models.EnrollmentEstadoActiva)
	f.repo.On("GetForParticipant", mock.Anything, "p1", []string{"a1"}).Return(map[string]*models.Submission{}, nil)

	var uploaded []string
	f.storage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		uploaded = append(uploaded, args.String(1))
	}).Return(nil)
	f.repo.On("CreateWithAttendance", mock.Anything, activity, mock.Anything).Return(func(_ context.Context, _ *models.Activity, s models.Submission) *models.Submission {
		return &s
	}, nil)

	sub, err := f.svc.Submit(context.Background(), "a1", "p1", "", []models.SubmissionUpload{
		{Nombre: `C:\Users\yo\mi trabajo final.docx`, Tipo: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Data: []byte("docx")},
		{Nombre: "foto.png", Tipo: "image/png", Data: []byte("png")},
	})
	require.NoError(t, err)
	require.Len(t, sub.Archivos, 2)
	assert.Equal(t, "mi trabajo final.docx", sub.Archivos[0].Nombre)
	assert.Equal(t, int64(4), sub.Archivos[0].Tamano)
	require.Len(t, uploaded, 2)
	for i, p := range uploaded {
		assert.True(t, strings.HasPrefix(p, "entregas/a1/p1/"), p)
		assert.Equal(t, p, sub.Archivos[i].Ruta)
		assert.NotContains(t, p, " ")
	}
}

func TestSubmissionService_Submit_UploadErrorCleansUp(t *testing.T) {
	f := newSubmissionFixture(true)
	f.givenActivityAndEnrollment(models.EnrollmentEstadoActiva)
	f.repo.On("GetForParticipant", mock.Anything, "p1", []string{"a1"}).Return(map[string]*models.Submission{}, nil)
	f.storage.On("Upload", mock.Anything, mock.MatchedBy(func(p string) bool { return strings.Contains(p, "-0-") }), mock.Anything, mock.Anything).Return(nil)
	f.storage.On("Upload", mock.Anything, mock.MatchedBy(func(p string) bool { return strings.Contains(p, "-1-") }), mock.Anything, mock.Anything).Return(errors.New("bucket caído"))
	f.storage.On("DeletePrefix", mock.Anything, mock.Anything).Return(nil)

	_, err := f.svc.Submit(context.Background(), "a1", "p1", "", []models.SubmissionUpload{
		{Nombre: "a.pdf", Tipo: "application/pdf", Data: []byte("1")},
		{Nombre: "b.pdf", Tipo: "application/pdf", Data: []byte("2")},
	})
	assert.Error(t, err)
	f.storage.AssertCalled(t, "DeletePrefix", mock.Anything, mock.Anything)
	f.repo.AssertNotCalled(t, "CreateWithAttendance", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubmissionService_Submit_UploadValidation(t *testing.T) {
	pdf := func(name string, size int) models.SubmissionUpload {
		return models.SubmissionUpload{Nombre: name, Tipo: "application/pdf", Data: bytes.Repeat([]byte("x"), size)}
	}
	cases := map[string][]models.SubmissionUpload{
		"tipo no permitido": {{Nombre: "virus.exe", Tipo: "application/x-msdownload", Data: []byte("x")}},
		"vacío":             {pdf("a.pdf", 0)},
		"muy grande":        {pdf("a.pdf", 10<<20+1)},
		"sin nombre":        {pdf("  ", 1)},
		"demasiados":        {pdf("1.pdf", 1), pdf("2.pdf", 1), pdf("3.pdf", 1), pdf("4.pdf", 1), pdf("5.pdf", 1), pdf("6.pdf", 1)},
	}
	for name, uploads := range cases {
		f := newSubmissionFixture(true)
		_, err := f.svc.Submit(context.Background(), "a1", "p1", "", uploads)
		assert.True(t, services.IsValidationError(err), name)
		f.storage.AssertNotCalled(t, "Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	}
}

func TestSubmissionService_Submit_FilesWithoutBucket(t *testing.T) {
	f := newSubmissionFixture(false)
	_, err := f.svc.Submit(context.Background(), "a1", "p1", "", []models.SubmissionUpload{{Nombre: "a.pdf", Tipo: "application/pdf", Data: []byte("x")}})
	assert.True(t, services.IsValidationError(err))
}

func TestSubmissionService_Submit_TextTooLong(t *testing.T) {
	f := newSubmissionFixture(false)
	_, err := f.svc.Submit(context.Background(), "a1", "p1", strings.Repeat("a", 10001), nil)
	assert.True(t, services.IsValidationError(err))
}

func TestSubmissionService_ListByActivity_MergesEnrollmentsSubmissionsAndNames(t *testing.T) {
	f := newSubmissionFixture(false)
	activity := &models.Activity{ID: "a1", CursoID: "c1"}
	f.enrollment.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{
		{ParticipanteID: "p1", Estado: models.EnrollmentEstadoActiva},
		{ParticipanteID: "p2", Estado: models.EnrollmentEstadoActiva},
		{ParticipanteID: "p3", Estado: models.EnrollmentEstadoBaja},
		{ParticipanteID: "p4", Estado: models.EnrollmentEstadoBaja},
	}, nil)
	// p4 se dio de baja después de entregar: su entrega igual se muestra.
	f.repo.On("ListByActivity", mock.Anything, "a1").Return([]models.Submission{
		{ID: "a1_p2", ParticipanteID: "p2", ActividadID: "a1"},
		{ID: "a1_p4", ParticipanteID: "p4", ActividadID: "a1"},
	}, nil)
	f.directory.On("ListParticipants", mock.Anything, "tok").Return([]models.AuthParticipant{
		{ParticipantID: "p1", UserID: "u1"}, {ParticipantID: "p2", UserID: "u2"},
	}, nil)
	f.directory.On("ListUsers", mock.Anything, "tok").Return([]models.AuthUser{
		{UserID: "u1", FirstName: "Ana", LastName: "Zapata"}, {UserID: "u2", FirstName: "Bruno", LastName: "Alvarez"},
	}, nil)

	rows, err := f.svc.ListByActivity(context.Background(), activity, "tok")
	require.NoError(t, err)
	require.Len(t, rows, 3, "p3 (baja, sin entrega) no aparece")

	byID := map[string]models.SubmissionStatusRow{}
	for _, r := range rows {
		byID[r.ParticipanteID] = r
	}
	assert.Equal(t, models.SubmissionEstadoSinEntregar, byID["p1"].Estado)
	assert.Nil(t, byID["p1"].Entrega)
	assert.Equal(t, models.SubmissionEstadoEntregado, byID["p2"].Estado)
	assert.Equal(t, "a1_p2", byID["p2"].Entrega.ID)
	assert.Equal(t, models.SubmissionEstadoEntregado, byID["p4"].Estado)
	assert.Equal(t, "Alvarez", rows[0].Apellido, "ordenado por apellido")
}

func TestSubmissionService_ListByActivity_NamesAreBestEffort(t *testing.T) {
	f := newSubmissionFixture(false)
	f.enrollment.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{{ParticipanteID: "p1", Estado: models.EnrollmentEstadoActiva}}, nil)
	f.repo.On("ListByActivity", mock.Anything, "a1").Return([]models.Submission{}, nil)
	f.directory.On("ListParticipants", mock.Anything, "tok").Return(nil, errors.New("403"))

	rows, err := f.svc.ListByActivity(context.Background(), &models.Activity{ID: "a1", CursoID: "c1"}, "tok")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Empty(t, rows[0].Nombre)
}

func TestSubmissionService_ListByActivity_RepoError(t *testing.T) {
	f := newSubmissionFixture(false)
	f.enrollment.On("ListByCourse", mock.Anything, "c1").Return(nil, errors.New("firestore caído"))

	_, err := f.svc.ListByActivity(context.Background(), &models.Activity{ID: "a1", CursoID: "c1"}, "tok")
	assert.Error(t, err)
}

func TestSubmissionService_GetParticipantCourse_ProgressFromSubmissions(t *testing.T) {
	f := newSubmissionFixture(false)
	f.enrollment.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{{CursoID: "c1", Estado: models.EnrollmentEstadoActiva}}, nil)
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", Title: "Panadería"}, nil)
	f.activities.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{
		{ID: "a2", Fecha: fechaActividad.AddDate(0, 0, 7)},
		{ID: "a1", Fecha: fechaActividad},
		{ID: "a3", Fecha: fechaActividad.AddDate(0, 0, 14)},
		{ID: "a4", Fecha: fechaActividad.AddDate(0, 0, 21)},
	}, nil)
	f.repo.On("GetForParticipant", mock.Anything, "p1", []string{"a1", "a2", "a3", "a4"}).Return(map[string]*models.Submission{
		"a1": {ID: "a1_p1"},
	}, nil)

	detail, err := f.svc.GetParticipantCourse(context.Background(), "p1", "c1")
	require.NoError(t, err)
	assert.Equal(t, "Panadería", detail.Curso.Title)
	require.Len(t, detail.Actividades, 4)
	assert.Equal(t, "a1", detail.Actividades[0].ID, "ordenadas por fecha")
	assert.NotNil(t, detail.Actividades[0].Entrega)
	assert.Nil(t, detail.Actividades[1].Entrega)
	assert.Equal(t, 4, detail.Progreso.ActividadesTotal)
	assert.Equal(t, 1, detail.Progreso.ActividadesAsistidas)
	assert.InDelta(t, 25.0, detail.Progreso.PorcentajeAsistencia, 0.001)
}

func TestSubmissionService_GetParticipantCourse_NotEnrolled(t *testing.T) {
	f := newSubmissionFixture(false)
	f.enrollment.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{{CursoID: "c1", Estado: models.EnrollmentEstadoBaja}}, nil)

	_, err := f.svc.GetParticipantCourse(context.Background(), "p1", "c1")
	assert.True(t, services.IsForbiddenError(err))
}

func TestSubmissionService_GetParticipantCourse_FinishedEnrollmentCanStillView(t *testing.T) {
	f := newSubmissionFixture(false)
	f.enrollment.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{{CursoID: "c1", Estado: models.EnrollmentEstadoFinalizada}}, nil)
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)
	f.activities.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{}, nil)
	f.repo.On("GetForParticipant", mock.Anything, "p1", []string{}).Return(map[string]*models.Submission{}, nil)

	detail, err := f.svc.GetParticipantCourse(context.Background(), "p1", "c1")
	require.NoError(t, err)
	assert.Equal(t, 0.0, detail.Progreso.PorcentajeAsistencia)
}

func TestSubmissionService_ListParticipantCourses_SkipsBajaAndUnpublished(t *testing.T) {
	f := newSubmissionFixture(false)
	f.enrollment.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{
		{CursoID: "c2", Estado: models.EnrollmentEstadoActiva},
		{CursoID: "c1", Estado: models.EnrollmentEstadoFinalizada},
		{CursoID: "c3", Estado: models.EnrollmentEstadoBaja},
		{CursoID: "c4", Estado: models.EnrollmentEstadoActiva},
	}, nil)
	f.courses.On("Get", mock.Anything, "c2").Return(&models.Course{ID: "c2", Title: "Carpintería"}, nil)
	f.courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", Title: "Albañilería"}, nil)
	f.courses.On("Get", mock.Anything, "c4").Return(nil, repositories.ErrNotFound)

	courses, err := f.svc.ListParticipantCourses(context.Background(), "p1")
	require.NoError(t, err)
	require.Len(t, courses, 2)
	assert.Equal(t, "Albañilería", courses[0].Title)
	f.courses.AssertNotCalled(t, "Get", mock.Anything, "c3")
}

func TestSubmissionService_ListParticipantCourses_CourseError(t *testing.T) {
	f := newSubmissionFixture(false)
	f.enrollment.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{{CursoID: "c1", Estado: models.EnrollmentEstadoActiva}}, nil)
	f.courses.On("Get", mock.Anything, "c1").Return(nil, errors.New("contentful caído"))

	_, err := f.svc.ListParticipantCourses(context.Background(), "p1")
	assert.Error(t, err)
}

func TestSubmissionService_OpenFile(t *testing.T) {
	f := newSubmissionFixture(true)
	sub := &models.Submission{Archivos: []models.SubmissionFile{{Nombre: "a.pdf", Ruta: "entregas/a1/p1/x-0-a.pdf"}}}
	f.storage.On("Open", mock.Anything, "entregas/a1/p1/x-0-a.pdf").Return(io.NopCloser(strings.NewReader("pdf")), nil)

	file, rc, err := f.svc.OpenFile(context.Background(), sub, 0)
	require.NoError(t, err)
	defer func() { _ = rc.Close() }()
	assert.Equal(t, "a.pdf", file.Nombre)

	_, _, err = f.svc.OpenFile(context.Background(), sub, 1)
	assert.True(t, repositories.IsNotFound(err))
}

func TestSubmissionService_OpenFile_WithoutBucket(t *testing.T) {
	f := newSubmissionFixture(false)
	sub := &models.Submission{Archivos: []models.SubmissionFile{{Nombre: "a.pdf"}}}
	_, _, err := f.svc.OpenFile(context.Background(), sub, 0)
	assert.ErrorIs(t, err, repositories.ErrFileStorageNotConfigured)
}

func TestSubmissionService_GetSubmission(t *testing.T) {
	f := newSubmissionFixture(false)
	_, err := f.svc.GetSubmission(context.Background(), " ")
	assert.True(t, services.IsValidationError(err))

	f.repo.On("Get", mock.Anything, "a1_p1").Return(&models.Submission{ID: "a1_p1"}, nil)
	sub, err := f.svc.GetSubmission(context.Background(), "a1_p1")
	require.NoError(t, err)
	assert.Equal(t, "a1_p1", sub.ID)
}
