package models

import "github.com/google/uuid"

// WarrantyRuleReport is a user flag that a product's resolved warranty
// period looks wrong (see the "expand warranty_rules coverage" issue's
// community-correction half). This table IS the review queue -- a
// maintainer works through open reports directly rather than through a
// dedicated admin UI, which is out of scope for this first pass.
type WarrantyRuleReport struct {
	BaseModel
	ProductID uuid.UUID `gorm:"type:uuid;index;not null" json:"product_id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Note      string    `gorm:"type:text" json:"note"`
}
