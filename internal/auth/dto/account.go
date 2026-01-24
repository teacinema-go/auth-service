package dto

import (
	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
)

type CreateAccountParams struct {
	ID    valueobject.ID
	Phone *string
	Email *string
	Role  valueobject.Role
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int32
}
