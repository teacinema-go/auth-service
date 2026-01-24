package entities

import (
	"time"

	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
)

type Account struct {
	ID        valueobject.ID   `json:"id"`
	Phone     *string          `json:"phone"`
	Email     *string          `json:"email"`
	Role      valueobject.Role `json:"role"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}
