package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/teacinema-go/auth-service/internal/auth/dto"
	"github.com/teacinema-go/auth-service/internal/auth/entity"
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	"github.com/teacinema-go/auth-service/pkg/utils"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
)

type Repository interface {
	Create(ctx context.Context, arg dto.CreateAccountParams) (*entity.Account, error)
	GetByEmail(ctx context.Context, email *string) (*entity.Account, error)
	GetByPhone(ctx context.Context, phone *string) (*entity.Account, error)
	UpdateIsEmailVerified(ctx context.Context, arg dto.UpdateAccountIsEmailVerifiedParams) error
	UpdateIsPhoneVerified(ctx context.Context, arg dto.UpdateAccountIsPhoneVerifiedParams) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
}

type Service struct {
	repo  Repository
	cache Cache
}

func NewService(repo Repository, cache Cache) *Service {
	return &Service{
		repo:  repo,
		cache: cache,
	}
}

func (s *Service) GetAccount(ctx context.Context, identifier string, identifierType authv1.IdentifierType) (*entity.Account, error) {
	var err error
	var acc *entity.Account
	if identifierType == authv1.IdentifierType_PHONE {
		acc, err = s.repo.GetByPhone(ctx, &identifier)
	} else {
		acc, err = s.repo.GetByEmail(ctx, &identifier)
	}

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get acc: %w", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErrors.ErrNotFound
	}

	return acc, nil
}

func (s *Service) CreateAccount(ctx context.Context, identifier string, identifierType authv1.IdentifierType) (*entity.Account, error) {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}
	params := dto.CreateAccountParams{
		ID: uuidV7,
	}
	if identifierType == authv1.IdentifierType_PHONE {
		params.Phone = &identifier
	} else {
		params.Email = &identifier
	}
	acc, err := s.repo.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return acc, nil
}

func (s *Service) VerifyAccountByIdentifierType(ctx context.Context, acc *entity.Account, identifierType authv1.IdentifierType) error {
	var err error
	if identifierType == authv1.IdentifierType_PHONE {
		err = s.repo.UpdateIsPhoneVerified(ctx, dto.UpdateAccountIsPhoneVerifiedParams{
			ID:              acc.ID,
			IsPhoneVerified: true,
		})
	} else {
		err = s.repo.UpdateIsEmailVerified(ctx, dto.UpdateAccountIsEmailVerifiedParams{
			ID:              acc.ID,
			IsEmailVerified: true,
		})
	}
	return err
}

func (s *Service) GenerateOtp(ctx context.Context, identifier string, identifierType authv1.IdentifierType) (string, error) {
	otp, err := utils.Generate6Digit()
	if err != nil {
		return otp, fmt.Errorf("failed to generate otp: %w", err)
	}

	hash := utils.GenerateHash(otp)
	key := fmt.Sprintf("otp:%s:%s", identifierType, identifier)
	return otp, s.cache.Set(ctx, key, hash, 5*time.Minute)
}

func (s *Service) VerifyOtp(ctx context.Context, otp string, identifier string, identifierType authv1.IdentifierType) (bool, error) {
	key := fmt.Sprintf("otp:%s:%s", identifierType, identifier)
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, appErrors.ErrNotFound
		}
		return false, err
	}
	hash := utils.GenerateHash(otp)
	return hash == val, nil
}
