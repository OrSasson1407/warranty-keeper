package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	BaseModel
	HouseholdID       uuid.UUID  `gorm:"type:uuid;index;not null" json:"household_id"`
	Name              string     `gorm:"not null" json:"name"`
	Category          string     `gorm:"index" json:"category"`
	Brand             string     `json:"brand"`
	PurchaseDate      time.Time  `json:"purchase_date"`
	Price             *float64   `gorm:"type:numeric(12,2)" json:"price"`
	Room              string     `json:"room"`
	WarrantyExpiresAt time.Time  `gorm:"index" json:"warranty_expires_at"`
	WarrantyUncertain bool       `json:"warranty_uncertain"`
	PhotoURL          string     `json:"photo_url"`
	ReceiptID         *uuid.UUID `gorm:"type:uuid;index" json:"receipt_id"`
}
