package repositories

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"buenvivir/notifications/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const queryRecipientAndUser = "recipient_id = ? AND user_id = ?"

// NotificationRepository defines persistence operations for notifications, recipients, and user preferences.
type NotificationRepository interface {
	FindByUserID(ctx context.Context, userID string) ([]*models.Notification, error)
	CreateNotification(ctx context.Context, notification *models.Notification) error
	CreateRecipients(ctx context.Context, recipients []models.NotificationRecipient) error
	FindRecipientsByUserID(ctx context.Context, userID string, limit, offset int) ([]models.NotificationRecipient, error)
	FindRecipientByID(ctx context.Context, recipientID, userID string) (*models.NotificationRecipient, error)
	CountUnreadByUserID(ctx context.Context, userID string) (int64, error)
	MarkRecipientAsRead(ctx context.Context, recipientID, userID string) error
	MarkAllRecipientsAsRead(ctx context.Context, userID string) error
	HardDeleteRecipient(ctx context.Context, recipientID, userID string) error
	FindPendingRecipients(ctx context.Context, limit int) ([]models.NotificationRecipient, error)
	UpdateDeliveryStatus(ctx context.Context, recipientID string, emailStatus, pushStatus models.DeliveryStatus, emailMsgID, pushMsgID, errorMsg *string) error
	GetPreference(ctx context.Context, userID string, category models.NotificationCategory) (*models.NotificationPreference, error)
	GetPreferencesByUserID(ctx context.Context, userID string) ([]models.NotificationPreference, error)
	UpsertPreference(ctx context.Context, pref *models.NotificationPreference) error
}

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository returns a GORM-backed implementation if db is provided, otherwise falls back to in-memory.
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	if db == nil {
		return NewMemoryNotificationRepository()
	}
	return &notificationRepository{db: db}
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID string) ([]*models.Notification, error) {
	recipients, err := r.FindRecipientsByUserID(ctx, userID, 50, 0)
	if err != nil {
		return nil, err
	}
	result := make([]*models.Notification, 0, len(recipients))
	for _, rec := range recipients {
		if rec.Notification != nil {
			n := *rec.Notification
			n.ID = rec.NotificationID
			n.UserID = rec.UserID
			n.Message = rec.Notification.Message
			n.Read = rec.IsRead
			result = append(result, &n)
		}
	}
	return result, nil
}

func (r *notificationRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	if notification == nil {
		return errors.New("notification cannot be nil")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if rec := recover(); rec != nil {
				tx.Rollback()
			}
		}()
		return tx.Create(notification).Error
	})
}

func (r *notificationRepository) CreateRecipients(ctx context.Context, recipients []models.NotificationRecipient) error {
	if len(recipients) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if rec := recover(); rec != nil {
				tx.Rollback()
			}
		}()
		return tx.Create(&recipients).Error
	})
}

func (r *notificationRepository) FindRecipientsByUserID(ctx context.Context, userID string, limit, offset int) ([]models.NotificationRecipient, error) {
	var list []models.NotificationRecipient
	query := r.db.WithContext(ctx).
		Preload("Notification").
		Where("user_id = ?", userID).
		Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&list).Error
	return list, err
}

func (r *notificationRepository) FindRecipientByID(ctx context.Context, recipientID, userID string) (*models.NotificationRecipient, error) {
	var recipient models.NotificationRecipient
	err := r.db.WithContext(ctx).
		Preload("Notification").
		Where(queryRecipientAndUser, recipientID, userID).
		First(&recipient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}
	return &recipient, nil
}

func (r *notificationRepository) CountUnreadByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.NotificationRecipient{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) MarkRecipientAsRead(ctx context.Context, recipientID, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if rec := recover(); rec != nil {
				tx.Rollback()
			}
		}()
		now := time.Now()
		res := tx.Model(&models.NotificationRecipient{}).
			Where(queryRecipientAndUser, recipientID, userID).
			Updates(map[string]interface{}{
				"is_read": true,
				"read_at": &now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("recipient record not found")
		}
		return nil
	})
}

func (r *notificationRepository) MarkAllRecipientsAsRead(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if rec := recover(); rec != nil {
				tx.Rollback()
			}
		}()
		now := time.Now()
		return tx.Model(&models.NotificationRecipient{}).
			Where("user_id = ? AND is_read = ?", userID, false).
			Updates(map[string]interface{}{
				"is_read": true,
				"read_at": &now,
			}).Error
	})
}

