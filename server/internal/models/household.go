package models

import "github.com/google/uuid"

const (
	HouseholdTierFree    = "free"
	HouseholdTierPremium = "premium"

	// FreeTierProductLimit is the free-tier product cap from the v2 scope
	// doc's basic Premium/freemium tier. Set well above the MVP's observed
	// healthy average (5+ products/active user) so it doesn't suppress the
	// exact engagement behavior the MVP validated.
	FreeTierProductLimit = 20
)

// Household groups up to 2 users (MVP limit, enforced in the service layer,
// not the schema, so the cap can be relaxed later without a migration).
type Household struct {
	BaseModel
	Name       string    `gorm:"not null" json:"name"`
	CreatedBy  uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	InviteCode string    `gorm:"uniqueIndex;not null" json:"invite_code"`
	Tier       string    `gorm:"not null;default:free" json:"tier"`
	Users      []User    `gorm:"foreignKey:HouseholdID" json:"-"`
}
