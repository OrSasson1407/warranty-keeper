package models

// ManufacturerContact holds support contact info for a brand, keyed by the
// brand name as it appears on a Product (or a Receipt's parsed vendor).
// Server-managed so it can be updated without an app release, replacing the
// static list that used to be bundled into the mobile app.
type ManufacturerContact struct {
	BaseModel
	Brand   string `gorm:"uniqueIndex;not null" json:"brand"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
}
