package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"buenvivir/notifications/internal/clients"
	"buenvivir/notifications/internal/dto"
	"buenvivir/notifications/internal/models"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/templates"
)

var (
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrInvalidRecipientID = errors.New("invalid recipient id")
	ErrInvalidInput       = errors.New("invalid notification input")
)

// NotificationService defines domain operations for notification management and ingestion.
type NotificationService interface {
	GetNotifications(ctx context.Context, userID string, limit, offset int) ([]dto.NotificationResponseDTO, error)
	GetNotificationDetail(ctx context.Context, recipientID, userID string) (*dto.NotificationDetailDTO, error)
	MarkAsRead(ctx context.Context, recipientID, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	DeleteNotification(ctx context.Context, recipientID, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
	IngestNotification(ctx context.Context, req dto.SendNotificationRequestDTO) (string, int, error)
	GetPreferences(ctx context.Context, userID string) ([]dto.NotificationPreferenceDTO, error)
	GetPreferenceByCategory(ctx context.Context, userID string, category string) (*dto.NotificationPreferenceDTO, error)
	SavePreference(ctx context.Context, userID string, req dto.UpdatePreferenceRequestDTO) (*dto.NotificationPreferenceDTO, error)
	PatchPreference(ctx context.Context, userID string, category string, req dto.PatchPreferenceRequestDTO) (*dto.NotificationPreferenceDTO, error)
	InitDefaultPreferences(ctx context.Context, userID string) error
	ResetPreferences(ctx context.Context, userID string) error

	// Deprecated / Backwards compatibility
	FindByUserID(ctx context.Context, userID string) ([]*models.Notification, error)
}

type notificationService struct {
	repo         repositories.NotificationRepository
	emailService EmailService
	pushService  PushService
	tmplEngine   templates.TemplateEngine
	authClient   clients.AuthClient
	syncDispatch bool
}

// NotificationServiceOption allows configuring optional service dependencies.
type NotificationServiceOption func(*notificationService)

func WithEmailService(emailSvc EmailService) NotificationServiceOption {
	return func(s *notificationService) {
		s.emailService = emailSvc
	}
}

func WithPushService(pushSvc PushService) NotificationServiceOption {
	return func(s *notificationService) {
		s.pushService = pushSvc
	}
}

func WithTemplateEngine(tmplEngine templates.TemplateEngine) NotificationServiceOption {
	return func(s *notificationService) {
		s.tmplEngine = tmplEngine
	}
}

func WithAuthClient(authClient clients.AuthClient) NotificationServiceOption {
	return func(s *notificationService) {
		s.authClient = authClient
	}
}

func WithSyncDispatch(sync bool) NotificationServiceOption {
	return func(s *notificationService) {
		s.syncDispatch = sync
	}
}

func NewNotificationService(repo repositories.NotificationRepository, opts ...NotificationServiceOption) NotificationService {
	s := &notificationService{repo: repo}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *notificationService) FindByUserID(ctx context.Context, userID string) ([]*models.Notification, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *notificationService) GetNotifications(ctx context.Context, userID string, limit, offset int) ([]dto.NotificationResponseDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	recipients, err := s.repo.FindRecipientsByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]dto.NotificationResponseDTO, 0, len(recipients))
	for _, rec := range recipients {
		item := dto.NotificationResponseDTO{
			RecipientID:    rec.RecipientID,
			NotificationID: rec.NotificationID,
			IsRead:         rec.IsRead,
			ReadAt:         rec.ReadAt,
			CreatedAt:      rec.CreatedAt,
		}
		if rec.Notification != nil {
			item.Title = rec.Notification.Title
			item.Message = rec.Notification.Message
			item.Category = string(rec.Notification.Category)
			item.Priority = string(rec.Notification.Priority)
			item.SourceService = rec.Notification.SourceService
			item.ActionURL = rec.Notification.ActionURL
			item.IsCritical = rec.Notification.IsCritical
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *notificationService) GetNotificationDetail(ctx context.Context, recipientID, userID string) (*dto.NotificationDetailDTO, error) {
	if strings.TrimSpace(recipientID) == "" {
		return nil, ErrInvalidRecipientID
	}
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}

	rec, err := s.repo.FindRecipientByID(ctx, recipientID, userID)
	if err != nil {
		return nil, err
	}

	detail := &dto.NotificationDetailDTO{
		NotificationResponseDTO: dto.NotificationResponseDTO{
			RecipientID:    rec.RecipientID,
			NotificationID: rec.NotificationID,
			IsRead:         rec.IsRead,
			ReadAt:         rec.ReadAt,
			CreatedAt:      rec.CreatedAt,
		},
	}

	if rec.Notification != nil {
		detail.Title = rec.Notification.Title
		detail.Message = rec.Notification.Message
		detail.Category = string(rec.Notification.Category)
		detail.Priority = string(rec.Notification.Priority)
		detail.SourceService = rec.Notification.SourceService
		detail.ActionURL = rec.Notification.ActionURL
		detail.IsCritical = rec.Notification.IsCritical
		detail.SenderType = string(rec.Notification.SenderType)
		detail.SenderID = rec.Notification.SenderID

		if strings.TrimSpace(rec.Notification.Data) != "" {
			var parsedData map[string]interface{}
			if err := json.Unmarshal([]byte(rec.Notification.Data), &parsedData); err == nil {
				detail.Data = parsedData
			}
		}
	}

	return detail, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, recipientID, userID string) error {
	if strings.TrimSpace(recipientID) == "" {
		return ErrInvalidRecipientID
	}
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}
	return s.repo.MarkRecipientAsRead(ctx, recipientID, userID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}
	return s.repo.MarkAllRecipientsAsRead(ctx, userID)
}

