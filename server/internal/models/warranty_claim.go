package models

import (
	"time"

	"github.com/google/uuid"
)

type ClaimStatus string

const (
	ClaimStatusOpen       ClaimStatus = "open"
	ClaimStatusInProgress ClaimStatus = "in_progress"
	ClaimStatusClosed     ClaimStatus = "closed"
)

type WarrantyClaim struct {
	BaseModel
	ProductID        uuid.UUID   `gorm:"type:uuid;index;not null" json:"product_id"`
	IssueDescription string      `gorm:"type:text;not null" json:"issue_description"`
	Status           ClaimStatus `gorm:"type:varchar(20);default:open" json:"status"`
	ResolvedAt       *time.Time  `json:"resolved_at"`
}
