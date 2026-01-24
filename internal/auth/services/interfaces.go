package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/teacinema-go/auth-service/internal/auth/dto"
	"github.com/teacinema-go/auth-service/internal/auth/entities"
	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, arg dto.CreateAccountParams) error
	GetAccountByEmail(ctx context.Context, email valueobject.Identifier) (*entities.Account, error)
	GetAccountByPhone(ctx context.Context, phone valueobject.Identifier) (*entities.Account, error)
	UpdateAccountIsEmailVerified(ctx context.Context, arg dto.UpdateAccountIsEmailVerifiedParams) error
	UpdateAccountIsPhoneVerified(ctx context.Context, arg dto.UpdateAccountIsPhoneVerifiedParams) error
}

type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, arg dto.CreateRefreshTokenParams) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)
	DeleteRefreshTokenByHash(ctx context.Context, tokenHash string) error
	DeleteRefreshTokensByAccountID(ctx context.Context, accountID uuid.UUID) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type TxManager interface {
	WithTransaction(ctx context.Context, fn func(repos TxRepositories) (any, error)) (any, error)
}

type TxRepositories interface {
	Account() AccountRepository
	RefreshToken() RefreshTokenRepository
}
