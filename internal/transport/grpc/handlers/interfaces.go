package handlers

import (
	"context"

	"github.com/teacinema-go/auth-service/internal/auth/entities"
	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
	"github.com/teacinema-go/passport"
)

type AuthService interface {
	GetAccount(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (*entities.Account, error)
	CreateAccount(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) error
	GenerateOtp(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (string, error)
	VerifyOtp(ctx context.Context, otp string, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (bool, error)
	CompleteAccountVerification(ctx context.Context, acc *entities.Account, refreshToken *passport.Token) error
	RotateRefreshToken(ctx context.Context, oldTokenValue string, newToken *passport.Token) error
}
