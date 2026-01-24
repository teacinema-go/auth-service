package errors

import (
	"errors"
)

var (
	ErrNotFound              = errors.New("not found")
	ErrInvalidEmail          = errors.New("invalid email")
	ErrInvalidE164Phone      = errors.New("invalid e.164 phone number")
	ErrInvalidIdentifierType = errors.New("invalid identifier type")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrRefreshTokenNotFound  = errors.New("refresh token not found")
	ErrAccountNotFound       = errors.New("account not found")
	ErrAccountAlreadyExists  = errors.New("account already exists")
	ErrInvalidRole           = errors.New("invalid role")
)
