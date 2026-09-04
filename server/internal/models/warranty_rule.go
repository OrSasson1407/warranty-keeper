package models

// WarrantyRule maps a product category (+ optional brand) to a default
// warranty duration. An empty Brand means "general rule for this category".
// See the rules engine fallback order in internal/warranty.
type WarrantyRule struct {
	BaseModel
	Category       string `gorm:"index:idx_warranty_rule_lookup,priority:1;not null" json:"category"`
	Brand          string `gorm:"index:idx_warranty_rule_lookup,priority:2" json:"brand"`
	DurationMonths int    `gorm:"not null" json:"duration_months"`
	Source         string `gorm:"not null;default:default" json:"source"`
}
