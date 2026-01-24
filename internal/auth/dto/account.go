package dto

import (
	"github.com/google/uuid"
)

type CreateAccountParams struct {
	ID    uuid.UUID
	Phone *string
	Email *string
}

type UpdateAccountIsEmailVerifiedParams struct {
	ID              uuid.UUID
	IsEmailVerified bool
}

type UpdateAccountIsPhoneVerifiedParams struct {
	ID              uuid.UUID
	IsPhoneVerified bool
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int32
}
