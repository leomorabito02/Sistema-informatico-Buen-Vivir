package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"buenvivir/labor-integration/api/proto/authpb"
	"buenvivir/labor-integration/internal/models"
)

type mockAuthForDispatcher struct {
	adminVolunteerIDs []string
	participant       *authpb.GetParticipantResponse
}

func (m mockAuthForDispatcher) GetParticipant(ctx context.Context, participantID string) (*authpb.GetParticipantResponse, error) {
	if m.participant != nil {
		return m.participant, nil
	}
	return &authpb.GetParticipantResponse{
		Id:     participantID,
		UserId: "usr-" + participantID,
	}, nil
}

func (m mockAuthForDispatcher) GetCompany(ctx context.Context, companyID string) (*authpb.GetCompanyResponse, error) {
	return &authpb.GetCompanyResponse{}, nil
}

func (m mockAuthForDispatcher) GetActiveUserIDsByRoles(ctx context.Context, roles []string) ([]string, error) {
	return m.adminVolunteerIDs, nil
}

func (m mockAuthForDispatcher) Close() error {
	return nil
}

func TestNotificationDispatcher_OfferCreated(t *testing.T) {
	// Arrange
	var (
		mu           sync.Mutex
		receivedReqs []notificationPayload
		receivedHmac string
	)
	secret := "test-secret-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		receivedHmac = r.Header.Get("X-Internal-Signature")
		body, _ := io.ReadAll(r.Body)

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedSig := hex.EncodeToString(mac.Sum(nil))
		if receivedHmac != expectedSig {
			t.Errorf("HMAC mismatch: expected %s, got %s", expectedSig, receivedHmac)
		}

		var payload notificationPayload
		_ = json.Unmarshal(body, &payload)
		receivedReqs = append(receivedReqs, payload)

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	auth := mockAuthForDispatcher{
		adminVolunteerIDs: []string{"usr-admin-1", "usr-vol-1", "usr-creator"},
	}
	dispatcher := NewHTTPNotificationDispatcher(server.URL, secret, auth)

	offer := &models.JobOffer{
		ID:        "off-123",
		CompanyID: "comp-456",
		Title:     "Desarrollador Go",
		Status:    "Pendiente",
	}

	// Act: creator is usr-creator, so it should be excluded
	dispatcher.OfferCreated(context.Background(), offer, "usr-creator")

	// Assert
	mu.Lock()
	defer mu.Unlock()
	if len(receivedReqs) != 1 {
		t.Fatalf("expected 1 notification sent, got %d", len(receivedReqs))
	}
	notif := receivedReqs[0]
	if notif.Category != "labor" {
		t.Errorf("expected category 'labor', got %s", notif.Category)
	}
	if notif.EventType != "offer.created" {
		t.Errorf("expected event_type 'offer.created', got %s", notif.EventType)
	}
	if len(notif.RecipientUserIDs) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(notif.RecipientUserIDs))
	}
	if notif.RecipientUserIDs[0] != "usr-admin-1" || notif.RecipientUserIDs[1] != "usr-vol-1" {
		t.Errorf("unexpected recipients: %v", notif.RecipientUserIDs)
	}
}

func TestNotificationDispatcher_OfferReviewed(t *testing.T) {
	// Arrange
	var (
		mu           sync.Mutex
		receivedReqs []notificationPayload
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		var payload notificationPayload
		_ = json.NewDecoder(r.Body).Decode(&payload)
		receivedReqs = append(receivedReqs, payload)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	dispatcher := NewHTTPNotificationDispatcher(server.URL, "", mockAuthForDispatcher{})
	offer := &models.JobOffer{
		ID:        "off-789",
		CompanyID: "usr-comp-1",
		Title:     "Panadero",
	}

	// Act 1: Approved
	dispatcher.OfferReviewed(context.Background(), offer, true)

	// Act 2: Rejected
	dispatcher.OfferReviewed(context.Background(), offer, false)

	// Assert
	mu.Lock()
	defer mu.Unlock()
	if len(receivedReqs) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(receivedReqs))
	}
	if receivedReqs[0].EventType != "offer.approved" {
		t.Errorf("expected offer.approved, got %s", receivedReqs[0].EventType)
	}
	if receivedReqs[0].RecipientUserIDs[0] != "usr-comp-1" {
		t.Errorf("expected recipient usr-comp-1, got %v", receivedReqs[0].RecipientUserIDs)
	}
	if receivedReqs[1].EventType != "offer.rejected" {
		t.Errorf("expected offer.rejected, got %s", receivedReqs[1].EventType)
	}
}

