package entities

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID              uuid.UUID `json:"id"`
	Phone           *string   `json:"phone"`
	Email           *string   `json:"email"`
	IsPhoneVerified bool      `json:"is_phone_verified"`
	IsEmailVerified bool      `json:"is_email_verified"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
