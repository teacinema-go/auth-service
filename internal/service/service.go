package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/teacinema-go/auth-service/internal/database/sqlc/accounts"
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	"github.com/teacinema-go/auth-service/pkg/enum"
	"github.com/teacinema-go/auth-service/pkg/utils"
)

type Service struct {
	accountsQ accounts.Querier
	db        *pgxpool.Pool
	rdb       *redis.Client
}

func NewService(queries *accounts.Queries, db *pgxpool.Pool, rdb *redis.Client) *Service {
	return &Service{
		accountsQ: queries,
		db:        db,
		rdb:       rdb,
	}
}

func (s *Service) GetAccount(ctx context.Context, identifier string, identifierType enum.IdentifierType) (*accounts.Account, error) {
	var err error
	var acc accounts.Account
	if identifierType == "phone" {
		acc, err = s.accountsQ.GetByPhone(ctx, &identifier)
	} else {
		acc, err = s.accountsQ.GetByEmail(ctx, &identifier)
	}

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get acc: %w", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErrors.ErrNotFound
	}

	return &acc, nil
}

func (s *Service) CreateAccount(ctx context.Context, identifier string, identifierType enum.IdentifierType) (*accounts.Account, error) {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}
	params := accounts.CreateParams{
		ID: uuidV7,
	}
	if identifierType == "phone" {
		params.Phone = &identifier
	} else {
		params.Email = &identifier
	}
	acc, err := s.accountsQ.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return &acc, nil
}

func (s *Service) VerifyAccountByIdentifierType(ctx context.Context, acc *accounts.Account, identifierType enum.IdentifierType) error {
	var err error
	if identifierType == "phone" {
		err = s.accountsQ.UpdateIsPhoneVerified(ctx, accounts.UpdateIsPhoneVerifiedParams{
			ID:              acc.ID,
			IsPhoneVerified: true,
		})
	} else {
		err = s.accountsQ.UpdateIsEmailVerified(ctx, accounts.UpdateIsEmailVerifiedParams{
			ID:              acc.ID,
			IsEmailVerified: true,
		})
	}
	return err
}

func (s *Service) GenerateOtp(ctx context.Context, identifier string, identifierType enum.IdentifierType) (string, error) {
	otp, err := utils.Generate6Digit()
	if err != nil {
		return otp, fmt.Errorf("failed to generate otp: %w", err)
	}

	hash := utils.GenerateHash(otp)
	key := fmt.Sprintf("otp:%s:%s", identifierType, identifier)
	return otp, s.rdb.Set(ctx, key, hash, 5*time.Minute).Err()
}

func (s *Service) VerifyOtp(ctx context.Context, otp string, identifier string, identifierType enum.IdentifierType) (bool, error) {
	key := fmt.Sprintf("otp:%s:%s", identifierType, identifier)
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, appErrors.ErrNotFound
		}
		return false, err
	}
	hash := utils.GenerateHash(otp)
	return hash == val, nil
}
