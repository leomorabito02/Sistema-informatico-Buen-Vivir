package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/models"
)

type countingService struct {
	calls map[Fuente]int
	err   error
}

func (s *countingService) hit(f Fuente) error {
	if s.calls == nil {
		s.calls = map[Fuente]int{}
	}
	s.calls[f]++
	return s.err
}

func (s *countingService) Participantes(context.Context, clients.Credentials) ([]models.Participante, error) {
	return []models.Participante{{ID: "p1"}}, s.hit(FuenteParticipantes)
}
func (s *countingService) Cursos(context.Context, clients.Credentials) ([]models.Curso, error) {
	return []models.Curso{{}}, s.hit(FuenteCursos)
}
func (s *countingService) Habilidades(context.Context, clients.Credentials) ([]models.ParticipanteHabilidad, error) {
	return nil, s.hit(FuenteHabilidades)
}
func (s *countingService) Postulaciones(context.Context, clients.Credentials) ([]models.Postulacion, error) {
	return nil, s.hit(FuentePostulaciones)
}
func (s *countingService) Areas(context.Context, clients.Credentials) ([]models.Area, error) {
	return nil, s.hit(FuenteAreas)
}

func TestCacheReusaEInvalida(t *testing.T) {
	inner := &countingService{}
	c := NewCachedDashboardService(inner, time.Minute)
	ctx, cred := context.Background(), clients.Credentials{}

	for i := 0; i < 3; i++ {
		got, err := c.Participantes(ctx, cred)
		if err != nil || len(got) != 1 || got[0].ID != "p1" {
			t.Fatalf("participantes: %v %v", got, err)
		}
		_, _ = c.Cursos(ctx, cred)
		_, _ = c.Habilidades(ctx, cred)
		_, _ = c.Postulaciones(ctx, cred)
		_, _ = c.Areas(ctx, cred)
	}
	for _, f := range []Fuente{FuenteParticipantes, FuenteCursos, FuenteHabilidades, FuentePostulaciones, FuenteAreas} {
		if inner.calls[f] != 1 {
			t.Errorf("%s: %d consultas al origen", f, inner.calls[f])
		}
	}

	c.Invalidate(FuenteParticipantes)
	_, _ = c.Participantes(ctx, cred)
	_, _ = c.Cursos(ctx, cred)
	if inner.calls[FuenteParticipantes] != 2 || inner.calls[FuenteCursos] != 1 {
		t.Errorf("invalidar una fuente no toca las otras: %v", inner.calls)
	}

	c.Invalidate()
	_, _ = c.Cursos(ctx, cred)
	if inner.calls[FuenteCursos] != 2 {
		t.Error("sin argumentos invalida todo")
	}

	now := time.Now().Add(2 * time.Minute)
	c.now = func() time.Time { return now }
	_, _ = c.Cursos(ctx, cred)
	if inner.calls[FuenteCursos] != 3 {
		t.Error("vencido el TTL se vuelve a consultar")
	}
}

func TestCacheNoGuardaErroresNiConTTLCero(t *testing.T) {
	inner := &countingService{err: errors.New("origen caído")}
	c := NewCachedDashboardService(inner, time.Minute)
	for i := 0; i < 2; i++ {
		if _, err := c.Areas(context.Background(), clients.Credentials{}); err == nil {
			t.Fatal("el error se propaga")
		}
	}
	if inner.calls[FuenteAreas] != 2 {
		t.Error("los errores no se cachean")
	}

	off := NewCachedDashboardService(&countingService{}, 0)
	_, _ = off.Areas(context.Background(), clients.Credentials{})
	_, _ = off.Areas(context.Background(), clients.Credentials{})
	if off.inner.(*countingService).calls[FuenteAreas] != 2 {
		t.Error("TTL 0 desactiva la caché")
	}
}