func (s *notificationService) DeleteNotification(ctx context.Context, recipientID, userID string) error {
	if strings.TrimSpace(recipientID) == "" {
		return ErrInvalidRecipientID
	}
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}
	return s.repo.HardDeleteRecipient(ctx, recipientID, userID)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	if strings.TrimSpace(userID) == "" {
		return 0, ErrInvalidUserID
	}
	return s.repo.CountUnreadByUserID(ctx, userID)
}

func validateIngestRequest(req dto.SendNotificationRequestDTO) error {
	if strings.TrimSpace(req.SourceService) == "" || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Message) == "" {
		return ErrInvalidInput
	}
	if len(req.RecipientUserIDs) == 0 && req.TargetRole == nil && req.TargetType != string(models.TargetTypeBroadcast) {
		return errors.New("notification must target at least one recipient user, role, or broadcast")
	}
	return nil
}

func (s *notificationService) persistAndDispatchRecipients(
	ctx context.Context,
	notif *models.Notification,
	req dto.SendNotificationRequestDTO,
) (int, error) {
	prefMap, _ := s.repo.GetPreferencesByUserIDsAndCategory(ctx, req.RecipientUserIDs, notif.Category)
	if prefMap == nil {
		prefMap = make(map[string]*models.NotificationPreference)
	}

	recipients := buildNotificationRecipients(notif, req, prefMap)
	if len(recipients) == 0 {
		return 0, nil
	}

	if err := s.repo.CreateRecipients(ctx, recipients); err != nil {
		return 0, fmt.Errorf("failed to persist recipients: %w", err)
	}

	for _, rec := range recipients {
		if prefMap[rec.UserID] == nil {
			_ = s.InitDefaultPreferences(context.Background(), rec.UserID)
		}
	}

	if notif.SendEmail || notif.SendPush {
		if s.syncDispatch {
			s.dispatchNotification(notif, recipients)
		} else {
			go s.dispatchNotification(notif, recipients)
		}
	}
	return len(recipients), nil
}

func (s *notificationService) IngestNotification(ctx context.Context, req dto.SendNotificationRequestDTO) (string, int, error) {
	if err := validateIngestRequest(req); err != nil {
		return "", 0, err
	}

	s.resolveCampaignBroadcastRecipients(ctx, &req)

	notifID := generateEntityID("ntf")
	notif := buildNotificationModel(notifID, req)

	if err := s.repo.CreateNotification(ctx, notif); err != nil {
		return "", 0, fmt.Errorf("failed to persist notification: %w", err)
	}

	count, err := s.persistAndDispatchRecipients(ctx, notif, req)
	if err != nil {
		return notifID, 0, err
	}

	return notifID, count, nil
}