func TestNotificationDispatcher_ApplicationReceived(t *testing.T) {
	// Arrange
	var (
		mu           sync.Mutex
		receivedReqs []notificationPayload
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		var payload notificationPayload
		_ = json.NewDecoder(r.Body).Decode(&payload)
		receivedReqs = append(receivedReqs, payload)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	auth := mockAuthForDispatcher{
		participant: &authpb.GetParticipantResponse{
			Id:     "part-999",
			UserId: "usr-part-999",
		},
	}
	dispatcher := NewHTTPNotificationDispatcher(server.URL, "", auth)

	offer := &models.JobOffer{
		ID:        "off-1",
		CompanyID: "usr-company-owner",
		Title:     "Electricista",
	}
	app := &models.JobApplication{
		ID:            "app-1",
		OfferID:       "off-1",
		ParticipantID: "part-999",
		UserID:        "usr-part-999",
	}

	// Act
	dispatcher.ApplicationReceived(context.Background(), app, offer, "usr-part-999")

	// Assert
	mu.Lock()
	defer mu.Unlock()
	if len(receivedReqs) != 2 {
		t.Fatalf("expected 2 notifications (company and participant), got %d", len(receivedReqs))
	}

	// Company notification
	if receivedReqs[0].RecipientUserIDs[0] != "usr-company-owner" {
		t.Errorf("expected recipient usr-company-owner, got %v", receivedReqs[0].RecipientUserIDs)
	}
	if receivedReqs[0].EventType != "application.received" {
		t.Errorf("expected application.received, got %s", receivedReqs[0].EventType)
	}
	// ADR-0007: Message must not contain participant name
	if receivedReqs[0].Message != "Un participante se postuló a «Electricista»." {
		t.Errorf("unexpected message: %s", receivedReqs[0].Message)
	}

	// Participant confirmation notification
	if receivedReqs[1].RecipientUserIDs[0] != "usr-part-999" {
		t.Errorf("expected recipient usr-part-999, got %v", receivedReqs[1].RecipientUserIDs)
	}
	if receivedReqs[1].EventType != "application.submitted" {
		t.Errorf("expected application.submitted, got %s", receivedReqs[1].EventType)
	}
}

func TestNotificationDispatcher_ApplicationDecided(t *testing.T) {
	// Arrange
	var (
		mu           sync.Mutex
		receivedReqs []notificationPayload
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		var payload notificationPayload
		_ = json.NewDecoder(r.Body).Decode(&payload)
		receivedReqs = append(receivedReqs, payload)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	auth := mockAuthForDispatcher{
		adminVolunteerIDs: []string{"usr-admin", "usr-vol"},
		participant: &authpb.GetParticipantResponse{
			Id:     "part-1",
			UserId: "usr-part-1",
		},
	}
	dispatcher := NewHTTPNotificationDispatcher(server.URL, "", auth)

	offer := &models.JobOffer{
		ID:        "off-1",
		CompanyID: "usr-comp",
		Title:     "Carpintero",
	}
	app := &models.JobApplication{
		ID:            "app-1",
		OfferID:       "off-1",
		ParticipantID: "part-1",
		UserID:        "usr-part-1",
	}

	// Act 1: Aceptado
	dispatcher.ApplicationDecided(context.Background(), app, offer, "Aceptado")

	// Act 2: Rechazado
	dispatcher.ApplicationDecided(context.Background(), app, offer, "Rechazado")

	// Act 3: Status not decisive (e.g. "Pendiente")
	dispatcher.ApplicationDecided(context.Background(), app, offer, "Pendiente")

	// Assert
	mu.Lock()
	defer mu.Unlock()
	// Act 1 sends 2 (participant + staff), Act 2 sends 2 (participant + staff), Act 3 sends 0 => 4 total
	if len(receivedReqs) != 4 {
		t.Fatalf("expected 4 notifications, got %d", len(receivedReqs))
	}

	// Check Accepted for participant
	if receivedReqs[0].EventType != "application.accepted" || receivedReqs[0].RecipientUserIDs[0] != "usr-part-1" {
		t.Errorf("unexpected accepted participant notification: %+v", receivedReqs[0])
	}
	// Check Accepted for staff
	if receivedReqs[1].EventType != "application.accepted.staff" || len(receivedReqs[1].RecipientUserIDs) != 2 {
		t.Errorf("unexpected accepted staff notification: %+v", receivedReqs[1])
	}

	// Check Rejected for participant
	if receivedReqs[2].EventType != "application.rejected" || receivedReqs[2].RecipientUserIDs[0] != "usr-part-1" {
		t.Errorf("unexpected rejected participant notification: %+v", receivedReqs[2])
	}
	// Check Rejected for staff
	if receivedReqs[3].EventType != "application.rejected.staff" || len(receivedReqs[3].RecipientUserIDs) != 2 {
		t.Errorf("unexpected rejected staff notification: %+v", receivedReqs[3])
	}
}

