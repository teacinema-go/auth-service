package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateRefreshTokenParams struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"account_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
}
