package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

// NotificationDispatcher defines event notification dispatching for labor-integration.
type NotificationDispatcher interface {
	OfferCreated(ctx context.Context, offer *models.JobOffer, actorID string)
	OfferReviewed(ctx context.Context, offer *models.JobOffer, approved bool)
	ApplicationReceived(ctx context.Context, app *models.JobApplication, offer *models.JobOffer, actorID string)
	ApplicationDecided(ctx context.Context, app *models.JobApplication, offer *models.JobOffer, status string)
}

type noopNotificationDispatcher struct{}

// NewNoopNotificationDispatcher returns a dispatcher that performs no action.
func NewNoopNotificationDispatcher() NotificationDispatcher {
	return &noopNotificationDispatcher{}
}

func (d *noopNotificationDispatcher) OfferCreated(ctx context.Context, offer *models.JobOffer, actorID string) {}
func (d *noopNotificationDispatcher) OfferReviewed(ctx context.Context, offer *models.JobOffer, approved bool) {}
func (d *noopNotificationDispatcher) ApplicationReceived(ctx context.Context, app *models.JobApplication, offer *models.JobOffer, actorID string) {}
func (d *noopNotificationDispatcher) ApplicationDecided(ctx context.Context, app *models.JobApplication, offer *models.JobOffer, status string) {}

type httpNotificationDispatcher struct {
	baseURL    string
	secret     string
	authClient repositories.AuthServiceClient
	httpClient *http.Client
}

