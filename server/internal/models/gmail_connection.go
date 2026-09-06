package models

import (
	"time"

	"github.com/google/uuid"
)

// GmailConnection stores one user's opt-in Gmail OAuth grant. Tokens are
// encrypted at rest (see internal/crypto) since, unlike a password hash,
// they must be recoverable in plaintext to call the Gmail API on the
// user's behalf during a scan.
type GmailConnection struct {
	BaseModel
	UserID                uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	GmailAddress          string     `json:"gmail_address"`
	EncryptedAccessToken  string     `gorm:"type:text" json:"-"`
	EncryptedRefreshToken string     `gorm:"type:text" json:"-"`
	TokenExpiry           time.Time  `json:"-"`
	LastScanAt            *time.Time `json:"last_scan_at"`
}
