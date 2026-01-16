package repository

import (
	"context"

	"github.com/teacinema-go/auth-service/internal/auth/dto"
	"github.com/teacinema-go/auth-service/internal/auth/entity"
	"github.com/teacinema-go/auth-service/internal/infra/storage/postgres/sqlc/accounts"
)

type PostgresRepository struct {
	q accounts.Querier
}

func NewPostgresRepository(q accounts.Querier) *PostgresRepository {
	return &PostgresRepository{q: q}
}

func (r *PostgresRepository) Create(ctx context.Context, arg dto.CreateAccountParams) (*entity.Account, error) {
	param := accounts.CreateParams{
		ID:    arg.ID,
		Email: arg.Email,
		Phone: arg.Phone,
	}
	acc, err := r.q.Create(ctx, param)
	if err != nil {
		return &entity.Account{}, err
	}
	return mapSqlcAccount(acc), nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email *string) (*entity.Account, error) {
	acc, err := r.q.GetByEmail(ctx, email)
	if err != nil {
		return &entity.Account{}, err
	}
	return mapSqlcAccount(acc), nil
}

func (r *PostgresRepository) GetByPhone(ctx context.Context, phone *string) (*entity.Account, error) {
	acc, err := r.q.GetByPhone(ctx, phone)
	if err != nil {
		return &entity.Account{}, err
	}
	return mapSqlcAccount(acc), nil
}

func (r *PostgresRepository) UpdateIsEmailVerified(ctx context.Context, arg dto.UpdateAccountIsEmailVerifiedParams) error {
	param := accounts.UpdateIsEmailVerifiedParams{
		ID:              arg.ID,
		IsEmailVerified: arg.IsEmailVerified,
	}
	err := r.q.UpdateIsEmailVerified(ctx, param)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) UpdateIsPhoneVerified(ctx context.Context, arg dto.UpdateAccountIsPhoneVerifiedParams) error {
	param := accounts.UpdateIsPhoneVerifiedParams{
		ID:              arg.ID,
		IsPhoneVerified: arg.IsPhoneVerified,
	}
	err := r.q.UpdateIsPhoneVerified(ctx, param)
	if err != nil {
		return err
	}
	return nil
}

func mapSqlcAccount(a accounts.Account) *entity.Account {
	return &entity.Account{
		ID:              a.ID,
		Email:           a.Email,
		Phone:           a.Phone,
		IsPhoneVerified: a.IsPhoneVerified,
		IsEmailVerified: a.IsEmailVerified,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}
