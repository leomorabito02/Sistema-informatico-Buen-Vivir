package services

import (
	"context"
	"sort"
	"strings"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// MaxResultadosParticipantes acota la respuesta del buscador: es un
// selector, no un listado.
const MaxResultadosParticipantes = 20

// ParticipanteService reúne lo que admin y voluntario consultan de un
// participante en Educación: el buscador del selector de inscripciones
// y su historial de asistencia (HU-19).
type ParticipanteService interface {
	// Buscar devuelve hasta 20 participantes aprobados cuyo nombre,
	// apellido o DNI contiene q (sin distinguir mayúsculas ni tildes).
	// Con cursoID, deja afuera a quienes ya tienen una inscripción
	// activa en ese curso.
	Buscar(ctx context.Context, q, cursoID, bearerToken string) ([]models.ParticipanteResumen, error)
	// HistorialAsistencia arma, por cada curso en el que se inscribió,
	// la lista de actividades con su asistencia.
	HistorialAsistencia(ctx context.Context, participanteID string) ([]models.HistorialCurso, error)
}

type participanteService struct {
	authClient     *repositories.AuthServiceClient
	enrollmentRepo repositories.EnrollmentRepository
	activityRepo   repositories.ActivityRepository
	attendanceRepo repositories.AttendanceRepository
	courseRepo     repositories.CourseRepository
}

func NewParticipanteService(
	authClient *repositories.AuthServiceClient,
	enrollmentRepo repositories.EnrollmentRepository,
	activityRepo repositories.ActivityRepository,
	attendanceRepo repositories.AttendanceRepository,
	courseRepo repositories.CourseRepository,
) ParticipanteService {
	return &participanteService{
		authClient:     authClient,
		enrollmentRepo: enrollmentRepo,
		activityRepo:   activityRepo,
		attendanceRepo: attendanceRepo,
		courseRepo:     courseRepo,
	}
}

func (s *participanteService) Buscar(ctx context.Context, q, cursoID, bearerToken string) ([]models.ParticipanteResumen, error) {
	q = normalizar(q)
	if len([]rune(q)) < 2 {
		return nil, NewValidationError("Escribí al menos 2 letras del nombre o del DNI")
	}
	if len([]rune(q)) > MaxBusquedaCursos {
		return nil, NewValidationError("La búsqueda admite hasta 100 caracteres")
	}

	directorio, err := directorioParticipantes(ctx, s.authClient, bearerToken)
	if err != nil {
		return nil, err
	}

	inscriptos := map[string]bool{}
	if strings.TrimSpace(cursoID) != "" {
		enrollments, err := s.enrollmentRepo.ListByCourse(ctx, cursoID)
		if err != nil {
			return nil, err
		}
		for _, e := range enrollments {
			if e.Estado == models.EnrollmentEstadoActiva {
				inscriptos[e.ParticipanteID] = true
			}
		}
	}

	out := []models.ParticipanteResumen{}
	for id, p := range directorio {
		if !p.Aprobado || inscriptos[id] || !coincideParticipante(p, q) {
			continue
		}
		out = append(out, models.ParticipanteResumen{ParticipanteID: id, Nombre: p.Nombre, Apellido: p.Apellido})
	}
	sort.Slice(out, func(i, j int) bool {
		a := normalizar(out[i].Apellido + " " + out[i].Nombre)
		b := normalizar(out[j].Apellido + " " + out[j].Nombre)
		if a != b {
			return a < b
		}
		return out[i].ParticipanteID < out[j].ParticipanteID
	})
	if len(out) > MaxResultadosParticipantes {
		out = out[:MaxResultadosParticipantes]
	}
	return out, nil
}

// coincideParticipante: q (ya normalizado) en "nombre apellido",
// "apellido nombre" o el DNI.
func coincideParticipante(p participanteDirectorio, q string) bool {
	return strings.Contains(normalizar(p.Nombre+" "+p.Apellido), q) ||
		strings.Contains(normalizar(p.Apellido+" "+p.Nombre), q) ||
		(p.DNI != "" && strings.Contains(p.DNI, q))
}

func (s *participanteService) HistorialAsistencia(ctx context.Context, participanteID string) ([]models.HistorialCurso, error) {
	if strings.TrimSpace(participanteID) == "" {
		return nil, NewValidationError("participanteId es obligatorio")
	}
	enrollments, err := s.enrollmentRepo.ListByParticipant(ctx, participanteID)
	if err != nil {
		return nil, err
	}
	attendances, err := s.attendanceRepo.ListByParticipant(ctx, participanteID)
	if err != nil {
		return nil, err
	}
	presente := make(map[string]bool, len(attendances))
	for _, a := range attendances {
		if a.Presente {
			presente[a.ActividadID] = true
		}
	}

	out := make([]models.HistorialCurso, 0, len(enrollments))
	for _, e := range enrollments {
		activities, err := s.activityRepo.ListByCourse(ctx, e.CursoID)
		if err != nil {
			return nil, err
		}
		sort.Slice(activities, func(i, j int) bool { return activities[i].Fecha.Before(activities[j].Fecha) })

		h := models.HistorialCurso{
			CursoID:           e.CursoID,
			CursoTitulo:       s.tituloCurso(ctx, e.CursoID),
			EstadoInscripcion: e.Estado,
			FechaInscripcion:  e.FechaInscripcion,
			ActividadesTotal:  len(activities),
			Actividades:       make([]models.HistorialActividad, 0, len(activities)),
		}
		for _, a := range activities {
			if presente[a.ID] {
				h.ActividadesAsistidas++
			}
			h.Actividades = append(h.Actividades, models.HistorialActividad{
				ActividadID: a.ID, Titulo: a.Titulo, Fecha: a.Fecha, Presente: presente[a.ID],
			})
		}
		if h.ActividadesTotal > 0 {
			h.PorcentajeAsistencia = float64(h.ActividadesAsistidas) / float64(h.ActividadesTotal) * 100
		}
		out = append(out, h)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].FechaInscripcion.After(out[j].FechaInscripcion) })
	return out, nil
}

// tituloCurso lee el título del curso publicado (CDA, con caché). Un
// curso despublicado o borrado ya no está ahí.
func (s *participanteService) tituloCurso(ctx context.Context, cursoID string) string {
	if c, err := s.courseRepo.Get(ctx, cursoID); err == nil && c != nil && c.Title != "" {
		return c.Title
	}
	return "Curso no disponible"
}
