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

const (
	accessTokenTTL  = 40 * time.Minute
	refreshTokenTTL = 14 * 24 * time.Hour
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

func (s *AuthService) CompleteAccountVerification(ctx context.Context, acc *entities.Account) (dto.Tokens, error) {
	res, err := s.txManager.WithTransaction(ctx, func(repos TxRepositories) (any, error) {
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
			return nil, fmt.Errorf("failed to verify account: %w", err)
		}

		tokenID, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("failed to generate token ID: %w", err)
		}

		refreshToken := passport.GenerateToken(s.secretKey, acc.ID.String(), refreshTokenTTL)

		err = repos.RefreshToken().CreateRefreshToken(ctx, dto.CreateRefreshTokenParams{
			ID:        tokenID,
			AccountID: acc.ID,
			TokenHash: utils.GenerateHash(refreshToken.Val),
			ExpiresAt: time.Unix(refreshToken.Exp, 0),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create refresh token: %w", err)
		}
		accessToken := passport.GenerateToken(s.secretKey, acc.ID.String(), accessTokenTTL)

		return dto.Tokens{
			AccessToken:  accessToken.Val,
			RefreshToken: refreshToken.Val,
			ExpiresIn:    int32(accessTokenTTL.Seconds()),
		}, nil
	})
	if err != nil {
		return dto.Tokens{}, err
	}

	return res.(dto.Tokens), nil
}

func (s *AuthService) VerifyToken(token *passport.Token) bool {
	return token.VerifyToken(s.secretKey)
}

func (s *AuthService) RotateRefreshToken(ctx context.Context, oldToken *passport.Token) (dto.Tokens, error) {
	res, err := s.txManager.WithTransaction(ctx, func(repos TxRepositories) (any, error) {
		oldHash := utils.GenerateHash(oldToken.Val)

		rowsAffected, err := repos.RefreshToken().DeleteRefreshTokenByHash(ctx, oldHash)
		if err != nil {
			return nil, fmt.Errorf("failed to delete old refresh token: %w", err)
		}

		if rowsAffected == 0 {
			return nil, appErrors.ErrInvalidRefreshToken
		}

		tokenID, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("failed to generate token ID: %w", err)
		}

		newRefreshToken := passport.GenerateToken(s.secretKey, oldToken.UserID, refreshTokenTTL)

		accountID, err := uuid.Parse(newRefreshToken.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse account ID: %w", err)
		}

		err = repos.RefreshToken().CreateRefreshToken(ctx, dto.CreateRefreshTokenParams{
			ID:        tokenID,
			AccountID: accountID,
			TokenHash: utils.GenerateHash(newRefreshToken.Val),
			ExpiresAt: time.Unix(newRefreshToken.Exp, 0),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create new refresh token: %w", err)
		}
		newAccessToken := passport.GenerateToken(s.secretKey, oldToken.UserID, accessTokenTTL)

		return &dto.Tokens{
			AccessToken:  newAccessToken.Val,
			RefreshToken: newRefreshToken.Val,
			ExpiresIn:    int32(accessTokenTTL.Seconds()),
		}, nil
	})

	if err != nil {
		return dto.Tokens{}, err
	}

	return res.(dto.Tokens), nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := utils.GenerateHash(refreshToken)

	rowsAffected, err := s.refreshTokenRepo.DeleteRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	if rowsAffected == 0 {
		return appErrors.ErrInvalidRefreshToken
	}

	return nil
}