func (s *notificationService) resolveCampaignBroadcastRecipients(ctx context.Context, req *dto.SendNotificationRequestDTO) {
	isCampaignBroadcast := (req.TargetType == string(models.TargetTypeBroadcast) || len(req.RecipientUserIDs) == 0) &&
		models.NotificationCategory(req.Category) == models.CategoryCampaigns
	if !isCampaignBroadcast || s.authClient == nil {
		return
	}

	contacts, err := s.authClient.GetCampaignRecipients(ctx)
	if err != nil || len(contacts) == 0 {
		return
	}

	req.TargetType = string(models.TargetTypeBroadcast)
	userIDs, userEmails, userNames, userVerified := extractCampaignRecipientMaps(contacts)
	req.RecipientUserIDs = userIDs

	var currentData map[string]interface{}
	if strings.TrimSpace(req.Data) != "" {
		_ = json.Unmarshal([]byte(req.Data), &currentData)
	}
	if currentData == nil {
		currentData = make(map[string]interface{})
	}
	currentData["user_emails"] = userEmails
	currentData["user_names"] = userNames
	currentData["user_verified_emails"] = userVerified
	if mergedJSON, err := json.Marshal(currentData); err == nil {
		req.Data = string(mergedJSON)
	}
}

func extractCampaignRecipientMaps(contacts []clients.RecipientContact) ([]string, map[string]string, map[string]string, map[string]bool) {
	userIDs := make([]string, 0, len(contacts))
	userEmails := make(map[string]string, len(contacts))
	userNames := make(map[string]string, len(contacts))
	userVerified := make(map[string]bool, len(contacts))

	for _, c := range contacts {
		if c.UserID == "" {
			continue
		}
		userIDs = append(userIDs, c.UserID)
		if c.Email != "" {
			userEmails[c.UserID] = c.Email
		}
		name := strings.TrimSpace(c.FirstName + " " + c.LastName)
		if name != "" {
			userNames[c.UserID] = name
		}
		userVerified[c.UserID] = c.IsEmailVerified
	}
	return userIDs, userEmails, userNames, userVerified
}

func buildNotificationModel(notifID string, req dto.SendNotificationRequestDTO) *models.Notification {
	category := models.NotificationCategory(req.Category)
	if category == "" {
		category = models.CategorySystem
	}

	priority := models.NotificationPriority(req.Priority)
	if priority == "" {
		priority = models.PriorityNormal
	}

	senderType := models.SenderType(req.SenderType)
	if senderType == "" {
		senderType = models.SenderTypeSystem
	}
	senderID := req.SenderID
	if senderID == "" {
		senderID = "system"
	}

	targetType := models.TargetType(req.TargetType)
	if targetType == "" {
		if len(req.RecipientUserIDs) > 1 {
			targetType = models.TargetTypeBroadcast
		} else {
			targetType = models.TargetTypeIndividual
		}
	}

	sendInApp := true
	if req.SendInApp != nil {
		sendInApp = *req.SendInApp
	}

	return &models.Notification{
		ID:            notifID,
		SourceService: req.SourceService,
		EventType:     req.EventType,
		Category:      category,
		Priority:      priority,
		SenderType:    senderType,
		SenderID:      senderID,
		TargetType:    targetType,
		TargetRole:    req.TargetRole,
		Title:         req.Title,
		Message:       req.Message,
		ActionURL:     req.ActionURL,
		Data:          req.Data,
		IsCritical:    req.IsCritical,
		SendEmail:     req.SendEmail,
		SendPush:      req.SendPush,
		SendInApp:     sendInApp,
		CreatedAt:     time.Now(),
	}
}

