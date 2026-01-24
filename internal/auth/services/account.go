package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/teacinema-go/auth-service/internal/auth/dto"
	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	"github.com/teacinema-go/auth-service/pkg/utils"
	"github.com/teacinema-go/passport"
)

const (
	accessTokenTTL  = 40 * time.Minute
	refreshTokenTTL = 14 * 24 * time.Hour
)

func (s *AuthService) AccountExists(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (bool, error) {
	if identifierType == valueobject.IdentifierTypePhone {
		return s.accountRepo.AccountExistsByPhone(ctx, identifier)
	}

	return s.accountRepo.AccountExistsByEmail(ctx, identifier)
}

func (s *AuthService) CreateAccountWithTokens(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (dto.Tokens, error) {
	res, err := s.txManager.WithTransaction(ctx, func(repos TxRepositories) (any, error) {
		// Create an account
		accountID, err := valueobject.NewID()
		if err != nil {
			return nil, err
		}
		params := dto.CreateAccountParams{
			ID:   accountID,
			Role: valueobject.RoleUser,
		}
		strIdentifier := string(identifier)
		if identifierType == valueobject.IdentifierTypePhone {
			params.Phone = &strIdentifier
		} else {
			params.Email = &strIdentifier
		}
		err = s.accountRepo.CreateAccount(ctx, params)
		if err != nil {
			if errors.Is(err, appErrors.ErrAccountAlreadyExists) {
				return nil, appErrors.ErrAccountAlreadyExists
			}
			return nil, fmt.Errorf("failed to create account: %w", err)
		}

		// Generate tokens
		tokenID, err := valueobject.NewID()
		if err != nil {
			return nil, err
		}

		refreshToken := passport.GenerateToken(s.secretKey, accountID.ToUUID().String(), refreshTokenTTL)
		err = repos.RefreshToken().CreateRefreshToken(ctx, dto.CreateRefreshTokenParams{
			ID:        tokenID.ToUUID(),
			AccountID: accountID.ToUUID(),
			TokenHash: utils.GenerateHash(refreshToken.Val),
			ExpiresAt: time.Unix(refreshToken.Exp, 0),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create refresh token: %w", err)
		}
		accessToken := passport.GenerateToken(s.secretKey, accountID.ToUUID().String(), accessTokenTTL)

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
