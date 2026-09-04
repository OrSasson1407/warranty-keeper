package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel is embedded by every entity: a UUID primary key generated
// app-side (no reliance on a Postgres UUID extension) plus timestamps.
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
