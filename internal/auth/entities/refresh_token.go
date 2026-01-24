package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
)

type RefreshToken struct {
	ID        valueobject.ID `json:"id"`
	AccountID uuid.UUID      `json:"account_id"`
	TokenHash string         `json:"token_hash"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
}
