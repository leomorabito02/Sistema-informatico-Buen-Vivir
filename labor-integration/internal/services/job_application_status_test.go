package services

import (
	"context"
	"testing"

	"buenvivir/labor-integration/internal/models"
)

type mockProducer struct {
	emittedCount int
	lastPostID   string
	lastOfferID  string
	lastPartID   string
	lastEmpID    string
	notified     bool
}

func (m *mockProducer) OfertaCreada(context.Context, string, string, string) error { return nil }
func (m *mockProducer) OfertaRevisada(context.Context, string, string, bool) error  { return nil }
func (m *mockProducer) PostulacionCreada(context.Context, string, string, string, string) error {
	return nil
}
func (m *mockProducer) PostulacionDecidida(context.Context, string, string, string, string, string) error {
	return nil
}
func (m *mockProducer) PostulacionSeleccionada(_ context.Context, postID, offerID, partID, empID string) error {
	m.emittedCount++
	m.lastPostID = postID
	m.lastOfferID = offerID
	m.lastPartID = partID
	m.lastEmpID = empID
	return nil
}

func (m *mockProducer) OfertaEliminada(context.Context, string, string, bool) error { return nil }
func (m *mockProducer) OfertaCompatible(context.Context, string, []string) error  { return nil }
func (m *mockProducer) Enabled() bool { return true }
func (m *mockProducer) Notify()       { m.notified = true }

type fakeOfferRepo struct {
	offers map[string]*models.JobOffer
}

func (f *fakeOfferRepo) Create(context.Context, *models.JobOffer) error { return nil }
func (f *fakeOfferRepo) List(context.Context, string) ([]*models.JobOffer, error) {
	return nil, nil
}
func (f *fakeOfferRepo) Update(context.Context, string, *models.JobOffer) error { return nil }
func (f *fakeOfferRepo) Delete(context.Context, string) error                   { return nil }
func (f *fakeOfferRepo) UpdateStatus(context.Context, string, string) error    { return nil }
func (f *fakeOfferRepo) UpdateStatusAndApprover(context.Context, string, string, string) error {
	return nil
}
func (f *fakeOfferRepo) GetByID(_ context.Context, id string) (*models.JobOffer, error) {
	if o, ok := f.offers[id]; ok {
		return o, nil
	}
	return nil, nil
}

func TestUpdateStatus_EmitePostulacionSeleccionada(t *testing.T) {
	appRepo := &fakeApps{apps: []*models.JobApplication{
		{ID: "app-1", OfferID: "off-1", ParticipantID: "part-1", UserID: "usr-1", Status: "Pendiente"},
	}}
	offerRepo := &fakeOfferRepo{offers: map[string]*models.JobOffer{
		"off-1": {ID: "off-1", CompanyID: "empresa-1", Title: "Desarrollador Go"},
	}}
	prod := &mockProducer{}

	svc := NewJobApplicationService(
		appRepo,
		offerRepo,
		fakeAuth{},
		WithApplicationProducer(prod),
	)

	// 1. Transición a "Aceptado" desde "Pendiente": DEBE emitir
	err := svc.UpdateStatus(context.Background(), "app-1", "empresa-1", "Aceptado")
	if err != nil {
		t.Fatalf("UpdateStatus falló: %v", err)
	}

	if prod.emittedCount != 1 {
		t.Fatalf("se esperaba 1 emisión, se obtuvieron: %d", prod.emittedCount)
	}
	if prod.lastPostID != "app-1" || prod.lastOfferID != "off-1" || prod.lastPartID != "part-1" || prod.lastEmpID != "empresa-1" {
		t.Errorf("datos emitidos incorrectos: %+v", prod)
	}
	if !prod.notified {
		t.Error("se esperaba llamada a Notify()")
	}

	// 2. Transición de "Aceptado" a "Contratado" (ambos seleccionados): NO DEBE volver a emitir
	appRepo.apps[0].Status = "Aceptado"
	err = svc.UpdateStatus(context.Background(), "app-1", "empresa-1", "Contratado")
	if err != nil {
		t.Fatalf("UpdateStatus falló: %v", err)
	}
	if prod.emittedCount != 1 {
		t.Errorf("no debía reemitir al pasar entre estados seleccionados, conteo: %d", prod.emittedCount)
	}

	// 3. Transición a "Rechazado": NO DEBE emitir
	appRepo.apps[0].Status = "Pendiente"
	err = svc.UpdateStatus(context.Background(), "app-1", "empresa-1", "Rechazado")
	if err != nil {
		t.Fatalf("UpdateStatus falló: %v", err)
	}
	if prod.emittedCount != 1 {
		t.Errorf("no debía emitir para Rechazado, conteo: %d", prod.emittedCount)
	}
}
