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
	// WarningDays distinguishes which expiry-warning tier (30/14/3) this log
	// row covers, so a product can be notified at each tier independently
	// instead of the first send blocking all the others. Left at 0 (its zero
	// value) for non-tiered notification types like annual_summary.
	WarningDays int `json:"warning_days"`
}
