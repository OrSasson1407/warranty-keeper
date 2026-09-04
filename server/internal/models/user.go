package models

import "github.com/google/uuid"

type User struct {
	BaseModel
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"column:password_hash" json:"-"`
	GoogleID     string     `gorm:"column:google_id;index" json:"-"`
	FullName     string     `json:"full_name"`
	HouseholdID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"household_id"`
}
