package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductCost is a post-purchase cost logged against a product (an
// out-of-pocket repair, an accessory, etc.), for the basic total-cost-of-
// ownership view: purchase Price + sum of these. Deliberately minimal --
// no depreciation, no currency conversion, no category benchmarks.
type ProductCost struct {
	BaseModel
	ProductID   uuid.UUID `gorm:"type:uuid;index;not null" json:"product_id"`
	Amount      float64   `gorm:"type:numeric(12,2);not null" json:"amount"`
	Description string    `json:"description"`
	IncurredAt  time.Time `json:"incurred_at"`
}