func buildNotificationRecipients(
	notif *models.Notification,
	req dto.SendNotificationRequestDTO,
	prefMap map[string]*models.NotificationPreference,
) []models.NotificationRecipient {
	recipients := make([]models.NotificationRecipient, 0, len(req.RecipientUserIDs))
	isCritical := notif.IsCritical || notif.Category == models.CategorySecurity

	for _, uid := range req.RecipientUserIDs {
		trimmedUID := strings.TrimSpace(uid)
		if trimmedUID == "" {
			continue
		}

		pref := prefMap[trimmedUID]
		emailAllowed := pref == nil || pref.AllowsChannel("email", isCritical)
		pushAllowed := pref == nil || pref.AllowsChannel("push", isCritical)
		inAppAllowed := pref == nil || pref.AllowsChannel("in_app", isCritical)

		// If user chose "Ninguna" (both channels false and in_app disabled) and notification is not critical/security:
		// completely omit recipient from inbox and dispatching.
		if !isCritical && !emailAllowed && !pushAllowed && !inAppAllowed {
			continue
		}

		emailStatus := models.DeliveryStatusNotRequested
		if req.SendEmail {
			if emailAllowed {
				emailStatus = models.DeliveryStatusPending
			} else {
				emailStatus = models.DeliveryStatusSkipped
			}
		}

		pushStatus := models.DeliveryStatusNotRequested
		if req.SendPush {
			if pushAllowed {
				pushStatus = models.DeliveryStatusPending
			} else {
				pushStatus = models.DeliveryStatusSkipped
			}
		}

		rec := models.NotificationRecipient{
			RecipientID:    generateEntityID("rcp"),
			NotificationID: notif.ID,
			UserID:         trimmedUID,
			RoleID:         req.TargetRole,
			IsRead:         false,
			EmailStatus:    emailStatus,
			PushStatus:     pushStatus,
			CreatedAt:      time.Now(),
		}
		recipients = append(recipients, rec)
	}
	return recipients
}

func generateEntityID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}

var standardCategories = []models.NotificationCategory{
	models.CategorySecurity,
	models.CategoryEducation,
	models.CategoryLabor,
	models.CategoryCampaigns,
	models.CategoryHumanPromotion,
	models.CategorySystem,
}

func (s *notificationService) InitDefaultPreferences(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}

	saved, err := s.repo.GetPreferencesByUserID(ctx, userID)
	if err != nil {
		return err
	}

	prefMap := make(map[models.NotificationCategory]bool, len(saved))
	for _, p := range saved {
		prefMap[p.Category] = true
	}

	for _, cat := range standardCategories {
		if !prefMap[cat] {
			newPref := &models.NotificationPreference{
				ID:           generateEntityID("prf"),
				UserID:       userID,
				Category:     cat,
				EmailEnabled: true,
				PushEnabled:  true,
				InAppEnabled: true,
				UpdatedAt:    time.Now(),
			}
			if err := s.repo.UpsertPreference(ctx, newPref); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *notificationService) ResetPreferences(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}

	for _, cat := range standardCategories {
		pref := &models.NotificationPreference{
			ID:           generateEntityID("prf"),
			UserID:       userID,
			Category:     cat,
			EmailEnabled: true,
			PushEnabled:  true,
			InAppEnabled: true,
			UpdatedAt:    time.Now(),
		}
		if err := s.repo.UpsertPreference(ctx, pref); err != nil {
			return err
		}
	}
	return nil
}

func (s *notificationService) GetPreferences(ctx context.Context, userID string) ([]dto.NotificationPreferenceDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}

	if err := s.InitDefaultPreferences(ctx, userID); err != nil {
		return nil, err
	}

	saved, err := s.repo.GetPreferencesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.NotificationPreferenceDTO, 0, len(saved))
	for _, p := range saved {
		result = append(result, dto.NotificationPreferenceDTO{
			Category:     string(p.Category),
			EmailEnabled: p.EmailEnabled,
			PushEnabled:  p.PushEnabled,
			InAppEnabled: p.InAppEnabled,
		})
	}

	return result, nil
}

