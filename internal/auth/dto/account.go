package dto

import (
	"github.com/google/uuid"
)

type CreateAccountParams struct {
	ID    uuid.UUID `json:"id"`
	Phone *string   `json:"phone"`
	Email *string   `json:"email"`
}

type UpdateAccountIsEmailVerifiedParams struct {
	ID              uuid.UUID `json:"id"`
	IsEmailVerified bool      `json:"is_email_verified"`
}

type UpdateAccountIsPhoneVerifiedParams struct {
	ID              uuid.UUID `json:"id"`
	IsPhoneVerified bool      `json:"is_phone_verified"`
}