func (r *notificationRepository) HardDeleteRecipient(ctx context.Context, recipientID, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if rec := recover(); rec != nil {
				tx.Rollback()
			}
		}()
		res := tx.Unscoped().
			Where(queryRecipientAndUser, recipientID, userID).
			Delete(&models.NotificationRecipient{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("recipient record not found")
		}
		return nil
	})
}

func (r *notificationRepository) FindPendingRecipients(ctx context.Context, limit int) ([]models.NotificationRecipient, error) {
	var pending []models.NotificationRecipient
	query := r.db.WithContext(ctx).
		Preload("Notification").
		Where("email_status = ? OR push_status = ?", models.DeliveryStatusPending, models.DeliveryStatusPending).
		Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&pending).Error
	return pending, err
}

func (r *notificationRepository) UpdateDeliveryStatus(
	ctx context.Context,
	recipientID string,
	emailStatus, pushStatus models.DeliveryStatus,
	emailMsgID, pushMsgID, errorMsg *string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if rec := recover(); rec != nil {
				tx.Rollback()
			}
		}()
		updates := map[string]interface{}{
			"email_status": emailStatus,
			"push_status":  pushStatus,
		}
		if emailMsgID != nil {
			updates["email_message_id"] = emailMsgID
		}
		if pushMsgID != nil {
			updates["push_message_id"] = pushMsgID
		}
		if errorMsg != nil {
			updates["error_message"] = errorMsg
		}
		return tx.Model(&models.NotificationRecipient{}).
			Where("recipient_id = ?", recipientID).
			Updates(updates).Error
	})
}

func (r *notificationRepository) GetPreference(ctx context.Context, userID string, category models.NotificationCategory) (*models.NotificationPreference, error) {
	var pref models.NotificationPreference
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND category = ?", userID, category).
		First(&pref).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pref, nil
}

func (r *notificationRepository) GetPreferencesByUserID(ctx context.Context, userID string) ([]models.NotificationPreference, error) {
	var list []models.NotificationPreference
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("category ASC").
		Find(&list).Error
	return list, err
}

func (r *notificationRepository) UpsertPreference(ctx context.Context, pref *models.NotificationPreference) error {
	if pref == nil {
		return errors.New("preference cannot be nil")
	}
	if pref.ID == "" {
		pref.ID = generatePrefID()
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if rec := recover(); rec != nil {
				tx.Rollback()
			}
		}()
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "category"}},
			DoUpdates: clause.AssignmentColumns([]string{"email_enabled", "push_enabled", "in_app_enabled", "updated_at"}),
		}).Create(pref).Error
	})
}

// Memory fallback implementation for isolated local testing without a database
type memoryNotificationRepository struct {
	mu            sync.RWMutex
	notifications map[string]*models.Notification
	recipients    []models.NotificationRecipient
	preferences   map[string]*models.NotificationPreference
}

// NewMemoryNotificationRepository returns an in-memory repository for development fallback and unit tests.
func NewMemoryNotificationRepository() NotificationRepository {
	return &memoryNotificationRepository{
		notifications: make(map[string]*models.Notification),
		preferences:   make(map[string]*models.NotificationPreference),
	}
}

func (m *memoryNotificationRepository) FindByUserID(_ context.Context, userID string) ([]*models.Notification, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*models.Notification
	for _, rec := range m.recipients {
		if rec.UserID == userID {
			if n, ok := m.notifications[rec.NotificationID]; ok {
				cp := *n
				cp.ID = rec.NotificationID
				cp.UserID = rec.UserID
				cp.Message = n.Message
				cp.Read = rec.IsRead
				result = append(result, &cp)
			}
		}
	}
	if len(result) == 0 {
		result = append(result, &models.Notification{
			ID:        "ntf-1",
			UserID:    userID,
			Message:   "Bienvenido a la PWA Buen Vivir",
			Read:      false,
			CreatedAt: time.Now(),
		})
	}
	return result, nil
}

