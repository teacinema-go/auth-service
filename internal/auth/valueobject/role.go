package valueobject

import (
	"github.com/teacinema-go/auth-service/internal/errors"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func (r Role) Validate() error {
	if r != RoleUser && r != RoleAdmin {
		return errors.ErrInvalidRole
	}

	return nil
}
