package valueobject

import (
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	"github.com/teacinema-go/core/validator"
)

type Identifier string

func (i Identifier) Validate(identifierType IdentifierType) error {
	switch identifierType {
	case IdentifierTypeEmail:
		if !validator.IsValidEmail(string(i)) {
			return appErrors.ErrInvalidEmail
		}
	case IdentifierTypePhone:
		if !validator.IsValidE164Phone(string(i)) {
			return appErrors.ErrInvalidE164Phone
		}
	default:
		return appErrors.ErrInvalidIdentifierType
	}

	return nil
}
