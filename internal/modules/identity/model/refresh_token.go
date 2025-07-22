package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash string     `json:"token_hash" db:"token_hash"`
	UserAgent string     `json:"user_agent" db:"user_agent"`
	IPAddress string     `json:"ip_address" db:"ip_address"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at" db:"revoked_at"`
}
