package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"buenvivir/notifications/internal/clients"
	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/models"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/services"
	"buenvivir/notifications/internal/templates"
)

const (
	defaultMaxAttempts          = 5
	defaultBatchSize            = 50
	defaultInlineFlushTimeout   = 3 * time.Second
	defaultShutdownFlushTimeout = 5 * time.Second
)

// InlineFlushEnabled returns true on Cloud Run (K_SERVICE set) or when
// DELIVERY_INLINE_FLUSH=true, dispatching email/push inside the active Pub/Sub push request
// so Cloud Run can scale to zero without per-minute Cloud Scheduler polling.
func InlineFlushEnabled() bool {
	if raw := strings.TrimSpace(os.Getenv("DELIVERY_INLINE_FLUSH")); raw != "" {
		return strings.EqualFold(raw, "true") || raw == "1"
	}
	return strings.TrimSpace(os.Getenv("K_SERVICE")) != ""
}

// Worker runs background processing of pending notifications with exponential backoff and DLQ marking.
type Worker struct {
	repo         repositories.NotificationRepository
	emailService services.EmailService
	pushService  services.PushService
	authClient   clients.AuthClient
	tmplEngine   templates.TemplateEngine
	logger       *slog.Logger

	notify chan struct{}
	mu     sync.Mutex

	BatchSize   int
	MaxAttempts int
	Interval    time.Duration
	FrontendURL string
	Now         func() time.Time
	// Disabled apaga el envío (Flush no hace nada). Lo usa el stack Docker
	// local con DELIVERY_WORKER_ENABLED=false: comparte Firestore con
	// testing y, si procesaba los pendientes de testing, mandaba los mails
	// con enlaces a http://localhost:4200 (22/11/2026).
	Disabled bool
}

// WorkerEnabledFromEnv lee DELIVERY_WORKER_ENABLED (por defecto true).
func WorkerEnabledFromEnv() bool {
	raw := strings.TrimSpace(os.Getenv("DELIVERY_WORKER_ENABLED"))
	if raw == "" {
		return true
	}
	return !strings.EqualFold(raw, "false") && raw != "0"
}

// NewWorker instantiates a DeliveryWorker.
func NewWorker(
	repo repositories.NotificationRepository,
	emailSvc services.EmailService,
	pushSvc services.PushService,
	authClient clients.AuthClient,
	tmplEngine templates.TemplateEngine,
	logger *slog.Logger,
) *Worker {
	if logger == nil {
		logger = slog.Default()
	}
	return &Worker{
		repo:         repo,
		emailService: emailSvc,
		pushService:  pushSvc,
		authClient:   authClient,
		tmplEngine:   tmplEngine,
		logger:       logger,
		notify:       make(chan struct{}, 1),
		BatchSize:    defaultBatchSize,
		MaxAttempts:  defaultMaxAttempts,
		Interval:     30 * time.Second,
		Now:          time.Now,
	}
}

// Notify triggers an immediate async flush without blocking.
func (w *Worker) Notify() {
	if w == nil || w.notify == nil {
		return
	}
	select {
	case w.notify <- struct{}{}:
	default:
	}
}

// NotifyInline dispatches pending notifications inline (best-effort) on Cloud Run
// (or DELIVERY_INLINE_FLUSH=true), falling back to async Notify otherwise.
func (w *Worker) NotifyInline(ctx context.Context) {
	if w == nil {
		return
	}
	if InlineFlushEnabled() {
		w.FlushBestEffort(ctx, defaultInlineFlushTimeout)
		return
	}
	w.Notify()
}

// FlushBestEffort delivers pending notifications within timeout without propagating errors
// to the caller. Pending rows remain in the database for retry on the next request or shutdown.
func (w *Worker) FlushBestEffort(ctx context.Context, timeout time.Duration) int {
	if w == nil || w.repo == nil {
		return 0
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		timeout = defaultInlineFlushTimeout
	}
	flushCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	n, err := w.Flush(flushCtx)
	if err != nil {
		logger := w.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Warn("delivery worker: best-effort flush failed", slog.String("error", err.Error()))
	}
	return n
}

