package services

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

// ofertaCompleta tiene todos los campos que pide la HU-39.
func ofertaCompleta(titulo string) *models.JobOffer {
	return &models.JobOffer{
		Title: titulo, Description: "Atención al público", Location: "Alta Gracia",
		Modality: "Presencial", JobType: "Medio tiempo", Salary: "A convenir",
		Schedule: "Lunes a viernes 9 a 13", Requirements: []string{"Atención al cliente"},
		Deadline: time.Now().Add(30 * 24 * time.Hour),
	}
}

// HU-39: cada campo obligatorio faltante responde ErrOfferInvalid.
func TestValidateOfferFields(t *testing.T) {
	if err := validateOfferFields(ofertaCompleta("Cajero"), true); err != nil {
		t.Fatalf("oferta completa: %v", err)
	}
	cambios := map[string]func(o *models.JobOffer){
		"ubicacion":  func(o *models.JobOffer) { o.Location = " " },
		"modalidad":  func(o *models.JobOffer) { o.Modality = "" },
		"tipo":       func(o *models.JobOffer) { o.JobType = "" },
		"salario":    func(o *models.JobOffer) { o.Salary = "" },
		"horario":    func(o *models.JobOffer) { o.Schedule = "" },
		"requisitos": func(o *models.JobOffer) { o.Requirements = []string{" "} },
		"fecha":      func(o *models.JobOffer) { o.Deadline = time.Time{} },
		"vencida":    func(o *models.JobOffer) { o.Deadline = time.Now().AddDate(0, 0, -3) },
	}
	for name, cambio := range cambios {
		o := ofertaCompleta("Cajero")
		cambio(o)
		if err := validateOfferFields(o, true); !errors.Is(err, ErrOfferInvalid) {
			t.Errorf("%s: esperaba ErrOfferInvalid, dio %v", name, err)
		}
	}
}

// HU-44 (decisión 15/11): requisito = nombre de habilidad, sin mayúsculas
// ni acentos; cada participante una sola vez.
func TestMatchCompatibles(t *testing.T) {
	skills := []repositories.GrantedSkill{
		{ParticipanteID: "part-2", Nombre: "Atención al cliente"},
		{ParticipanteID: "part-1", Nombre: "ATENCION  AL CLIENTE"},
		{ParticipanteID: "part-1", Nombre: "Trabajo en equipo"},
		{ParticipanteID: "part-3", Nombre: "Carpintería"},
	}
	got := MatchCompatibles([]string{"atención al cliente", "Trabajo en equipo"}, skills)
	if want := []string{"part-1", "part-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got := MatchCompatibles([]string{"Electricidad"}, skills); len(got) != 0 {
		t.Fatalf("no debería haber compatibles: %v", got)
	}
}

type fakeSkills struct {
	token string
	out   []repositories.GrantedSkill
}

func (f *fakeSkills) GrantedSkills(_ context.Context, bearer string) ([]repositories.GrantedSkill, error) {
	f.token = bearer
	return f.out, nil
}

// HU-44: al aprobar, se consulta Educación con el token del admin y se
// emite oferta.compatible con los participantes que coinciden.
func TestApproveOffer_EmiteCompatibles(t *testing.T) {
	repo := newMockOfferRepo()
	o := ofertaCompleta("Cajero")
	o.ID, o.CompanyID, o.Status = "off-1", "usr-emp", "Pendiente"
	repo.offers["off-1"] = o
	prod := &mockLaborProducer{}
	skills := &fakeSkills{out: []repositories.GrantedSkill{{ParticipanteID: "part-9", Nombre: "atencion al cliente"}}}
	svc := NewJobOfferService(repo, &fakeAppsWithGetByID{apps: map[string]*models.JobApplication{}}, nil, WithOfferProducer(prod), WithSkillsClient(skills))

	if err := svc.ApproveOffer(WithBearer(context.Background(), "tok-admin"), "off-1", "usr-admin"); err != nil {
		t.Fatal(err)
	}
	if skills.token != "tok-admin" {
		t.Errorf("no reenvió el token del admin: %q", skills.token)
	}
	if !reflect.DeepEqual(prod.compatibleIDs, []string{"part-9"}) {
		t.Errorf("compatibles: %v", prod.compatibleIDs)
	}
}

// HU-41: borrar avisa a la empresa (por_gestion=true si lo hizo el admin).
func TestDeleteOffer_EmiteEliminada(t *testing.T) {
	repo := newMockOfferRepo()
	repo.offers["off-1"] = &models.JobOffer{ID: "off-1", CompanyID: "usr-emp", Status: "Activa"}
	prod := &mockLaborProducer{}
	svc := NewJobOfferService(repo, &fakeAppsWithGetByID{apps: map[string]*models.JobApplication{}}, nil, WithOfferProducer(prod))

	if err := svc.DeleteOffer(context.Background(), "off-1", "role_admin", "usr-admin"); err != nil {
		t.Fatal(err)
	}
	if prod.deletedOfferID != "off-1" || !prod.deletedGestion {
		t.Errorf("evento inesperado: %q %v", prod.deletedOfferID, prod.deletedGestion)
	}
}
