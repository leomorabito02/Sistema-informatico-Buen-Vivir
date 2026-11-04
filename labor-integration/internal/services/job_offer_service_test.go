package services

import (
	"context"
	"testing"

	"buenvivir/labor-integration/internal/models"
)

type mockOfferRepo struct {
	offers map[string]*models.JobOffer
}

func newMockOfferRepo() *mockOfferRepo {
	return &mockOfferRepo{offers: make(map[string]*models.JobOffer)}
}

func (m *mockOfferRepo) Create(ctx context.Context, offer *models.JobOffer) error {
	offer.ID = "off-1"
	m.offers[offer.ID] = offer
	return nil
}

func (m *mockOfferRepo) List(ctx context.Context, companyID string) ([]*models.JobOffer, error) {
	var list []*models.JobOffer
	for _, o := range m.offers {
		list = append(list, o)
	}
	return list, nil
}

func (m *mockOfferRepo) Update(ctx context.Context, id string, offer *models.JobOffer) error {
	m.offers[id] = offer
	return nil
}

func (m *mockOfferRepo) Delete(ctx context.Context, id string) error {
	delete(m.offers, id)
	return nil
}

func (m *mockOfferRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	if o, ok := m.offers[id]; ok {
		o.Status = status
	}
	return nil
}

func (m *mockOfferRepo) UpdateStatusAndApprover(ctx context.Context, id string, status string, approvedBy string) error {
	if o, ok := m.offers[id]; ok {
		o.Status = status
		o.ApprovedBy = approvedBy
	}
	return nil
}

func (m *mockOfferRepo) GetByID(ctx context.Context, id string) (*models.JobOffer, error) {
	return m.offers[id], nil
}

type mockLaborProducer struct {
	createdOfferID  string
	createdEmpresa  string
	createdEstado   string
	reviewedOfferID string
	reviewedEmpresa string
	reviewedApprove bool
	appliedPostID   string
	appliedOfferID  string
	appliedPartID   string
	decidedPostID   string
	decidedResult   string
	notified        bool
}

func (m *mockLaborProducer) Enabled() bool { return true }
func (m *mockLaborProducer) Notify()       { m.notified = true }
func (m *mockLaborProducer) OfertaCreada(ctx context.Context, ofertaID, empresaID, estado string) error {
	m.createdOfferID = ofertaID
	m.createdEmpresa = empresaID
	m.createdEstado = estado
	return nil
}
func (m *mockLaborProducer) OfertaRevisada(ctx context.Context, ofertaID, empresaID string, aprobada bool) error {
	m.reviewedOfferID = ofertaID
	m.reviewedEmpresa = empresaID
	m.reviewedApprove = aprobada
	return nil
}
func (m *mockLaborProducer) PostulacionCreada(ctx context.Context, postulacionID, ofertaID, participanteID, empresaID string) error {
	m.appliedPostID = postulacionID
	m.appliedOfferID = ofertaID
	m.appliedPartID = participanteID
	return nil
}
func (m *mockLaborProducer) PostulacionSeleccionada(ctx context.Context, postulacionID, ofertaID, participanteID, empresaID string) error {
	return nil
}
func (m *mockLaborProducer) PostulacionDecidida(ctx context.Context, postulacionID, ofertaID, participanteID, empresaID, resultado string) error {
	m.decidedPostID = postulacionID
	m.decidedResult = resultado
	return nil
}

