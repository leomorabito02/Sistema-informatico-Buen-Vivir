package models

// NotificationCategory classifies the business domain of the notification.
type NotificationCategory string

const (
	CategorySecurity  NotificationCategory = "security"
	CategoryEducation NotificationCategory = "education"
	CategoryLabor     NotificationCategory = "labor"
	CategoryCampaigns NotificationCategory = "campaigns"
	CategorySystem    NotificationCategory = "system"
)

// NotificationPriority defines delivery urgency.
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityNormal   NotificationPriority = "normal"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// SenderType identifies whether notification originated from an authenticated user or automated system.
type SenderType string

const (
	SenderTypeSystem SenderType = "system"
	SenderTypeUser   SenderType = "user"
)

// TargetType defines the recipient resolution strategy (individual user, all users in a role, or broadcast).
type TargetType string

const (
	TargetTypeIndividual TargetType = "individual"
	TargetTypeRole       TargetType = "role"
	TargetTypeBroadcast  TargetType = "broadcast"
)

// DeliveryStatus tracks individual delivery outcomes across Push and Email channels.
type DeliveryStatus string

const (
	DeliveryStatusPending      DeliveryStatus = "pending"
	DeliveryStatusSent         DeliveryStatus = "sent"
	DeliveryStatusFailed       DeliveryStatus = "failed"
	DeliveryStatusSkipped      DeliveryStatus = "skipped"
	DeliveryStatusNotRequested DeliveryStatus = "not_requested"
)
