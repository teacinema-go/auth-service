package handlers

import (
	"context"

	"github.com/teacinema-go/auth-service/internal/auth/dto"
	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
	"github.com/teacinema-go/passport"
)

type AuthService interface {
	AccountExists(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (bool, error)
	GenerateOtp(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (string, error)
	VerifyOtp(ctx context.Context, otp string, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (bool, error)
	CreateAccountWithTokens(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (dto.Tokens, error)
	VerifyToken(token *passport.Token) bool
	RotateRefreshToken(ctx context.Context, oldToken *passport.Token) (dto.Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
}