func (s *notificationService) GetPreferenceByCategory(ctx context.Context, userID string, category string) (*dto.NotificationPreferenceDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}
	cat, err := parseCategory(category)
	if err != nil {
		return nil, err
	}

	p, err := s.repo.GetPreference(ctx, userID, cat)
	if err != nil {
		return nil, err
	}

	if p == nil {
		newPref := &models.NotificationPreference{
			ID:           generateEntityID("prf"),
			UserID:       userID,
			Category:     cat,
			EmailEnabled: true,
			PushEnabled:  true,
			InAppEnabled: true,
			UpdatedAt:    time.Now(),
		}
		if err := s.repo.UpsertPreference(ctx, newPref); err != nil {
			return nil, err
		}
		p = newPref
	}

	return &dto.NotificationPreferenceDTO{
		Category:     string(p.Category),
		EmailEnabled: p.EmailEnabled,
		PushEnabled:  p.PushEnabled,
		InAppEnabled: p.InAppEnabled,
	}, nil
}

func (s *notificationService) SavePreference(ctx context.Context, userID string, req dto.UpdatePreferenceRequestDTO) (*dto.NotificationPreferenceDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}
	cat, err := parseCategory(req.Category)
	if err != nil {
		return nil, err
	}

	emailEnabled := true
	if req.EmailEnabled != nil {
		emailEnabled = *req.EmailEnabled
	}
	pushEnabled := true
	if req.PushEnabled != nil {
		pushEnabled = *req.PushEnabled
	}
	inAppEnabled := true
	if req.InAppEnabled != nil {
		inAppEnabled = *req.InAppEnabled
	}

	pref := &models.NotificationPreference{
		ID:           generateEntityID("prf"),
		UserID:       userID,
		Category:     cat,
		EmailEnabled: emailEnabled,
		PushEnabled:  pushEnabled,
		InAppEnabled: inAppEnabled,
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.UpsertPreference(ctx, pref); err != nil {
		return nil, err
	}

	return &dto.NotificationPreferenceDTO{
		Category:     string(cat),
		EmailEnabled: emailEnabled,
		PushEnabled:  pushEnabled,
		InAppEnabled: inAppEnabled,
	}, nil
}

func (s *notificationService) PatchPreference(ctx context.Context, userID string, category string, req dto.PatchPreferenceRequestDTO) (*dto.NotificationPreferenceDTO, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}
	cat, err := parseCategory(category)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetPreference(ctx, userID, cat)
	if err != nil {
		return nil, err
	}

	emailEnabled := true
	pushEnabled := true
	inAppEnabled := true
	if existing != nil {
		emailEnabled = existing.EmailEnabled
		pushEnabled = existing.PushEnabled
		inAppEnabled = existing.InAppEnabled
	}

	if req.EmailEnabled != nil {
		emailEnabled = *req.EmailEnabled
	}
	if req.PushEnabled != nil {
		pushEnabled = *req.PushEnabled
	}
	if req.InAppEnabled != nil {
		inAppEnabled = *req.InAppEnabled
	}

	pref := &models.NotificationPreference{
		ID:           generateEntityID("prf"),
		UserID:       userID,
		Category:     cat,
		EmailEnabled: emailEnabled,
		PushEnabled:  pushEnabled,
		InAppEnabled: inAppEnabled,
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.UpsertPreference(ctx, pref); err != nil {
		return nil, err
	}

	return &dto.NotificationPreferenceDTO{
		Category:     string(cat),
		EmailEnabled: emailEnabled,
		PushEnabled:  pushEnabled,
		InAppEnabled: inAppEnabled,
	}, nil
}

func parseCategory(category string) (models.NotificationCategory, error) {
	cat := models.NotificationCategory(strings.TrimSpace(category))
	switch cat {
	case models.CategorySecurity, models.CategoryEducation, models.CategoryLabor, models.CategoryCampaigns, models.CategoryHumanPromotion, models.CategorySystem:
		return cat, nil
	default:
		return "", fmt.Errorf("invalid notification category: %s", category)
	}
}

func (s *notificationService) dispatchNotification(notif *models.Notification, recipients []models.NotificationRecipient) {
	if notif == nil || len(recipients) == 0 {
		return
	}

	var dataMap map[string]interface{}
	if notif.Data != "" {
		_ = json.Unmarshal([]byte(notif.Data), &dataMap)
	}
	if dataMap == nil {
		dataMap = make(map[string]interface{})
	}

	for _, rec := range recipients {
		s.dispatchForRecipient(notif, rec, dataMap)
	}
}