func TestJobOfferService_NotificationDispatching(t *testing.T) {
	// Arrange
	repo := newMockOfferRepo()
	producer := &mockLaborProducer{}
	svc := NewJobOfferService(
		repo,
		nil,
		fakeAuth{},
		WithOfferProducer(producer),
	)

	// Act 1: Create offer as admin
	offer := &models.JobOffer{
		ID:    "off-1",
		Title: "Recepcionista",
	}
	err := svc.CreateOffer(context.Background(), offer, "role_admin", "usr-admin-1")
	if err != nil {
		t.Fatalf("unexpected error creating offer: %v", err)
	}

	// Assert 1
	if producer.createdOfferID != "off-1" || producer.createdEstado != "Activa" {
		t.Errorf("expected producer.OfertaCreada with off-1 Activa, got %s %s", producer.createdOfferID, producer.createdEstado)
	}
	if offer.CompanyID != "usr-admin-1" {
		t.Errorf("expected creator admin to be assigned as CompanyID, got %s", offer.CompanyID)
	}

	// Act 2: Approve offer
	err = svc.ApproveOffer(context.Background(), "off-1", "usr-admin-1")
	if err != nil {
		t.Fatalf("unexpected error approving offer: %v", err)
	}

	// Assert 2
	if producer.reviewedOfferID != "off-1" || !producer.reviewedApprove {
		t.Errorf("expected producer.OfertaRevisada with approved=true")
	}

	// Act 3: Reject offer
	err = svc.RejectOffer(context.Background(), "off-1", "usr-admin-1")
	if err != nil {
		t.Fatalf("unexpected error rejecting offer: %v", err)
	}

	// Assert 3
	if producer.reviewedOfferID != "off-1" || producer.reviewedApprove {
		t.Errorf("expected producer.OfertaRevisada with approved=false")
	}
}

func TestJobApplicationService_NotificationDispatching(t *testing.T) {
	// Arrange
	offerRepo := newMockOfferRepo()
	offerRepo.offers["off-1"] = &models.JobOffer{
		ID:        "off-1",
		CompanyID: "comp-1",
		Title:     "Pintor",
		Status:    "Activa",
	}

	appRepo := &fakeAppsWithGetByID{
		apps: make(map[string]*models.JobApplication),
	}
	producer := &mockLaborProducer{}

	svc := NewJobApplicationService(
		appRepo,
		offerRepo,
		fakeAuth{},
		WithApplicationProducer(producer),
	)

	// Act 1: Apply to job
	app := &models.JobApplication{
		OfferID:       "off-1",
		ParticipantID: "part-1",
	}
	err := svc.ApplyToJob(context.Background(), app)
	if err != nil {
		t.Fatalf("unexpected error applying: %v", err)
	}

	// Assert 1
	if producer.appliedOfferID != "off-1" || producer.appliedPartID != "part-1" {
		t.Fatalf("expected PostulacionCreada to be emitted")
	}

	// Act 2: Update status
	err = svc.UpdateStatus(context.Background(), "app-1", "comp-1", "Aceptado")
	if err != nil {
		t.Fatalf("unexpected error updating status: %v", err)
	}

	// Assert 2
	if producer.decidedPostID != "app-1" || producer.decidedResult != "Aceptado" {
		t.Errorf("expected PostulacionDecidida with status Aceptado, got %s", producer.decidedResult)
	}
}

type fakeAppsWithGetByID struct {
	apps map[string]*models.JobApplication
}

func (f *fakeAppsWithGetByID) Create(_ context.Context, a *models.JobApplication) error {
	a.ID = "app-1"
	f.apps[a.ID] = a
	return nil
}
func (f *fakeAppsWithGetByID) ListByOffer(context.Context, string) ([]*models.JobApplication, error) {
	var list []*models.JobApplication
	for _, a := range f.apps {
		list = append(list, a)
	}
	return list, nil
}
func (f *fakeAppsWithGetByID) HasApplicationsForOffer(context.Context, string) (bool, error) {
	return len(f.apps) > 0, nil
}
func (f *fakeAppsWithGetByID) HasApplied(_ context.Context, offerID string, ids []string) (bool, error) {
	return false, nil
}
func (f *fakeAppsWithGetByID) ListByParticipant(_ context.Context, _ []string) ([]*models.JobApplication, error) {
	var list []*models.JobApplication
	for _, a := range f.apps {
		list = append(list, a)
	}
	return list, nil
}
func (f *fakeAppsWithGetByID) UpdateStatus(_ context.Context, id string, status string) error {
	if a, ok := f.apps[id]; ok {
		a.Status = status
	}
	return nil
}
func (f *fakeAppsWithGetByID) Delete(_ context.Context, id string) error {
	delete(f.apps, id)
	return nil
}
func (f *fakeAppsWithGetByID) GetByID(_ context.Context, id string) (*models.JobApplication, error) {
	return f.apps[id], nil
}
