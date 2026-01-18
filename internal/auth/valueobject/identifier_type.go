package valueobject

import (
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
)

type IdentifierType string

const (
	IdentifierTypePhone IdentifierType = "phone"
	IdentifierTypeEmail IdentifierType = "email"
)

func (it IdentifierType) ToProto() authv1.IdentifierType {
	switch it {
	case IdentifierTypeEmail:
		return authv1.IdentifierType_EMAIL
	case IdentifierTypePhone:
		return authv1.IdentifierType_PHONE
	default:
		return authv1.IdentifierType_IDENTIFIER_TYPE_UNSPECIFIED
	}
}

func (it IdentifierType) FromProto(protoIdentifierType authv1.IdentifierType) (IdentifierType, error) {
	switch protoIdentifierType {
	case authv1.IdentifierType_EMAIL:
		return IdentifierTypeEmail, nil
	case authv1.IdentifierType_PHONE:
		return IdentifierTypePhone, nil
	}

	return "", appErrors.ErrInvalidIdentifierType
}
