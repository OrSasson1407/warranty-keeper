package models

import "github.com/google/uuid"

// Household groups up to 2 users (MVP limit, enforced in the service layer,
// not the schema, so the cap can be relaxed later without a migration).
type Household struct {
	BaseModel
	Name       string    `gorm:"not null" json:"name"`
	CreatedBy  uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	InviteCode string    `gorm:"uniqueIndex;not null" json:"invite_code"`
	Users      []User    `gorm:"foreignKey:HouseholdID" json:"-"`
}
