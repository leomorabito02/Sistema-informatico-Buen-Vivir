package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/services"
)

// fakeBajas es el repositorio de bajas de cursos en memoria.
type fakeBajas struct {
	ids   map[string]bool
	ultim models.CursoBaja
	err   error
}

func (f *fakeBajas) DarDeBaja(_ context.Context, b models.CursoBaja) error {
	if f.err != nil {
		return f.err
	}
	if f.ids[b.CursoID] {
		return repositories.ErrCursoYaDadoDeBaja
	}
	if f.ids == nil {
		f.ids = map[string]bool{}
	}
	f.ids[b.CursoID] = true
	f.ultim = b
	return nil
}

func (f *fakeBajas) IDs(context.Context) (map[string]bool, error) { return f.ids, f.err }

func (f *fakeBajas) EstaDadoDeBaja(_ context.Context, id string) (bool, error) {
	return f.ids[id], f.err
}

// ---- HU-16: baja lógica ----

func TestDeleteCourse_EsBajaLogica(t *testing.T) {
	repo := new(MockCourseRepository)
	enrollments := new(MockEnrollmentRepository)
	bajas := &fakeBajas{}
	svc := services.NewCourseServiceConBaja(repo, nil, bajas, enrollments)

	repo.On("ExistsInManagement", mock.Anything, "c1").Return(true, nil)
	repo.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", Title: "Huerta"}, nil)
	enrollments.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{
		{Estado: models.EnrollmentEstadoActiva}, {Estado: models.EnrollmentEstadoActiva}, {Estado: models.EnrollmentEstadoBaja},
	}, nil)

	require.NoError(t, svc.DeleteCourse(context.Background(), "c1"))
	assert.Equal(t, "Huerta", bajas.ultim.Titulo)
	assert.Equal(t, 2, bajas.ultim.InscriptosActivos)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)

	// una segunda baja del mismo curso
	err := svc.DeleteCourse(context.Background(), "c1")
	require.True(t, services.IsValidationError(err))
	assert.Contains(t, err.Error(), "ya fue dado de baja")
}

func TestDeleteCourse_CursoInexistenteYSinConfigurar(t *testing.T) {
	repo := new(MockCourseRepository)
	svc := services.NewCourseServiceConBaja(repo, nil, &fakeBajas{}, new(MockEnrollmentRepository))
	repo.On("ExistsInManagement", mock.Anything, "nada").Return(false, nil)
	assert.True(t, repositories.IsNotFound(svc.DeleteCourse(context.Background(), "nada")))

	assert.Error(t, services.NewCourseService(repo).DeleteCourse(context.Background(), "c1"))
}

func TestListCourses_SacaLosDadosDeBaja(t *testing.T) {
	repo := new(MockCourseRepository)
	bajas := &fakeBajas{ids: map[string]bool{"c2": true}}
	svc := services.NewCourseServiceConBaja(repo, nil, bajas, new(MockEnrollmentRepository))
	repo.On("List", mock.Anything).Return([]models.Course{{ID: "c1"}, {ID: "c2"}}, nil)
	repo.On("ListAllDrafts", mock.Anything).Return([]models.Course{{ID: "c2"}, {ID: "c3"}}, nil)

	list, err := svc.ListCourses(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []models.Course{{ID: "c1"}}, list)

	drafts, err := svc.ListAllDraftCourses(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []models.Course{{ID: "c3"}}, drafts)

	bajas.err = errors.New("firestore caído")
	_, err = svc.ListCourses(context.Background())
	assert.Error(t, err)
}

func TestEnroll_CursoDadoDeBaja(t *testing.T) {
	repo := new(MockEnrollmentRepository)
	courses := new(MockCourseRepository)
	auth := &fakeAuthServer{participants: []models.AuthParticipant{participanteAprobado}}
	svc := services.NewEnrollmentService(repo, courses, auth.start(t), &fakeBajas{ids: map[string]bool{"c1": true}})
	courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)

	_, err := svc.Enroll(context.Background(), "p1", "c1", "tok")
	require.True(t, services.IsValidationError(err))
	assert.Contains(t, err.Error(), "dado de baja")
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestEnrollment_CountActiveYNombres(t *testing.T) {
	repo := new(MockEnrollmentRepository)
	auth := &fakeAuthServer{
		participants: []models.AuthParticipant{{ParticipantID: "p1", UserID: "u1"}},
		users:        []models.AuthUser{{UserID: "u1", FirstName: "Ana", LastName: "Paz"}},
	}
	svc := services.NewEnrollmentService(repo, nil, auth.start(t))
	repo.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{
		{ParticipanteID: "p1", Estado: models.EnrollmentEstadoActiva},
		{ParticipanteID: "p9", Estado: models.EnrollmentEstadoFinalizada},
	}, nil)

	n, err := svc.CountActive(context.Background(), "c1")
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	_, err = svc.CountActive(context.Background(), " ")
	assert.True(t, services.IsValidationError(err))

	list, err := svc.ListByCourseConNombres(context.Background(), "c1", "tok")
	require.NoError(t, err)
	assert.Equal(t, "Ana", list[0].Nombre)
	assert.Equal(t, "Paz", list[0].Apellido)
	assert.Empty(t, list[1].Nombre, "un participante que auth no conoce queda sin nombre")

	// Sin permiso para listar (docente): los nombres se piden de a uno.
	auth.failList = true
	list, err = svc.ListByCourseConNombres(context.Background(), "c1", "tok")
	require.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "Ana", list[0].Nombre)
	assert.Equal(t, "Paz", list[0].Apellido)
	assert.Empty(t, list[1].Nombre)
}

