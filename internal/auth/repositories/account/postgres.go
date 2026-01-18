package account

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/teacinema-go/auth-service/internal/auth/dto"
	"github.com/teacinema-go/auth-service/internal/auth/entities"
	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	"github.com/teacinema-go/auth-service/internal/infra/storage/postgres/sqlc"
)

type PostgresAccountRepository struct {
	q sqlc.Querier
}

func NewPostgresAccountRepository(q sqlc.Querier) *PostgresAccountRepository {
	return &PostgresAccountRepository{q: q}
}

func (r *PostgresAccountRepository) CreateAccount(ctx context.Context, arg dto.CreateAccountParams) error {
	param := sqlc.CreateAccountParams{
		ID:    arg.ID,
		Email: arg.Email,
		Phone: arg.Phone,
	}
	return r.q.CreateAccount(ctx, param)
}

func (r *PostgresAccountRepository) GetAccountByEmail(ctx context.Context, email valueobject.Identifier) (*entities.Account, error) {
	strEmail := string(email)
	acc, err := r.q.GetAccountByEmail(ctx, &strEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrAccountNotFound
		}
		return nil, err
	}
	return mapSqlcAccount(acc), nil
}

func (r *PostgresAccountRepository) GetAccountByPhone(ctx context.Context, phone valueobject.Identifier) (*entities.Account, error) {
	strPhone := string(phone)
	acc, err := r.q.GetAccountByPhone(ctx, &strPhone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrAccountNotFound
		}
		return nil, err
	}
	return mapSqlcAccount(acc), nil
}

func (r *PostgresAccountRepository) UpdateAccountIsEmailVerified(ctx context.Context, arg dto.UpdateAccountIsEmailVerifiedParams) error {
	param := sqlc.UpdateAccountIsEmailVerifiedParams{
		ID:              arg.ID,
		IsEmailVerified: arg.IsEmailVerified,
	}
	return r.q.UpdateAccountIsEmailVerified(ctx, param)
}

func (r *PostgresAccountRepository) UpdateAccountIsPhoneVerified(ctx context.Context, arg dto.UpdateAccountIsPhoneVerifiedParams) error {
	param := sqlc.UpdateAccountIsPhoneVerifiedParams{
		ID:              arg.ID,
		IsPhoneVerified: arg.IsPhoneVerified,
	}
	return r.q.UpdateAccountIsPhoneVerified(ctx, param)
}

func mapSqlcAccount(a sqlc.Account) *entities.Account {
	return &entities.Account{
		ID:              a.ID,
		Email:           a.Email,
		Phone:           a.Phone,
		IsPhoneVerified: a.IsPhoneVerified,
		IsEmailVerified: a.IsEmailVerified,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}
