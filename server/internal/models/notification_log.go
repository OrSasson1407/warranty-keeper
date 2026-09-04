package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeExpiryWarning NotificationType = "expiry_warning"
	NotificationTypeAnnualSummary NotificationType = "annual_summary"
)

type NotificationLog struct {
	BaseModel
	UserID    uuid.UUID        `gorm:"type:uuid;index;not null" json:"user_id"`
	ProductID uuid.UUID        `gorm:"type:uuid;index;not null" json:"product_id"`
	Type      NotificationType `gorm:"type:varchar(30);not null" json:"type"`
	SentAt    time.Time        `json:"sent_at"`
}
