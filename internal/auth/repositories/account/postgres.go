package account

import (
	"context"
	"database/sql"
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
		ID:    arg.ID.ToUUID(),
		Email: arg.Email,
		Phone: arg.Phone,
		Role:  string(arg.Role),
	}
	_, err := r.q.CreateAccount(ctx, param)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.ErrAccountAlreadyExists
		}
		return err
	}
	return nil
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
	return mapSqlcAccount(acc)
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
	return mapSqlcAccount(acc)
}

func (r *PostgresAccountRepository) GetAccountByID(ctx context.Context, accountID valueobject.ID) (*entities.Account, error) {
	acc, err := r.q.GetAccountByID(ctx, accountID.ToUUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appErrors.ErrAccountNotFound
		}
		return nil, err
	}

	return mapSqlcAccount(acc)
}

func (r *PostgresAccountRepository) AccountExistsByEmail(ctx context.Context, email valueobject.Identifier) (bool, error) {
	strEmail := string(email)
	return r.q.AccountExistsByEmail(ctx, &strEmail)
}

func (r *PostgresAccountRepository) AccountExistsByPhone(ctx context.Context, phone valueobject.Identifier) (bool, error) {
	strPhone := string(phone)
	return r.q.AccountExistsByPhone(ctx, &strPhone)
}

func mapSqlcAccount(a sqlc.Account) (*entities.Account, error) {
	role := valueobject.Role(a.Role)
	err := role.Validate()
	if err != nil {
		return nil, err
	}

	return &entities.Account{
		ID:        valueobject.ID(a.ID),
		Email:     a.Email,
		Phone:     a.Phone,
		Role:      role,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}, nil
}
