// Package services arma los datos del dashboard a partir de lo que
// devuelve cada servicio de origen. No guarda nada: cada pedido consulta
// en vivo (el Dashboard es un agregador, sin base de datos propia).
package services

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/models"
)

// Consultas en paralelo por servicio (actividades de cada curso,
// postulaciones de cada oferta): sin saturar al servicio de origen.
const maxParalelo = 8

type DashboardService interface {
	Participantes(ctx context.Context, cred clients.Credentials) ([]models.Participante, error)
	Cursos(ctx context.Context, cred clients.Credentials) ([]models.Curso, error)
	Habilidades(ctx context.Context, cred clients.Credentials) ([]models.ParticipanteHabilidad, error)
	Postulaciones(ctx context.Context, cred clients.Credentials) ([]models.Postulacion, error)
	Areas(ctx context.Context, cred clients.Credentials) ([]models.Area, error)
	Promocion(ctx context.Context, cred clients.Credentials) (models.Promocion, error)
	Equipo(ctx context.Context, cred clients.Credentials) (models.Equipo, error)
}

type dashboardService struct {
	auth      *clients.Upstream
	education *clients.Upstream
	labor     *clients.Upstream
	promocion *clients.Upstream
}

// NewDashboardService: promocion puede ser nil (sin Promoción Humana, ese
// indicador responde "no configurado").
func NewDashboardService(auth, education, labor, promocion *clients.Upstream) DashboardService {
	if promocion == nil {
		promocion = &clients.Upstream{Name: "promoción humana"}
	}
	return &dashboardService{auth: auth, education: education, labor: labor, promocion: promocion}
}

// ── Participantes (Autenticación) ─────────────────────────────────────

func (s *dashboardService) Participantes(ctx context.Context, cred clients.Credentials) ([]models.Participante, error) {
	var res models.AuthResponse[[]models.ParticipantHistory]
	if err := s.auth.GetJSON(ctx, "/api/v1/users/participants/history", cred, &res); err != nil {
		return nil, err
	}
	out := make([]models.Participante, 0, len(res.Data))
	for _, h := range res.Data {
		if p, ok := MapParticipante(h); ok {
			out = append(out, p)
		}
	}
	return out, nil
}

// MapParticipante aplica la definición de "participante activo" del
// sistema (la misma que CountActiveParticipants en auth): rol
// participante, cuenta activa y aprobada, sin baja lógica.
//
//   - Alta: la fecha de aprobación, o la de creación si no hay.
//   - Baja: la baja lógica (deleted_at); si la cuenta está bloqueada, la
//     última modificación (auth no guarda cuándo se bloqueó).
//   - Los que nunca fueron aprobados no cuentan (ok = false).
func MapParticipante(h models.ParticipantHistory) (models.Participante, bool) {
	aprobado := h.ApprovalStatus == "approved" || h.ApprovalStatus == "accepted"
	if !aprobado {
		return models.Participante{}, false
	}
	alta := h.CreatedAt
	if h.ApprovalDate != nil && !h.ApprovalDate.IsZero() {
		alta = *h.ApprovalDate
	}
	p := models.Participante{ID: h.UserID, Rol: "participante", Estado: "activo", FechaAlta: alta}
	switch {
	case h.DeletedAt != nil:
		baja := *h.DeletedAt
		p.Estado, p.FechaBaja = "baja", &baja
	case h.Status != "active":
		baja := h.UpdatedAt
		p.Estado, p.FechaBaja = "baja", &baja
	}
	return p, true
}

// Equipo: empresas, docentes y voluntarios con la cuenta activa y aprobada.
func (s *dashboardService) Equipo(ctx context.Context, cred clients.Credentials) (models.Equipo, error) {
	var res models.AuthResponse[models.Equipo]
	if err := s.auth.GetJSON(ctx, "/api/v1/users/active-counts", cred, &res); err != nil {
		return models.Equipo{}, err
	}
	return res.Data, nil
}

// ── Cursos y áreas (Educación) ────────────────────────────────────────

func (s *dashboardService) Areas(ctx context.Context, cred clients.Credentials) ([]models.Area, error) {
	var areas []models.TrainingArea
	if err := s.education.GetJSON(ctx, "/api/areas-formativas", cred, &areas); err != nil {
		return nil, err
	}
	out := make([]models.Area, len(areas))
	for i, a := range areas {
		out[i] = models.Area{ID: a.ID, Nombre: a.Name}
	}
	return out, nil
}

// cursosBase trae publicados y borradores, sin actividades.
func (s *dashboardService) cursosBase(ctx context.Context, cred clients.Credentials) ([]models.Curso, error) {
	var publicados, borradores []models.Course
	if err := s.education.GetJSON(ctx, "/api/cursos", cred, &publicados); err != nil {
		return nil, err
	}
	if err := s.education.GetJSON(ctx, "/api/cursos/borradores", cred, &borradores); err != nil {
		return nil, err
	}
	out := make([]models.Curso, 0, len(publicados)+len(borradores))
	for _, c := range publicados {
		out = append(out, MapCurso(c, "activo"))
	}
	for _, c := range borradores {
		out = append(out, MapCurso(c, "borrador"))
	}
	return out, nil
}

func MapCurso(c models.Course, estado string) models.Curso {
	curso := models.Curso{ID: c.ID, Titulo: c.Title, Estado: estado, Actividades: []models.Actividad{}}
	if c.TrainingArea != nil {
		curso.AreaID = c.TrainingArea.ID
	}
	if t, err := time.Parse(time.RFC3339, c.CreatedAt); err == nil {
		curso.FechaCreacion = t
	}
	return curso
}