func (s *notificationService) dispatchForRecipient(
	notif *models.Notification,
	rec models.NotificationRecipient,
	dataMap map[string]interface{},
) {
	ctx := context.Background()
	pref, _ := s.repo.GetPreference(ctx, rec.UserID, notif.Category)

	emailStatus := rec.EmailStatus
	var emailMsgID *string
	var emailErr error
	if rec.EmailStatus == models.DeliveryStatusPending {
		emailStatus, emailMsgID, emailErr = s.dispatchEmailChannel(ctx, notif, rec.UserID, pref, dataMap)
		if emailStatus == models.DeliveryStatusNotRequested {
			emailStatus = rec.EmailStatus
		}
	}

	pushStatus := rec.PushStatus
	var pushMsgID *string
	var pushErr error
	if rec.PushStatus == models.DeliveryStatusPending {
		pushStatus, pushMsgID, pushErr = s.dispatchPushChannel(ctx, notif, rec.UserID, pref, dataMap)
		if pushStatus == models.DeliveryStatusNotRequested {
			pushStatus = rec.PushStatus
		}
	}

	var errorMsg *string
	var errStrs []string
	if emailErr != nil {
		errStrs = append(errStrs, emailErr.Error())
	}
	if pushErr != nil {
		errStrs = append(errStrs, pushErr.Error())
	}
	if len(errStrs) > 0 {
		combined := strings.Join(errStrs, "; ")
		errorMsg = &combined
	}

	_ = s.repo.UpdateDeliveryStatus(ctx, rec.RecipientID, emailStatus, pushStatus, emailMsgID, pushMsgID, errorMsg)
}

func isChannelPermitted(prefEnabled, isCritical bool, category models.NotificationCategory) bool {
	return prefEnabled || isCritical || category == models.CategorySecurity
}

func (s *notificationService) dispatchEmailChannel(
	ctx context.Context,
	notif *models.Notification,
	userID string,
	pref *models.NotificationPreference,
	dataMap map[string]interface{},
) (models.DeliveryStatus, *string, error) {
	if !notif.SendEmail || s.emailService == nil {
		return models.DeliveryStatusNotRequested, nil, nil
	}

	emailAllowed := pref == nil || isChannelPermitted(pref.EmailEnabled, notif.IsCritical, notif.Category)
	if !emailAllowed {
		return models.DeliveryStatusSkipped, nil, nil
	}

	toEmail := resolveEmail(userID, dataMap)
	isVerifiedKnown := false
	isVerifiedVal := false
	if toEmail == "" && s.authClient != nil && userID != "" && !strings.Contains(userID, "@") {
		contact, err := s.authClient.GetUserContact(ctx, userID)
		if err == nil && contact != nil && contact.Email != "" {
			toEmail = contact.Email
			isVerifiedKnown = true
			isVerifiedVal = contact.IsEmailVerified
			if contact.FirstName != "" || contact.LastName != "" {
				fullName := strings.TrimSpace(contact.FirstName + " " + contact.LastName)
				if dataMap == nil {
					dataMap = make(map[string]interface{})
				}
				if _, exists := dataMap["recipient_name"]; !exists && fullName != "" {
					dataMap["recipient_name"] = fullName
				}
			}
		}
	}
	if toEmail == "" {
		return models.DeliveryStatusSkipped, nil, nil
	}

	// Email Verification Validation: Recipient must have verified their email before receiving emails.
	// Exception: The verification email itself is allowed through to deliver the verification token.
	if !isEmailVerificationMessage(notif, dataMap) {
		verified := false
		if isVerifiedKnown {
			verified = isVerifiedVal
		} else {
			verified = s.isUserEmailVerified(ctx, userID, dataMap)
		}
		if !verified {
			return models.DeliveryStatusSkipped, nil, nil
		}
	}

	senderType := EmailSenderNotification
	if notif.IsCritical || notif.Category == models.CategorySecurity {
		senderType = EmailSenderAlert
	}

	htmlBody := s.renderNotificationEmail(notif, userID, toEmail, dataMap)
	id, err := s.emailService.SendEmail(ctx, EmailMessage{
		To:         []string{toEmail},
		Subject:    notif.Title,
		Html:       htmlBody,
		SenderType: senderType,
		Category:   string(notif.Category),
	})
	if err != nil {
		return models.DeliveryStatusFailed, nil, fmt.Errorf("email error: %w", err)
	}
	return models.DeliveryStatusSent, &id, nil
}

