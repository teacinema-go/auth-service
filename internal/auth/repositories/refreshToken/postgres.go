package refreshToken

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/teacinema-go/auth-service/internal/auth/dto"
	"github.com/teacinema-go/auth-service/internal/auth/entities"
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	"github.com/teacinema-go/auth-service/internal/infra/storage/postgres/sqlc"
)

type PostgresRefreshTokenRepository struct {
	q sqlc.Querier
}

func NewPostgresRefreshTokenRepository(q sqlc.Querier) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{q: q}
}

func (r *PostgresRefreshTokenRepository) CreateRefreshToken(ctx context.Context, arg dto.CreateRefreshTokenParams) error {
	param := sqlc.CreateRefreshTokenParams{
		ID:        arg.ID,
		AccountID: arg.AccountID,
		TokenHash: arg.TokenHash,
		ExpiresAt: arg.ExpiresAt,
	}

	return r.q.CreateRefreshToken(ctx, param)
}

func (r *PostgresRefreshTokenRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	token, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrRefreshTokenNotFound
		}
		return nil, err
	}
	return mapSqlcRefreshToken(token), nil
}

func (r *PostgresRefreshTokenRepository) DeleteRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	return r.q.DeleteRefreshTokenByHash(ctx, tokenHash)
}

func (r *PostgresRefreshTokenRepository) DeleteRefreshTokensByAccountID(ctx context.Context, accountID uuid.UUID) error {
	return r.q.DeleteRefreshTokensByAccountID(ctx, accountID)
}

func mapSqlcRefreshToken(a sqlc.RefreshToken) *entities.RefreshToken {
	return &entities.RefreshToken{
		ID:        a.ID,
		AccountID: a.AccountID,
		TokenHash: a.TokenHash,
		ExpiresAt: a.ExpiresAt,
	}
}