func (m *memoryNotificationRepository) CreateNotification(_ context.Context, n *models.Notification) error {
	if n == nil {
		return errors.New("notification cannot be nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifications[n.ID] = n
	return nil
}

func (m *memoryNotificationRepository) CreateRecipients(_ context.Context, recs []models.NotificationRecipient) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recipients = append(m.recipients, recs...)
	return nil
}

func (m *memoryNotificationRepository) FindRecipientsByUserID(_ context.Context, userID string, limit, offset int) ([]models.NotificationRecipient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []models.NotificationRecipient
	for _, rec := range m.recipients {
		if rec.UserID == userID {
			recCopy := rec
			recCopy.Notification = m.notifications[rec.NotificationID]
			list = append(list, recCopy)
		}
	}
	if offset > len(list) {
		return nil, nil
	}
	list = list[offset:]
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (m *memoryNotificationRepository) FindRecipientByID(_ context.Context, recipientID, userID string) (*models.NotificationRecipient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, rec := range m.recipients {
		if rec.RecipientID == recipientID && rec.UserID == userID {
			recCopy := rec
			recCopy.Notification = m.notifications[rec.NotificationID]
			return &recCopy, nil
		}
	}
	return nil, errors.New("notification not found")
}

func (m *memoryNotificationRepository) CountUnreadByUserID(_ context.Context, userID string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, rec := range m.recipients {
		if rec.UserID == userID && !rec.IsRead {
			count++
		}
	}
	return count, nil
}

func (m *memoryNotificationRepository) MarkRecipientAsRead(_ context.Context, recipientID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.recipients {
		if m.recipients[i].RecipientID == recipientID && m.recipients[i].UserID == userID {
			m.recipients[i].IsRead = true
			now := time.Now()
			m.recipients[i].ReadAt = &now
			return nil
		}
	}
	return errors.New("recipient record not found")
}

func (m *memoryNotificationRepository) MarkAllRecipientsAsRead(_ context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for i := range m.recipients {
		if m.recipients[i].UserID == userID && !m.recipients[i].IsRead {
			m.recipients[i].IsRead = true
			m.recipients[i].ReadAt = &now
		}
	}
	return nil
}

func (m *memoryNotificationRepository) HardDeleteRecipient(_ context.Context, recipientID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.recipients {
		if m.recipients[i].RecipientID == recipientID && m.recipients[i].UserID == userID {
			m.recipients = append(m.recipients[:i], m.recipients[i+1:]...)
			return nil
		}
	}
	return errors.New("recipient record not found")
}

func (m *memoryNotificationRepository) FindPendingRecipients(_ context.Context, limit int) ([]models.NotificationRecipient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var pending []models.NotificationRecipient
	for _, rec := range m.recipients {
		if rec.EmailStatus == models.DeliveryStatusPending || rec.PushStatus == models.DeliveryStatusPending {
			recCopy := rec
			recCopy.Notification = m.notifications[rec.NotificationID]
			pending = append(pending, recCopy)
			if limit > 0 && len(pending) >= limit {
				break
			}
		}
	}
	return pending, nil
}

func (m *memoryNotificationRepository) UpdateDeliveryStatus(
	_ context.Context,
	recipientID string,
	emailStatus, pushStatus models.DeliveryStatus,
	emailMsgID, pushMsgID, errorMsg *string,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.recipients {
		if m.recipients[i].RecipientID == recipientID {
			m.recipients[i].EmailStatus = emailStatus
			m.recipients[i].PushStatus = pushStatus
			m.recipients[i].EmailMessageID = emailMsgID
			m.recipients[i].PushMessageID = pushMsgID
			m.recipients[i].ErrorMessage = errorMsg
			return nil
		}
	}
	return errors.New("recipient record not found")
}

func (m *memoryNotificationRepository) GetPreference(_ context.Context, userID string, category models.NotificationCategory) (*models.NotificationPreference, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := userID + ":" + string(category)
	if p, ok := m.preferences[key]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *memoryNotificationRepository) GetPreferencesByUserID(_ context.Context, userID string) ([]models.NotificationPreference, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []models.NotificationPreference
	for _, p := range m.preferences {
		if p.UserID == userID {
			list = append(list, *p)
		}
	}
	return list, nil
}

func (m *memoryNotificationRepository) UpsertPreference(_ context.Context, pref *models.NotificationPreference) error {
	if pref == nil {
		return errors.New("preference cannot be nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if pref.ID == "" {
		pref.ID = generatePrefID()
	}
	key := pref.UserID + ":" + string(pref.Category)
	m.preferences[key] = pref
	return nil
}

func generatePrefID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("prf-%s", hex.EncodeToString(b))
}
