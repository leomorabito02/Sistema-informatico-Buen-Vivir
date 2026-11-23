package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"buenvivir/human-promotion/internal/models"
)

// fakeAuth responde como auth-service sin levantar gRPC.
type fakeAuth struct{}

func (fakeAuth) GetParticipante(context.Context, string) (*models.ParticipanteInfo, error) {
	return &models.ParticipanteInfo{Nombre: "Ana", Apellido: "Pérez"}, nil
}
func (fakeAuth) GetParticipanteByUserID(context.Context, string) (*models.ParticipanteInfo, error) {
	return nil, errors.New("no usado")
}
func (fakeAuth) SearchParticipantes(context.Context, string, int) ([]models.ParticipanteInfo, error) {
	return nil, nil
}
func (fakeAuth) GetNombresUsuarios(context.Context, []string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (fakeAuth) Close() error { return nil }

type fakeSkills struct {
	nombres []string
	err     error
}

func (f fakeSkills) HabilidadesDe(context.Context, string) ([]string, error) { return f.nombres, f.err }

func TestCompletarDetalle_Habilidades(t *testing.T) {
	ficha := &models.FichaProcesoPersonal{ID: "f1", ParticipanteID: "p1", UpdatedAt: time.Now()}

	con := &fichaService{auth: fakeAuth{}, skills: fakeSkills{nombres: []string{"Compostaje"}}, now: time.Now}
	if got := con.completarDetalle(context.Background(), ficha).Habilidades; len(got) != 1 || got[0] != "Compostaje" {
		t.Errorf("habilidades de Educación no llegaron a la ficha: %v", got)
	}

	// Educación caída: la ficha sale igual, sin habilidades.
	caida := &fichaService{auth: fakeAuth{}, skills: fakeSkills{err: errors.New("timeout")}, now: time.Now}
	d := caida.completarDetalle(context.Background(), ficha)
	if d.Habilidades == nil || len(d.Habilidades) != 0 {
		t.Errorf("se esperaba lista vacía, no %v", d.Habilidades)
	}
	if d.Participante == nil || d.Participante.Nombre != "Ana" {
		t.Error("los datos de Auth tienen que seguir apareciendo")
	}

	// Sin EDUCATION_SERVICE_URL (skills nil): la ficha sale sin habilidades.
	sin := &fichaService{auth: fakeAuth{}, now: time.Now}
	if got := sin.completarDetalle(context.Background(), ficha).Habilidades; len(got) != 0 {
		t.Errorf("sin cliente de Educación no debería haber habilidades: %v", got)
	}
}