func (s *notificationService) dispatchPushChannel(
	ctx context.Context,
	notif *models.Notification,
	userID string,
	pref *models.NotificationPreference,
	dataMap map[string]interface{},
) (models.DeliveryStatus, *string, error) {
	if !notif.SendPush || s.pushService == nil || s.authClient == nil {
		return models.DeliveryStatusNotRequested, nil, nil
	}

	pushAllowed := pref == nil || isChannelPermitted(pref.PushEnabled, notif.IsCritical, notif.Category)
	if !pushAllowed {
		return models.DeliveryStatusSkipped, nil, nil
	}

	tokens, err := s.authClient.GetUserDeviceTokens(ctx, userID)
	if err != nil {
		return models.DeliveryStatusFailed, nil, fmt.Errorf("tokens error: %w", err)
	}
	if len(tokens) == 0 {
		return models.DeliveryStatusSkipped, nil, nil
	}

	pushData := extractPushData(notif, dataMap)
	res, pushErr := s.pushService.SendMulticast(ctx, tokens, notif.Title, notif.Message, pushData)
	if pushErr != nil {
		return models.DeliveryStatusFailed, nil, fmt.Errorf("push error: %w", pushErr)
	}

	if len(res.InvalidTokens) > 0 {
		_ = s.authClient.RevokeDeviceTokens(ctx, userID, res.InvalidTokens)
	}

	if res.SuccessCount > 0 {
		msgID := fmt.Sprintf("fcm-%d-sent", res.SuccessCount)
		return models.DeliveryStatusSent, &msgID, nil
	}
	return models.DeliveryStatusFailed, nil, nil
}

func extractPushData(notif *models.Notification, dataMap map[string]interface{}) map[string]string {
	pushData := make(map[string]string)
	for k, v := range dataMap {
		if strVal, ok := v.(string); ok {
			pushData[k] = strVal
		}
	}
	pushData["notification_id"] = notif.ID
	pushData["category"] = string(notif.Category)
	return pushData
}

func extractExpiry(data map[string]interface{}, defaultVal int) int {
	if e, ok := data["expiry_minutes"].(float64); ok && e > 0 {
		return int(e)
	}
	return defaultVal
}

func extractString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func (s *notificationService) renderAuthTemplate(
	templateID templates.EmailTemplateID,
	recipientName, actionURL string,
	dataMap map[string]interface{},
	year int,
) (string, error) {
	switch templateID {
	case templates.TemplateEmailVerification:
		return s.tmplEngine.RenderEmail(templateID, templates.EmailVerificationData{
			RecipientName: recipientName,
			ActionURL:     actionURL,
			ExpiryMinutes: extractExpiry(dataMap, 60),
			Year:          year,
		})
	case templates.TemplateAccountApproved:
		return s.tmplEngine.RenderEmail(templateID, templates.AccountApprovedData{
			RecipientName: recipientName,
			ActionURL:     actionURL,
			Year:          year,
		})
	case templates.TemplateTwoFactorCode:
		return s.tmplEngine.RenderEmail(templateID, templates.TwoFactorCodeData{
			RecipientName: recipientName,
			Code:          extractString(dataMap, "code"),
			ExpiryMinutes: extractExpiry(dataMap, 10),
			Year:          year,
		})
	case templates.TemplatePasswordRecovery:
		return s.tmplEngine.RenderEmail(templateID, templates.PasswordRecoveryData{
			RecipientName: recipientName,
			ActionURL:     actionURL,
			ExpiryMinutes: extractExpiry(dataMap, 30),
			Year:          year,
		})
	case templates.TemplateEmailChange:
		return s.tmplEngine.RenderEmail(templateID, templates.EmailChangeData{
			RecipientName: recipientName,
			NewEmail:      extractString(dataMap, "new_email"),
			ActionURL:     actionURL,
			Year:          year,
		})
	default:
		return "", errors.New("not an auth template")
	}
}

