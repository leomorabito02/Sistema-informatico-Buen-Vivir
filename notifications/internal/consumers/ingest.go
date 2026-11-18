package consumers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"buenvivir/notifications/internal/audience"
	"buenvivir/notifications/internal/delivery"
	"buenvivir/notifications/internal/models"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/rules"
)

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}

// PersistPlans resolves audience and stores notification and recipients atomically.
func PersistPlans(
	ctx context.Context,
	plans []rules.NotificationPlan,
	resolver audience.Resolver,
	repo repositories.NotificationRepository,
	worker *delivery.Worker,
) error {
	for _, plan := range plans {
		userIDs, err := resolver.ResolveUserIDs(ctx, plan)
		if err != nil {
			return fmt.Errorf("failed to resolve audience for event %s: %w", plan.EventType, err)
		}
		if len(userIDs) == 0 {
			continue
		}

		prefMap, _ := repo.GetPreferencesByUserIDsAndCategory(ctx, userIDs, plan.Category)
		if prefMap == nil {
			prefMap = make(map[string]*models.NotificationPreference)
		}

		notifID := generateID("ntf")
		notif := &models.Notification{
			ID:            notifID,
			SourceService: plan.SourceService,
			EventType:     plan.EventType,
			Category:      plan.Category,
			Priority:      plan.Priority,
			SenderType:    models.SenderTypeSystem,
			SenderID:      "events-consumer",
			TargetType:    models.TargetTypeIndividual,
			TargetRole:    plan.TargetRole,
			Title:         plan.Title,
			Message:       plan.Message,
			ActionURL:     plan.ActionURL,
			Data:          plan.Data,
			IsCritical:    plan.IsCritical,
			SendEmail:     plan.SendEmail,
			SendPush:      plan.SendPush,
			SendInApp:     plan.SendInApp,
			CreatedAt:     time.Now().UTC(),
		}

		isCritical := plan.IsCritical || plan.Category == models.CategorySecurity
		recipients := make([]models.NotificationRecipient, 0, len(userIDs))

		for _, uid := range userIDs {
			trimmedUID := strings.TrimSpace(uid)
			if trimmedUID == "" {
				continue
			}

			pref := prefMap[trimmedUID]
			emailAllowed := pref == nil || pref.AllowsChannel("email", isCritical)
			pushAllowed := pref == nil || pref.AllowsChannel("push", isCritical)
			inAppAllowed := pref == nil || pref.AllowsChannel("in_app", isCritical)

			if !isCritical && !emailAllowed && !pushAllowed && !inAppAllowed {
				continue
			}

			emailStatus := models.DeliveryStatusNotRequested
			if plan.SendEmail {
				if emailAllowed {
					emailStatus = models.DeliveryStatusPending
				} else {
					emailStatus = models.DeliveryStatusSkipped
				}
			}

			pushStatus := models.DeliveryStatusNotRequested
			if plan.SendPush {
				if pushAllowed {
					pushStatus = models.DeliveryStatusPending
				} else {
					pushStatus = models.DeliveryStatusSkipped
				}
			}

			now := time.Now().UTC()
			rec := models.NotificationRecipient{
				RecipientID:    generateID("rcp"),
				NotificationID: notifID,
				UserID:         trimmedUID,
				RoleID:         plan.TargetRole,
				IsRead:         false,
				EmailStatus:    emailStatus,
				PushStatus:     pushStatus,
				NextAttemptAt:  &now,
				DedupeKey:      plan.DedupeKey,
				CreatedAt:      now,
			}
			recipients = append(recipients, rec)
		}

		if len(recipients) > 0 {
			if err := repo.CreateNotificationAndRecipients(ctx, notif, recipients); err != nil {
				return fmt.Errorf("failed to persist notification and recipients: %w", err)
			}
		}
	}

	if worker != nil {
		worker.Notify()
	}

	return nil
}

