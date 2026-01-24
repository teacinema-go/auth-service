package valueobject

import (
	"github.com/teacinema-go/auth-service/internal/errors"
	accountv1 "github.com/teacinema-go/contracts/gen/go/account/v1"
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

func (r Role) ToProto() accountv1.Role {
	if r == RoleUser {
		return accountv1.Role_USER
	}

	return accountv1.Role_ADMIN
}