func (s *notificationService) renderNotificationEmail(
	notif *models.Notification,
	userID, recipientEmail string,
	dataMap map[string]interface{},
) string {
	if s.tmplEngine == nil {
		return fmt.Sprintf("<h2>%s</h2><p>%s</p>", html.EscapeString(notif.Title), html.EscapeString(notif.Message))
	}

	actionURL := resolveNotificationActionURL(notif, dataMap)
	recipientName := resolveNotificationRecipientName(userID, recipientEmail, dataMap)

	templateID := templates.EmailTemplateID(extractString(dataMap, "template_id"))
	year := templates.CurrentYear()

	renderedHTML, err := s.renderAuthTemplate(templateID, recipientName, actionURL, dataMap, year)
	if err == nil {
		return renderedHTML
	}

	alertHTML, alertErr := s.tmplEngine.RenderEmail(templates.TemplateGeneralAlert, templates.GeneralAlertData{
		RecipientName: recipientName,
		Title:         notif.Title,
		Message:       notif.Message,
		Details:       extractString(dataMap, "details"),
		ActionURL:     actionURL,
		ActionText:    extractString(dataMap, "action_text"),
		Category:      string(notif.Category),
		IsCritical:    notif.IsCritical,
		Year:          year,
	})
	if alertErr != nil {
		return fmt.Sprintf("<h2>%s</h2><p>%s</p>", html.EscapeString(notif.Title), html.EscapeString(notif.Message))
	}
	return alertHTML
}

func resolveNotificationActionURL(notif *models.Notification, dataMap map[string]interface{}) string {
	actionURL := extractString(dataMap, "action_url")
	if actionURL == "" && notif.ActionURL != nil {
		return *notif.ActionURL
	}
	return actionURL
}

func resolveNotificationRecipientName(userID, recipientEmail string, dataMap map[string]interface{}) string {
	if recipientName := extractString(dataMap, "recipient_name"); recipientName != "" {
		return recipientName
	}
	if namesMap, ok := dataMap["user_names"].(map[string]interface{}); ok {
		if nm, ok := namesMap[userID].(string); ok && strings.TrimSpace(nm) != "" {
			return nm
		}
	} else if namesMap, ok := dataMap["user_names"].(map[string]string); ok {
		if nm, ok := namesMap[userID]; ok && strings.TrimSpace(nm) != "" {
			return nm
		}
	}
	return recipientEmail
}

func resolveEmail(userID string, data map[string]interface{}) string {
	if strings.Contains(userID, "@") {
		return userID
	}
	if email, ok := data["email"].(string); ok && strings.Contains(email, "@") {
		return email
	}
	if email, ok := data["recipient_email"].(string); ok && strings.Contains(email, "@") {
		return email
	}
	if emailsMap, ok := data["user_emails"].(map[string]interface{}); ok {
		if em, ok := emailsMap[userID].(string); ok && strings.Contains(em, "@") {
			return em
		}
	}
	if emailsMap, ok := data["user_emails"].(map[string]string); ok {
		if em, ok := emailsMap[userID]; ok && strings.Contains(em, "@") {
			return em
		}
	}
	return ""
}

func isEmailVerificationMessage(notif *models.Notification, dataMap map[string]interface{}) bool {
	if notif != nil && notif.EventType == "auth.email_verification" {
		return true
	}
	if templateID, ok := dataMap["template_id"].(string); ok && templateID == string(templates.TemplateEmailVerification) {
		return true
	}
	return false
}

func (s *notificationService) isUserEmailVerified(ctx context.Context, userID string, dataMap map[string]interface{}) bool {
	if verifiedMap, ok := dataMap["user_verified_emails"].(map[string]bool); ok {
		if verified, found := verifiedMap[userID]; found {
			return verified
		}
	}
	if verifiedMap, ok := dataMap["user_verified_emails"].(map[string]interface{}); ok {
		if v, found := verifiedMap[userID]; found {
			if verified, ok := v.(bool); ok {
				return verified
			}
		}
	}

	if s.authClient != nil && userID != "" {
		verified, err := s.authClient.IsUserEmailVerified(ctx, userID)
		if err == nil {
			return verified
		}
		return false
	}
	return false
}