// Shutdown drains pending notifications during graceful shutdown (Cloud Run SIGTERM).
func (w *Worker) Shutdown(ctx context.Context) {
	w.FlushBestEffort(ctx, defaultShutdownFlushTimeout)
}

// Run executes the worker loop until context is canceled.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	for {
		if _, err := w.Flush(ctx); err != nil && ctx.Err() == nil {
			w.logger.Error("delivery worker: flush failed", slog.String("error", err.Error()))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-w.notify:
		}
	}
}

// Flush fetches and delivers due pending notifications.
func (w *Worker) Flush(ctx context.Context) (int, error) {
	if w.Disabled {
		return 0, nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	totalProcessed := 0
	for {
		pending, err := w.repo.FindPendingRecipients(ctx, w.BatchSize)
		if err != nil {
			return totalProcessed, fmt.Errorf("error finding pending recipients: %w", err)
		}
		if len(pending) == 0 {
			break
		}

		for _, rec := range pending {
			if err := w.processRecipient(ctx, rec); err != nil {
				w.logger.Warn("delivery worker: recipient processing encountered error",
					slog.String("recipient_id", rec.RecipientID),
					slog.String("user_id", rec.UserID),
					slog.String("error", err.Error()),
				)
			}
			totalProcessed++
		}

		if len(pending) < w.BatchSize {
			break
		}
	}

	return totalProcessed, nil
}

func (w *Worker) processRecipient(ctx context.Context, rec models.NotificationRecipient) error {
	notif := rec.Notification
	if notif == nil {
		// If notification relation is missing, mark dead
		return w.repo.UpdateRecipientDead(ctx, rec.RecipientID, "associated notification not found")
	}

	pref, _ := w.repo.GetPreference(ctx, rec.UserID, notif.Category)
	var dataMap map[string]interface{}
	if strings.TrimSpace(notif.Data) != "" {
		_ = json.Unmarshal([]byte(notif.Data), &dataMap)
	}
	if dataMap == nil {
		dataMap = make(map[string]interface{})
	}

	var emailErr, pushErr error
	emailStatus := rec.EmailStatus
	var emailMsgID *string
	if rec.EmailStatus == models.DeliveryStatusPending {
		emailStatus, emailMsgID, emailErr = w.deliverEmail(ctx, notif, rec.UserID, pref, dataMap)
	}

	pushStatus := rec.PushStatus
	var pushMsgID *string
	if rec.PushStatus == models.DeliveryStatusPending {
		pushStatus, pushMsgID, pushErr = w.deliverPush(ctx, notif, rec.UserID, pref, dataMap)
	}

	// Consolidate errors
	var errMsgs []string
	if emailErr != nil {
		errMsgs = append(errMsgs, emailErr.Error())
	}
	if pushErr != nil {
		errMsgs = append(errMsgs, pushErr.Error())
	}

	if len(errMsgs) > 0 {
		combinedErr := strings.Join(errMsgs, "; ")
		nextAttempt := rec.Attempts + 1
		if nextAttempt >= w.MaxAttempts {
			w.logger.Error("delivery worker: recipient exceeded max attempts, marking dead",
				slog.String("recipient_id", rec.RecipientID),
				slog.String("user_id", rec.UserID),
				slog.Int("attempts", nextAttempt),
				slog.String("last_error", combinedErr),
			)
			return w.repo.UpdateRecipientDead(ctx, rec.RecipientID, combinedErr)
		}

		backoffDuration := w.calculateBackoff(nextAttempt)
		nextAttemptAt := w.Now().UTC().Add(backoffDuration)
		return w.repo.UpdateRecipientRetry(ctx, rec.RecipientID, nextAttempt, nextAttemptAt, combinedErr)
	}

	var errorMsg *string
	return w.repo.UpdateDeliveryStatus(ctx, rec.RecipientID, emailStatus, pushStatus, emailMsgID, pushMsgID, errorMsg)
}

func (w *Worker) calculateBackoff(attempt int) time.Duration {
	d := 5 * time.Second
	for i := 1; i < attempt && d < 10*time.Minute; i++ {
		d *= 3
	}
	if d > 10*time.Minute {
		d = 10 * time.Minute
	}
	return d
}

func (w *Worker) deliverEmail(
	ctx context.Context,
	notif *models.Notification,
	userID string,
	pref *models.NotificationPreference,
	dataMap map[string]interface{},
) (models.DeliveryStatus, *string, error) {
	if !notif.SendEmail || w.emailService == nil {
		return models.DeliveryStatusNotRequested, nil, nil
	}

	isCritical := notif.IsCritical || notif.Category == models.CategorySecurity
	if pref != nil && !pref.AllowsChannel("email", isCritical) {
		return models.DeliveryStatusSkipped, nil, nil
	}

	toEmail, ok, err := w.resolveRecipientEmail(ctx, userID, isCritical, dataMap)
	if err != nil {
		return models.DeliveryStatusFailed, nil, err
	}
	if !ok || toEmail == "" {
		return models.DeliveryStatusSkipped, nil, nil
	}

	if err := w.resolveActionLink(ctx, notif, userID, dataMap); err != nil {
		return models.DeliveryStatusFailed, nil, fmt.Errorf("failed to issue action link: %w", err)
	}

	senderType := services.EmailSenderNotification
	if isCritical {
		senderType = services.EmailSenderAlert
	}

	htmlBody := w.renderEmailBody(notif, dataMap)
	msgID, err := w.emailService.SendEmail(ctx, services.EmailMessage{
		To:         []string{toEmail},
		Subject:    notif.Title,
		Html:       htmlBody,
		SenderType: senderType,
		Category:   string(notif.Category),
	})
	if err != nil {
		return models.DeliveryStatusFailed, nil, fmt.Errorf("email delivery failed: %w", err)
	}

	return models.DeliveryStatusSent, &msgID, nil
}

func (w *Worker) resolveRecipientEmail(ctx context.Context, userID string, isCritical bool, dataMap map[string]interface{}) (string, bool, error) {
	if strings.Contains(userID, "@") {
		return userID, true, nil
	}
	if w.authClient == nil {
		return "", false, nil
	}
	contact, err := w.authClient.GetUserContact(ctx, userID)
	if err != nil {
		w.logger.Warn("delivery worker: failed to resolve recipient contact",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return "", false, fmt.Errorf("failed to resolve recipient contact: %w", err)
	}
	if contact == nil || contact.Email == "" {
		return "", false, nil
	}
	if !contact.IsEmailVerified && !isCritical {
		return "", false, nil
	}
	if _, hasName := dataMap["recipient_name"]; !hasName {
		dataMap["recipient_name"] = strings.TrimSpace(contact.FirstName + " " + contact.LastName)
	}
	return contact.Email, true, nil
}

func (w *Worker) resolveActionLink(ctx context.Context, notif *models.Notification, userID string, dataMap map[string]interface{}) error {
	reqID, ok := dataMap["request_id"].(string)
	if !ok || reqID == "" || w.authClient == nil {
		return nil
	}
	actionType, _ := dataMap["tipo"].(string)
	link, err := w.authClient.IssueActionLink(ctx, reqID, userID, actionType)
	if err != nil {
		return err
	}
	if actionType == "two_factor_code" {
		dataMap["code"] = link
	} else {
		notif.ActionURL = &link
		dataMap["action_url"] = link
	}
	return nil
}

func (w *Worker) deliverPush(
	ctx context.Context,
	notif *models.Notification,
	userID string,
	pref *models.NotificationPreference,
	dataMap map[string]interface{},
) (models.DeliveryStatus, *string, error) {
	if !notif.SendPush || w.pushService == nil || w.authClient == nil {
		return models.DeliveryStatusNotRequested, nil, nil
	}

	isCritical := notif.IsCritical || notif.Category == models.CategorySecurity
	if pref != nil && !pref.AllowsChannel("push", isCritical) {
		return models.DeliveryStatusSkipped, nil, nil
	}

	tokens, err := w.authClient.GetUserDeviceTokens(ctx, userID)
	if err != nil {
		return models.DeliveryStatusFailed, nil, fmt.Errorf("failed to retrieve device tokens: %w", err)
	}
	if len(tokens) == 0 {
		return models.DeliveryStatusSkipped, nil, nil
	}

	pushData := make(map[string]string)
	for k, v := range dataMap {
		if s, ok := v.(string); ok {
			pushData[k] = s
		}
	}
	pushData["notification_id"] = notif.ID
	pushData["category"] = string(notif.Category)

	res, pushErr := w.pushService.SendMulticast(ctx, tokens, notif.Title, notif.Message, pushData)
	if pushErr != nil {
		return models.DeliveryStatusFailed, nil, fmt.Errorf("push delivery failed: %w", pushErr)
	}

	if len(res.InvalidTokens) > 0 {
		_ = w.authClient.RevokeDeviceTokens(ctx, userID, res.InvalidTokens)
	}

	if res.SuccessCount > 0 {
		id := fmt.Sprintf("fcm-%d-sent", res.SuccessCount)
		return models.DeliveryStatusSent, &id, nil
	}

	return models.DeliveryStatusFailed, nil, errors.New("fcm dispatch failed for all tokens")
}

func (w *Worker) renderEmailBody(notif *models.Notification, dataMap map[string]interface{}) string {
	if w.tmplEngine == nil {
		return fmt.Sprintf("<h2>%s</h2><p>%s</p>", html.EscapeString(notif.Title), html.EscapeString(notif.Message))
	}

	recipientName := ""
	if rn, ok := dataMap["recipient_name"].(string); ok {
		recipientName = rn
	}
	actionURL := ""
	if notif.ActionURL != nil {
		actionURL = *notif.ActionURL
	} else if au, ok := dataMap["action_url"].(string); ok {
		actionURL = au
	}
	if strings.HasPrefix(actionURL, "/") && strings.TrimSpace(w.FrontendURL) != "" {
		actionURL = strings.TrimRight(strings.TrimSpace(w.FrontendURL), "/") + actionURL
	}

	templateID := templates.TemplateGeneralAlert
	if tid, ok := dataMap["template_id"].(string); ok && tid != "" {
		templateID = templates.EmailTemplateID(tid)
	}

	var alertHTML string
	var err error

	switch templateID {
	case templates.TemplateEmailVerification:
		alertHTML, err = w.tmplEngine.RenderEmail(templateID, templates.EmailVerificationData{
			RecipientName: recipientName,
			ActionURL:     actionURL,
			ExpiryMinutes: 24,
			Year:          templates.CurrentYear(),
		})
	case templates.TemplateAccountApproved:
		alertHTML, err = w.tmplEngine.RenderEmail(templateID, templates.AccountApprovedData{
			RecipientName: recipientName,
			ActionURL:     actionURL,
			Year:          templates.CurrentYear(),
		})
	case templates.TemplatePasswordRecovery:
		alertHTML, err = w.tmplEngine.RenderEmail(templateID, templates.PasswordRecoveryData{
			RecipientName: recipientName,
			ActionURL:     actionURL,
			ExpiryMinutes: 30,
			Year:          templates.CurrentYear(),
		})
	case templates.TemplateTwoFactorCode:
		code := ""
		if c, ok := dataMap["code"].(string); ok {
			code = c
		}
		alertHTML, err = w.tmplEngine.RenderEmail(templateID, templates.TwoFactorCodeData{
			RecipientName: recipientName,
			Code:          code,
			ExpiryMinutes: 10,
			Year:          templates.CurrentYear(),
		})
	default:
		alertHTML, err = w.tmplEngine.RenderEmail(templates.TemplateGeneralAlert, templates.GeneralAlertData{
			RecipientName: recipientName,
			Title:         notif.Title,
			Message:       notif.Message,
			ActionURL:     actionURL,
			Category:      string(notif.Category),
			IsCritical:    notif.IsCritical,
			Year:          templates.CurrentYear(),
		})
	}
	if err != nil {
		return fmt.Sprintf("<h2>%s</h2><p>%s</p>", html.EscapeString(notif.Title), html.EscapeString(notif.Message))
	}
	return alertHTML
}

// FlushHandler creates an HTTP handler for scheduled flush triggers (protected via events.Verifier).
func (w *Worker) FlushHandler(verifier events.Verifier) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if verifier != nil {
			if err := verifier.Verify(req); err != nil {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		n, err := w.Flush(req.Context())
		if err != nil {
			w.logger.Error("flush handler error", slog.String("error", err.Error()))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(rw, `{"flushed":%d}`, n)
	})
}
