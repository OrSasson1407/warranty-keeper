package models

import (
	"time"

	"github.com/google/uuid"
)

type ReceiptStatus string

const (
	ReceiptStatusPending   ReceiptStatus = "pending"
	ReceiptStatusProcessed ReceiptStatus = "processed"
	ReceiptStatusFailed    ReceiptStatus = "failed"
)

const (
	ReceiptSourcePhoto ReceiptSource = "photo"
	ReceiptSourceGmail ReceiptSource = "gmail"
)

type ReceiptSource string

type Receipt struct {
	BaseModel
	HouseholdID    uuid.UUID     `gorm:"type:uuid;index;not null" json:"household_id"`
	ImageURL       string        `json:"image_url"`
	RawOCRText     string        `gorm:"type:text" json:"raw_ocr_text"`
	ParsedVendor   string        `json:"parsed_vendor"`
	ParsedDate     *time.Time    `json:"parsed_date"`
	ParsedAmount   *float64      `gorm:"type:numeric(12,2)" json:"parsed_amount"`
	Status         ReceiptStatus `gorm:"type:varchar(20);default:pending;index" json:"status"`
	Confidence     float64       `json:"confidence"`
	Source         ReceiptSource `gorm:"type:varchar(20);default:photo;index" json:"source"`
	GmailMessageID string        `gorm:"index" json:"-"`
}