// ---- HU-28: filtro por área ----

func TestFiltrarCursosPorArea(t *testing.T) {
	cursos := []models.Course{
		{ID: "c1", Title: "Huerta", TrainingArea: &models.TrainingArea{ID: "a1"}},
		{ID: "c2", Title: "Huerta urbana", TrainingArea: &models.TrainingArea{ID: "a2"}},
		{ID: "c3", Title: "Sin área"},
	}
	assert.Len(t, services.FiltrarCursosPorArea(cursos, ""), 3)
	assert.Equal(t, "c2", services.FiltrarCursosPorArea(cursos, " a2 ")[0].ID)
	// combinado con el buscador por nombre
	got := services.FiltrarCursosPorArea(services.FiltrarCursos(cursos, "huerta"), "a1")
	require.Len(t, got, 1)
	assert.Equal(t, "c1", got[0].ID)
}

// ---- buscador de participantes ----

func participantesFixture(t *testing.T) (services.ParticipanteService, *MockEnrollmentRepository, *fakeAuthServer) {
	auth := &fakeAuthServer{
		participants: []models.AuthParticipant{
			{ParticipantID: "p1", UserID: "u1", DNI: "30111222", ApprovalStatus: "approved"},
			{ParticipantID: "p2", UserID: "u2", DNI: "40999888", ApprovalStatus: "approved"},
			{ParticipantID: "p3", UserID: "u3", ApprovalStatus: "pending"},
		},
		users: []models.AuthUser{
			{UserID: "u1", FirstName: "María José", LastName: "Pérez"},
			{UserID: "u2", FirstName: "Mario", LastName: "Gómez"},
			{UserID: "u3", FirstName: "Mariana", LastName: "Ríos"},
		},
	}
	enrollments := new(MockEnrollmentRepository)
	svc := services.NewParticipanteService(auth.start(t), enrollments, new(MockActivityRepository), new(MockAttendanceRepository), new(MockCourseRepository))
	return svc, enrollments, auth
}

func TestBuscarParticipantes(t *testing.T) {
	svc, enrollments, auth := participantesFixture(t)
	ctx := context.Background()

	got, err := svc.Buscar(ctx, "mari", "", "tok")
	require.NoError(t, err)
	require.Len(t, got, 2, "la pendiente no aparece")
	assert.Equal(t, "Gómez", got[0].Apellido, "ordenados por apellido")
	assert.Equal(t, "tok", auth.receivedToken)

	got, err = svc.Buscar(ctx, "PEREZ", "", "tok")
	require.NoError(t, err)
	require.Len(t, got, 1, "sin distinguir mayúsculas ni tildes")
	assert.Equal(t, "p1", got[0].ParticipanteID)

	got, err = svc.Buscar(ctx, "40999", "", "tok")
	require.NoError(t, err)
	require.Len(t, got, 1, "por DNI")
	assert.Equal(t, "p2", got[0].ParticipanteID)

	enrollments.On("ListByCourse", mock.Anything, "c1").Return([]models.Enrollment{
		{ParticipanteID: "p2", Estado: models.EnrollmentEstadoActiva},
		{ParticipanteID: "p1", Estado: models.EnrollmentEstadoBaja},
	}, nil)
	got, err = svc.Buscar(ctx, "mari", "c1", "tok")
	require.NoError(t, err)
	require.Len(t, got, 1, "deja afuera a los inscriptos activos del curso")
	assert.Equal(t, "p1", got[0].ParticipanteID)

	_, err = svc.Buscar(ctx, " m ", "", "tok")
	assert.True(t, services.IsValidationError(err))

	auth.failList = true
	_, err = svc.Buscar(ctx, "mari", "", "tok")
	assert.Error(t, err)
}

// ---- HU-19: historial de asistencia por participante ----

