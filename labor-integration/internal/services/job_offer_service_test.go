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

type mockDispatcher struct {
	createdOffer     *models.JobOffer
	createdActor     string
	reviewedOffer    *models.JobOffer
	reviewedApproved bool
	receivedApp      *models.JobApplication
	receivedOffer    *models.JobOffer
	decidedApp       *models.JobApplication
	decidedOffer     *models.JobOffer
	decidedStatus    string
}

func (m *mockDispatcher) OfferCreated(ctx context.Context, offer *models.JobOffer, actorID string) {
	m.createdOffer = offer
	m.createdActor = actorID
}

func (m *mockDispatcher) OfferReviewed(ctx context.Context, offer *models.JobOffer, approved bool) {
	m.reviewedOffer = offer
	m.reviewedApproved = approved
}

func (m *mockDispatcher) ApplicationReceived(ctx context.Context, app *models.JobApplication, offer *models.JobOffer, actorID string) {
	m.receivedApp = app
	m.receivedOffer = offer
}

func (m *mockDispatcher) ApplicationDecided(ctx context.Context, app *models.JobApplication, offer *models.JobOffer, status string) {
	m.decidedApp = app
	m.decidedOffer = offer
	m.decidedStatus = status
}

func TestJobOfferService_NotificationDispatching(t *testing.T) {
	// Arrange
	repo := newMockOfferRepo()
	dispatcher := &mockDispatcher{}
	svc := NewJobOfferService(
		repo,
		nil,
		fakeAuth{},
		WithOfferDispatcher(dispatcher),
		WithOfferSyncDispatch(true),
	)

	// Act 1: Create offer as admin
	offer := &models.JobOffer{
		Title: "Recepcionista",
	}
	err := svc.CreateOffer(context.Background(), offer, "role_admin", "usr-admin-1")
	if err != nil {
		t.Fatalf("unexpected error creating offer: %v", err)
	}

	// Assert 1
	if dispatcher.createdOffer == nil || dispatcher.createdOffer.Title != "Recepcionista" {
		t.Errorf("expected dispatcher.OfferCreated to be called with offer")
	}
	if dispatcher.createdActor != "usr-admin-1" {
		t.Errorf("expected actor usr-admin-1, got %s", dispatcher.createdActor)
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
	if dispatcher.reviewedOffer == nil || !dispatcher.reviewedApproved {
		t.Errorf("expected dispatcher.OfferReviewed with approved=true")
	}

	// Act 3: Reject offer
	err = svc.RejectOffer(context.Background(), "off-1", "usr-admin-1")
	if err != nil {
		t.Fatalf("unexpected error rejecting offer: %v", err)
	}

	// Assert 3
	if dispatcher.reviewedOffer == nil || dispatcher.reviewedApproved {
		t.Errorf("expected dispatcher.OfferReviewed with approved=false")
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
	dispatcher := &mockDispatcher{}

	svc := NewJobApplicationService(
		appRepo,
		offerRepo,
		fakeAuth{},
		WithApplicationDispatcher(dispatcher),
		WithApplicationSyncDispatch(true),
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
	if dispatcher.receivedApp == nil || dispatcher.receivedOffer == nil {
		t.Fatalf("expected ApplicationReceived to be called")
	}
	if dispatcher.receivedOffer.ID != "off-1" {
		t.Errorf("expected offer ID off-1, got %s", dispatcher.receivedOffer.ID)
	}

	// Act 2: Update status
	err = svc.UpdateStatus(context.Background(), "app-1", "comp-1", "Aceptado")
	if err != nil {
		t.Fatalf("unexpected error updating status: %v", err)
	}

	// Assert 2
	if dispatcher.decidedApp == nil || dispatcher.decidedStatus != "Aceptado" {
		t.Errorf("expected ApplicationDecided with status Aceptado, got %s", dispatcher.decidedStatus)
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
