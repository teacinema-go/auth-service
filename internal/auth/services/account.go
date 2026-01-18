package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/teacinema-go/auth-service/internal/auth/dto"
	"github.com/teacinema-go/auth-service/internal/auth/entities"
	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	"github.com/teacinema-go/auth-service/pkg/utils"
	"github.com/teacinema-go/passport"
)

func (s *AuthService) GetAccount(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (*entities.Account, error) {
	var err error
	var acc *entities.Account
	if identifierType == valueobject.IdentifierTypePhone {
		acc, err = s.accountRepo.GetAccountByPhone(ctx, identifier)
	} else {
		acc, err = s.accountRepo.GetAccountByEmail(ctx, identifier)
	}

	if err != nil {
		if errors.Is(err, appErrors.ErrAccountNotFound) {
			return nil, appErrors.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return acc, nil
}

func (s *AuthService) CreateAccount(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) error {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("failed to generate UUID: %w", err)
	}
	params := dto.CreateAccountParams{
		ID: uuidV7,
	}
	strIdentifier := string(identifier)
	if identifierType == valueobject.IdentifierTypePhone {
		params.Phone = &strIdentifier
	} else {
		params.Email = &strIdentifier
	}
	err = s.accountRepo.CreateAccount(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

func (s *AuthService) CompleteAccountVerification(ctx context.Context, acc *entities.Account, refreshToken *passport.Token) error {
	return s.txManager.WithTransaction(ctx, func(repos TxRepositories) error {
		var err error
		if acc.Phone != nil {
			err = repos.Account().UpdateAccountIsPhoneVerified(ctx, dto.UpdateAccountIsPhoneVerifiedParams{
				ID:              acc.ID,
				IsPhoneVerified: true,
			})
		} else {
			err = repos.Account().UpdateAccountIsEmailVerified(ctx, dto.UpdateAccountIsEmailVerifiedParams{
				ID:              acc.ID,
				IsEmailVerified: true,
			})
		}
		if err != nil {
			return fmt.Errorf("failed to verify account: %w", err)
		}

		tokenID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate token ID: %w", err)
		}

		err = repos.RefreshToken().CreateRefreshToken(ctx, dto.CreateRefreshTokenParams{
			ID:        tokenID,
			AccountID: acc.ID,
			TokenHash: utils.GenerateHash(refreshToken.Val),
			ExpiresAt: time.Unix(refreshToken.Exp, 0),
		})
		if err != nil {
			return fmt.Errorf("failed to create refresh token: %w", err)
		}

		return nil
	})
}

func (s *AuthService) RotateRefreshToken(ctx context.Context, oldTokenValue string, newToken *passport.Token) error {
	return s.txManager.WithTransaction(ctx, func(repos TxRepositories) error {
		oldHash := utils.GenerateHash(oldTokenValue)

		_, err := repos.RefreshToken().GetRefreshTokenByHash(ctx, oldHash)
		if err != nil {
			if errors.Is(err, appErrors.ErrRefreshTokenNotFound) {
				return appErrors.ErrInvalidRefreshToken
			}
			return fmt.Errorf("failed to get refresh token: %w", err)
		}

		err = repos.RefreshToken().DeleteRefreshTokenByHash(ctx, oldHash)
		if err != nil {
			return fmt.Errorf("failed to delete old refresh token: %w", err)
		}

		tokenID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate token ID: %w", err)
		}

		accountID, err := uuid.Parse(newToken.UserID)
		if err != nil {
			return fmt.Errorf("failed to parse account ID: %w", err)
		}

		err = repos.RefreshToken().CreateRefreshToken(ctx, dto.CreateRefreshTokenParams{
			ID:        tokenID,
			AccountID: accountID,
			TokenHash: utils.GenerateHash(newToken.Val),
			ExpiresAt: time.Unix(newToken.Exp, 0),
		})
		if err != nil {
			return fmt.Errorf("failed to create new refresh token: %w", err)
		}

		return nil
	})
}