// NewHTTPNotificationDispatcher creates a new HTTP-based NotificationDispatcher.
func NewHTTPNotificationDispatcher(serviceURL, secret string, authClient repositories.AuthServiceClient) NotificationDispatcher {
	cleanURL := strings.TrimRight(strings.TrimSpace(serviceURL), "/")
	if cleanURL == "" {
		cleanURL = "http://notifications-service:8080"
	}

	return &httpNotificationDispatcher{
		baseURL:    cleanURL,
		secret:     strings.TrimSpace(secret),
		authClient: authClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// notificationPayload matches the ingestion DTO of notifications-service.
type notificationPayload struct {
	SourceService    string   `json:"source_service"`
	EventType        string   `json:"event_type"`
	Category         string   `json:"category"`
	Priority         string   `json:"priority"`
	SenderType       string   `json:"sender_type"`
	SenderID         string   `json:"sender_id"`
	TargetType       string   `json:"target_type"`
	RecipientUserIDs []string `json:"recipient_user_ids"`
	Title            string   `json:"title"`
	Message          string   `json:"message"`
	ActionURL        *string  `json:"action_url,omitempty"`
	Data             string   `json:"data,omitempty"`
	IsCritical       bool     `json:"is_critical"`
	SendEmail        bool     `json:"send_email"`
	SendPush         bool     `json:"send_push"`
	SendInApp        bool     `json:"send_in_app"`
}

func (d *httpNotificationDispatcher) OfferCreated(ctx context.Context, offer *models.JobOffer, actorID string) {
	if offer == nil {
		return
	}

	var adminVolunteerIDs []string
	if d.authClient != nil {
		ids, err := d.authClient.GetActiveUserIDsByRoles(ctx, []string{"role_admin", "role_volunteer"})
		if err != nil {
			slog.Warn("failed to fetch admin and volunteer IDs for offer creation notification", slog.String("error", err.Error()))
		} else {
			adminVolunteerIDs = ids
		}
	}

	recipients := filterAndDeduplicate(adminVolunteerIDs, actorID)
	if len(recipients) == 0 {
		return
	}

	actionURL := "/dashboard/ofertas"
	dataJSON, _ := json.Marshal(map[string]interface{}{
		"offer_id":   offer.ID,
		"title":      offer.Title,
		"company_id": offer.CompanyID,
		"action_url": actionURL,
	})

	var message string
	if offer.Status == "Activa" {
		message = fmt.Sprintf("Se ha publicado la oferta laboral «%s».", offer.Title)
	} else {
		message = fmt.Sprintf("Se ha registrado la oferta laboral «%s» pendiente de revisión.", offer.Title)
	}

	payload := notificationPayload{
		SourceService:    "labor-integration",
		EventType:        "offer.created",
		Category:         "labor",
		Priority:         "normal",
		SenderType:       "system",
		SenderID:         actorID,
		TargetType:       "role",
		RecipientUserIDs: recipients,
		Title:            "Nueva oferta laboral registrada",
		Message:          message,
		ActionURL:        &actionURL,
		Data:             string(dataJSON),
		IsCritical:       false,
		SendEmail:        true,
		SendPush:         true,
		SendInApp:        true,
	}

	_ = d.sendNotification(ctx, payload)
}

func (d *httpNotificationDispatcher) OfferReviewed(ctx context.Context, offer *models.JobOffer, approved bool) {
	if offer == nil || offer.CompanyID == "" {
		return
	}

	actionURL := "/dashboard/ofertas"
	dataJSON, _ := json.Marshal(map[string]interface{}{
		"offer_id":   offer.ID,
		"title":      offer.Title,
		"approved":   approved,
		"action_url": actionURL,
	})

	var title, message, eventType string
	if approved {
		title = "Oferta laboral aprobada"
		message = fmt.Sprintf("Tu oferta laboral «%s» ha sido aprobada y ya se encuentra activa.", offer.Title)
		eventType = "offer.approved"
	} else {
		title = "Oferta laboral rechazada"
		message = fmt.Sprintf("Tu oferta laboral «%s» ha sido rechazada.", offer.Title)
		eventType = "offer.rejected"
	}

	payload := notificationPayload{
		SourceService:    "labor-integration",
		EventType:        eventType,
		Category:         "labor",
		Priority:         "normal",
		SenderType:       "system",
		SenderID:         "labor-integration",
		TargetType:       "individual",
		RecipientUserIDs: []string{offer.CompanyID},
		Title:            title,
		Message:          message,
		ActionURL:        &actionURL,
		Data:             string(dataJSON),
		IsCritical:       false,
		SendEmail:        true,
		SendPush:         true,
		SendInApp:        true,
	}

	_ = d.sendNotification(ctx, payload)
}

func (d *httpNotificationDispatcher) ApplicationReceived(ctx context.Context, app *models.JobApplication, offer *models.JobOffer, actorID string) {
	if app == nil || offer == nil {
		return
	}

	actionURL := "/dashboard/ofertas"

	// 1. Notificar a la empresa/empleador (sin PII, ADR-0007)
	if offer.CompanyID != "" {
		dataCompany, _ := json.Marshal(map[string]interface{}{
			"offer_id":       offer.ID,
			"application_id": app.ID,
			"title":          offer.Title,
			"action_url":     actionURL,
		})

		payloadCompany := notificationPayload{
			SourceService:    "labor-integration",
			EventType:        "application.received",
			Category:         "labor",
			Priority:         "normal",
			SenderType:       "system",
			SenderID:         actorID,
			TargetType:       "individual",
			RecipientUserIDs: []string{offer.CompanyID},
			Title:            "Nueva postulación recibida",
			Message:          fmt.Sprintf("Un participante se postuló a «%s».", offer.Title),
			ActionURL:        &actionURL,
			Data:             string(dataCompany),
			IsCritical:       false,
			SendEmail:        true,
			SendPush:         true,
			SendInApp:        true,
		}

		_ = d.sendNotification(ctx, payloadCompany)
	}

	// 2. Notificar al participante postulante
	participantUserID := d.resolveParticipantUserID(ctx, app)
	if participantUserID != "" {
		dataParticipant, _ := json.Marshal(map[string]interface{}{
			"offer_id":       offer.ID,
			"application_id": app.ID,
			"title":          offer.Title,
			"action_url":     actionURL,
		})

		payloadParticipant := notificationPayload{
			SourceService:    "labor-integration",
			EventType:        "application.submitted",
			Category:         "labor",
			Priority:         "normal",
			SenderType:       "system",
			SenderID:         "labor-integration",
			TargetType:       "individual",
			RecipientUserIDs: []string{participantUserID},
			Title:            "Postulación registrada",
			Message:          fmt.Sprintf("Te has postulado con éxito a la oferta «%s».", offer.Title),
			ActionURL:        &actionURL,
			Data:             string(dataParticipant),
			IsCritical:       false,
			SendEmail:        true,
			SendPush:         true,
			SendInApp:        true,
		}

		_ = d.sendNotification(ctx, payloadParticipant)
	}
}

func (d *httpNotificationDispatcher) ApplicationDecided(ctx context.Context, app *models.JobApplication, offer *models.JobOffer, status string) {
	if app == nil || offer == nil {
		return
	}

	lower := strings.ToLower(strings.TrimSpace(status))
	isAccepted := lower == "aceptado" || lower == "aceptada" || lower == "accepted" || lower == "aprobada" || lower == "aprobado" || lower == "seleccionado" || lower == "contratado"
	isRejected := lower == "rechazado" || lower == "rechazada" || lower == "rejected" || lower == "descartado" || lower == "descartada"

	if !isAccepted && !isRejected {
		return
	}

	actionURL := "/dashboard/ofertas"

	// 1. Notificar al participante
	participantUserID := d.resolveParticipantUserID(ctx, app)
	if participantUserID != "" {
		var partTitle, partMsg, partEvent string
		if isAccepted {
			partTitle = "Postulación aceptada"
			partMsg = fmt.Sprintf("¡Felicitaciones! Tu postulación para «%s» ha sido aceptada.", offer.Title)
			partEvent = "application.accepted"
		} else {
			partTitle = "Actualización de postulación"
			partMsg = fmt.Sprintf("Tu postulación para la oferta «%s» no ha sido seleccionada en esta oportunidad.", offer.Title)
			partEvent = "application.rejected"
		}

		dataPart, _ := json.Marshal(map[string]interface{}{
			"offer_id":       offer.ID,
			"application_id": app.ID,
			"title":          offer.Title,
			"status":         status,
			"action_url":     actionURL,
		})

		payloadPart := notificationPayload{
			SourceService:    "labor-integration",
			EventType:        partEvent,
			Category:         "labor",
			Priority:         "normal",
			SenderType:       "system",
			SenderID:         "labor-integration",
			TargetType:       "individual",
			RecipientUserIDs: []string{participantUserID},
			Title:            partTitle,
			Message:          partMsg,
			ActionURL:        &actionURL,
			Data:             string(dataPart),
			IsCritical:       false,
			SendEmail:        true,
			SendPush:         true,
			SendInApp:        true,
		}

		_ = d.sendNotification(ctx, payloadPart)
	}

	// 2. Notificar a todos los administradores y voluntarios
	var adminVolunteerIDs []string
	if d.authClient != nil {
		ids, err := d.authClient.GetActiveUserIDsByRoles(ctx, []string{"role_admin", "role_volunteer"})
		if err != nil {
			slog.Warn("failed to fetch admin and volunteer IDs for application decision notification", slog.String("error", err.Error()))
		} else {
			adminVolunteerIDs = ids
		}
	}

	staffRecipients := filterAndDeduplicate(adminVolunteerIDs, "")
	if len(staffRecipients) > 0 {
		var staffTitle, staffMsg, staffEvent string
		if isAccepted {
			staffTitle = "Postulación aceptada en oferta laboral"
			staffMsg = fmt.Sprintf("Se ha aceptado una postulación para la oferta «%s».", offer.Title)
			staffEvent = "application.accepted.staff"
		} else {
			staffTitle = "Postulación rechazada en oferta laboral"
			staffMsg = fmt.Sprintf("Se ha rechazado una postulación para la oferta «%s».", offer.Title)
			staffEvent = "application.rejected.staff"
		}

		dataStaff, _ := json.Marshal(map[string]interface{}{
			"offer_id":       offer.ID,
			"application_id": app.ID,
			"title":          offer.Title,
			"status":         status,
			"action_url":     actionURL,
		})

		payloadStaff := notificationPayload{
			SourceService:    "labor-integration",
			EventType:        staffEvent,
			Category:         "labor",
			Priority:         "normal",
			SenderType:       "system",
			SenderID:         "labor-integration",
			TargetType:       "role",
			RecipientUserIDs: staffRecipients,
			Title:            staffTitle,
			Message:          staffMsg,
			ActionURL:        &actionURL,
			Data:             string(dataStaff),
			IsCritical:       false,
			SendEmail:        true,
			SendPush:         true,
			SendInApp:        true,
		}

		_ = d.sendNotification(ctx, payloadStaff)
	}
}

func (d *httpNotificationDispatcher) sendNotification(ctx context.Context, payload notificationPayload) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Warn("failed to serialize notification payload", slog.String("error", err.Error()), slog.String("event_type", payload.EventType))
		return err
	}

	endpoint := d.baseURL + "/api/v1/notifications/send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Warn("failed to create notification request", slog.String("error", err.Error()), slog.String("event_type", payload.EventType))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if d.secret != "" {
		mac := hmac.New(sha256.New, []byte(d.secret))
		mac.Write(bodyBytes)
		req.Header.Set("X-Internal-Signature", hex.EncodeToString(mac.Sum(nil)))
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		slog.Warn("failed to dispatch notification", slog.String("error", err.Error()), slog.String("event_type", payload.EventType))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("notifications service responded with status code: %d", resp.StatusCode)
		slog.Warn("notification dispatch received non-success response", slog.Int("status", resp.StatusCode), slog.String("event_type", payload.EventType))
		return err
	}

	slog.Info("notification dispatched successfully", slog.String("event_type", payload.EventType), slog.Int("recipients", len(payload.RecipientUserIDs)))
	return nil
}

func (d *httpNotificationDispatcher) resolveParticipantUserID(ctx context.Context, app *models.JobApplication) string {
	if app.UserID != "" && !strings.HasPrefix(app.UserID, "part-") {
		return app.UserID
	}
	targetID := app.ParticipantID
	if targetID == "" {
		targetID = app.UserID
	}
	if targetID == "" {
		return ""
	}
	if d.authClient != nil {
		resp, err := d.authClient.GetParticipant(ctx, targetID)
		if err == nil && resp != nil && resp.UserId != "" {
			return resp.UserId
		}
	}
	return targetID
}

func filterAndDeduplicate(userIDs []string, excludeID string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, id := range userIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" || trimmed == excludeID || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		result = append(result, trimmed)
	}
	return result
}

