package txmanager

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teacinema-go/auth-service/internal/auth/repositories/account"
	"github.com/teacinema-go/auth-service/internal/auth/repositories/refreshToken"
	"github.com/teacinema-go/auth-service/internal/auth/services"
	"github.com/teacinema-go/auth-service/internal/infra/storage/postgres/sqlc"
	"github.com/teacinema-go/core/logger"
)

type PostgresTxManager struct {
	pool *pgxpool.Pool
}

func NewPostgresTxManager(pool *pgxpool.Pool) *PostgresTxManager {
	return &PostgresTxManager{pool: pool}
}

func (m *PostgresTxManager) WithTransaction(ctx context.Context, fn func(repos services.TxRepositories) (any, error)) (any, error) {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			logger.Error("rollback transaction failed", "error", err)
		}
	}(tx, ctx)

	repos := newTxRepositories(sqlc.New(tx))

	if res, err := fn(repos); err != nil {
		return res, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction failed: %w", err)
	}

	return nil, nil
}

type txRepositories struct {
	accountRepo      services.AccountRepository
	refreshTokenRepo services.RefreshTokenRepository
}

func newTxRepositories(q sqlc.Querier) *txRepositories {
	return &txRepositories{
		accountRepo:      account.NewPostgresAccountRepository(q),
		refreshTokenRepo: refreshToken.NewPostgresRefreshTokenRepository(q),
	}
}

func (r *txRepositories) Account() services.AccountRepository {
	return r.accountRepo
}

func (r *txRepositories) RefreshToken() services.RefreshTokenRepository {
	return r.refreshTokenRepo
}
