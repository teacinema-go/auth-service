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
	teacinema "github.com/teacinema-go/auth-service/internal/database/sqlc"
	"github.com/teacinema-go/auth-service/pkg/enum"
	"github.com/teacinema-go/auth-service/pkg/utils"
)

type Service struct {
	queries *teacinema.Queries
	db      *pgxpool.Pool
	rdb     *redis.Client
}

func NewService(queries *teacinema.Queries, db *pgxpool.Pool, rdb *redis.Client) *Service {
	return &Service{
		queries: queries,
		db:      db,
		rdb:     rdb,
	}
}

func (s *Service) GetOrCreateAccount(ctx context.Context, identifier string, identifierType enum.IdentifierType) (*teacinema.Account, error) {
	var err error
	var acc teacinema.Account
	if identifierType == "phone" {
		acc, err = s.queries.GetAccountByPhone(ctx, &identifier)
	} else {
		acc, err = s.queries.GetAccountByEmail(ctx, &identifier)
	}

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get acc: %w", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		uuidV7, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("failed to generate UUID: %w", err)
		}
		accParams := teacinema.CreateAccountParams{
			ID: uuidV7,
		}
		if identifierType == "phone" {
			accParams.Phone = &identifier
		} else {
			accParams.Email = &identifier
		}
		acc, err = s.queries.CreateAccount(ctx, accParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create acc: %w", err)
		}
	}

	return &acc, nil
}

func (s *Service) GenerateCode(ctx context.Context, identifier string, identifierType enum.IdentifierType) (int64, error) {
	code, err := utils.GenerateCode()
	if err != nil {
		return code, fmt.Errorf("failed to generate code: %w", err)
	}

	hash := utils.GenerateHashForCode(code)
	key := fmt.Sprintf("otp:%s:%s", identifierType, identifier)
	s.rdb.Set(ctx, key, hash, 5*time.Minute)
	return code, nil
}