func (s *dashboardService) Cursos(ctx context.Context, cred clients.Credentials) ([]models.Curso, error) {
	cursos, err := s.cursosBase(ctx, cred)
	if err != nil {
		return nil, err
	}
	err = enParalelo(ctx, len(cursos), func(ctx context.Context, i int) error {
		var acts []models.Activity
		path := "/api/cursos/" + url.PathEscape(cursos[i].ID) + "/actividades"
		if err := s.education.GetJSON(ctx, path, cred, &acts); err != nil {
			return err
		}
		for _, a := range acts {
			cursos[i].Actividades = append(cursos[i].Actividades, models.Actividad{ID: a.ID, Fecha: a.Fecha})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cursos, nil
}

// ── Habilidades (Educación) ───────────────────────────────────────────

func (s *dashboardService) Habilidades(ctx context.Context, cred clients.Credentials) ([]models.ParticipanteHabilidad, error) {
	var otorgadas []models.GrantedSkill
	if err := s.education.GetJSON(ctx, "/api/habilidades/otorgadas", cred, &otorgadas); err != nil {
		return nil, err
	}
	// El área sale del curso en que se ganó la habilidad.
	cursos, err := s.cursosBase(ctx, cred)
	if err != nil {
		return nil, err
	}
	areaPorCurso := make(map[string]string, len(cursos))
	for _, c := range cursos {
		areaPorCurso[c.ID] = c.AreaID
	}
	out := make([]models.ParticipanteHabilidad, len(otorgadas))
	for i, g := range otorgadas {
		out[i] = models.ParticipanteHabilidad{
			ID:             g.ParticipanteID + "_" + g.HabilidadID,
			ParticipanteID: g.ParticipanteID,
			CursoID:        g.CursoID,
			AreaID:         areaPorCurso[g.CursoID],
			Origen:         g.Origen,
			// Por diseño solo se ganan habilidades que el admin, el
			// docente o el voluntariado guardó en la actividad.
			Estado: "aceptada",
			Fecha:  g.Fecha,
		}
	}
	return out, nil
}

// ── Fichas e hitos (Promoción Humana) ─────────────────────────────────

func (s *dashboardService) Promocion(ctx context.Context, cred clients.Credentials) (models.Promocion, error) {
	var res models.IndicadoresPH
	if err := s.promocion.GetJSON(ctx, "/api/v1/human-promotion/indicadores", cred, &res); err != nil {
		return models.Promocion{}, err
	}
	return MapPromocion(res), nil
}

// MapPromocion traduce a los nombres que usa el frontend. Cualquier
// estado distinto de "baja" cuenta como activa.
func MapPromocion(res models.IndicadoresPH) models.Promocion {
	out := models.Promocion{Fichas: []models.FichaPromocion{}, Hitos: []models.HitoPromocion{}}
	for _, f := range res.Fichas {
		estado := "activa"
		if f.Estado == "baja" {
			estado = "baja"
		}
		out.Fichas = append(out.Fichas, models.FichaPromocion{ID: f.ID, ParticipanteID: f.ParticipanteID, Estado: estado, FechaAlta: f.CreadaEn, FechaBaja: f.FechaBaja})
	}
	for _, h := range res.Hitos {
		out.Hitos = append(out.Hitos, models.HitoPromocion{ParticipanteID: h.ParticipanteID, Fecha: h.Fecha})
	}
	return out
}

// ── Postulaciones (Inserción Laboral) ─────────────────────────────────

func (s *dashboardService) Postulaciones(ctx context.Context, cred clients.Credentials) ([]models.Postulacion, error) {
	var ofertas []models.JobOffer
	if err := s.labor.GetJSON(ctx, "/api/v1/labor-integration/offers", cred, &ofertas); err != nil {
		return nil, err
	}
	porOferta := make([][]models.JobApplication, len(ofertas))
	err := enParalelo(ctx, len(ofertas), func(ctx context.Context, i int) error {
		path := "/api/v1/labor-integration/offers/" + url.PathEscape(ofertas[i].ID) + "/applications"
		return s.labor.GetJSON(ctx, path, cred, &porOferta[i])
	})
	if err != nil {
		return nil, err
	}
	out := []models.Postulacion{}
	for _, apps := range porOferta {
		for _, a := range apps {
			out = append(out, MapPostulacion(a))
		}
	}
	return out, nil
}

// MapPostulacion traduce el estado que usa Inserción Laboral. La empresa
// marca "Aceptado" a quien elige: eso es una inserción ("seleccionado").
func MapPostulacion(a models.JobApplication) models.Postulacion {
	estado := "postulado"
	switch strings.ToLower(strings.TrimSpace(a.Status)) {
	case "aceptado", "seleccionado", "contratado":
		estado = "seleccionado"
	case "rechazado":
		estado = "rechazado"
	case "en revisión", "en revision", "en_revision":
		estado = "en_revision"
	}
	fecha := a.UpdatedAt
	if fecha.IsZero() {
		fecha = a.ApplicationDate
	}
	return models.Postulacion{ID: a.ID, ParticipanteID: a.ParticipantID, Estado: estado, FechaEstado: fecha}
}

// enParalelo corre fn para 0..n-1 con hasta maxParalelo a la vez y
// devuelve el primer error (cancela el resto).
func enParalelo(ctx context.Context, n int, fn func(ctx context.Context, i int) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sem := make(chan struct{}, maxParalelo)
	var wg sync.WaitGroup
	var once sync.Once
	var primero error
	for i := 0; i < n; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := fn(ctx, i); err != nil {
				once.Do(func() {
					primero = err
					cancel()
				})
			}
		}(i)
	}
	wg.Wait()
	return primero
}
