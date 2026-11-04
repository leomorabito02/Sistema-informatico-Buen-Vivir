package rules

import (
	"buenvivir/notifications/internal/models"
)

// AudienceType indicates how target recipients should be resolved.
type AudienceType string

const (
	AudienceIndividual  AudienceType = "individual"
	AudienceParticipant AudienceType = "participant"
	AudienceRole        AudienceType = "role"
)

// NotificationPlan represents the structured intent to deliver a notification.
type NotificationPlan struct {
	SourceService    string
	EventType        string
	Category         models.NotificationCategory
	Priority         models.NotificationPriority
	AudienceType     AudienceType
	TargetRole       *string
	RecipientUserIDs []string
	ParticipantID    string // When AudienceType == AudienceParticipant
	Title            string
	Message          string
	ActionURL        *string
	Data             string
	SendEmail        bool
	SendPush         bool
	SendInApp        bool
	IsCritical       bool
	DedupeKey        string
}