func TestHistorialAsistencia(t *testing.T) {
	enrollments := new(MockEnrollmentRepository)
	activities := new(MockActivityRepository)
	attendance := new(MockAttendanceRepository)
	courses := new(MockCourseRepository)
	svc := services.NewParticipanteService(nil, enrollments, activities, attendance, courses)

	d := func(day int) time.Time { return time.Date(2026, 10, day, 0, 0, 0, 0, time.UTC) }
	enrollments.On("ListByParticipant", mock.Anything, "p1").Return([]models.Enrollment{
		{CursoID: "c1", Estado: models.EnrollmentEstadoActiva, FechaInscripcion: d(1)},
		{CursoID: "c2", Estado: models.EnrollmentEstadoFinalizada, FechaInscripcion: d(5)},
	}, nil)
	attendance.On("ListByParticipant", mock.Anything, "p1").Return([]models.Attendance{
		{ActividadID: "a2", Presente: true},
	}, nil)
	activities.On("ListByCourse", mock.Anything, "c1").Return([]models.Activity{
		{ID: "a2", Titulo: "Segunda", Fecha: d(10)},
		{ID: "a1", Titulo: "Primera", Fecha: d(3)},
	}, nil)
	activities.On("ListByCourse", mock.Anything, "c2").Return([]models.Activity{}, nil)
	courses.On("Get", mock.Anything, "c1").Return(&models.Course{ID: "c1", Title: "Huerta"}, nil)
	courses.On("Get", mock.Anything, "c2").Return(nil, repositories.ErrNotFound)

	got, err := svc.HistorialAsistencia(context.Background(), "p1")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "c2", got[0].CursoID, "la inscripción más reciente primero")
	assert.Equal(t, "Curso no disponible", got[0].CursoTitulo)
	assert.Zero(t, got[0].PorcentajeAsistencia)

	h := got[1]
	assert.Equal(t, "Huerta", h.CursoTitulo)
	assert.Equal(t, 2, h.ActividadesTotal)
	assert.Equal(t, 1, h.ActividadesAsistidas)
	assert.InDelta(t, 50.0, h.PorcentajeAsistencia, 0.01)
	assert.Equal(t, "Primera", h.Actividades[0].Titulo, "actividades en orden de fecha")
	assert.False(t, h.Actividades[0].Presente)
	assert.True(t, h.Actividades[1].Presente)

	_, err = svc.HistorialAsistencia(context.Background(), " ")
	assert.True(t, services.IsValidationError(err))
}

// ---- auditoría ----

type fakeAudit struct {
	registros    []models.RegistroAuditoria
	desde, hasta time.Time
}

func (f *fakeAudit) List(_ context.Context, desde, hasta time.Time, _ int) ([]models.RegistroAuditoria, error) {
	f.desde, f.hasta = desde, hasta
	return f.registros, nil
}

func TestAuditoriaService_List(t *testing.T) {
	repo := &fakeAudit{registros: []models.RegistroAuditoria{
		{ID: "1", UsuarioID: "u1", Accion: "baja", Recurso: "curso"},
		{ID: "2", UsuarioID: "u2", Accion: "registrar", Recurso: "asistencia_general"},
	}}
	auth := &fakeAuthServer{users: []models.AuthUser{{UserID: "u1", FirstName: "Ana", LastName: "Paz"}}}
	svc := services.NewAuditoriaService(repo, auth.start(t))

	got, err := svc.List(context.Background(), services.FiltroAuditoria{Desde: "2026-10-01", Hasta: "2026-10-22", Recurso: "curso"}, "tok")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Ana Paz", got[0].UsuarioNombre)
	assert.Equal(t, time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC), repo.desde, "medianoche de Argentina")
	assert.Equal(t, time.Date(2026, 10, 23, 3, 0, 0, 0, time.UTC), repo.hasta, "incluye el día hasta")

	got, err = svc.List(context.Background(), services.FiltroAuditoria{Accion: "registrar"}, "tok")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Empty(t, got[0].UsuarioNombre)

	_, err = svc.List(context.Background(), services.FiltroAuditoria{Desde: "2026-10-30", Hasta: "2026-10-01"}, "tok")
	assert.True(t, services.IsValidationError(err))
}

func TestRangoAuditoria(t *testing.T) {
	now := time.Date(2026, 10, 23, 2, 0, 0, 0, time.UTC) // 22/10 23 h en Argentina
	desde, hasta, err := services.RangoAuditoria("", "", now)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 9, 23, 3, 0, 0, 0, time.UTC), desde, "últimos 30 días")
	assert.Equal(t, time.Date(2026, 10, 23, 3, 0, 0, 0, time.UTC), hasta)

	for _, c := range [][2]string{{"mal", ""}, {"", "mal"}, {"2024-01-01", "2026-01-01"}} {
		_, _, err := services.RangoAuditoria(c[0], c[1], now)
		assert.True(t, services.IsValidationError(err), c)
	}
}
