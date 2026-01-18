package services

import (
	"context"
	"crypto/hmac"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	"github.com/teacinema-go/auth-service/pkg/utils"
)

func (s *AuthService) GenerateOtp(ctx context.Context, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (string, error) {
	otp, err := utils.Generate6Digit()
	if err != nil {
		return otp, fmt.Errorf("failed to generate otp: %w", err)
	}

	hash := utils.GenerateHash(otp)
	key := fmt.Sprintf("otp:%s:%s", identifierType, identifier)
	return otp, s.cache.Set(ctx, key, hash, 5*time.Minute)
}

func (s *AuthService) VerifyOtp(ctx context.Context, otp string, identifier valueobject.Identifier, identifierType valueobject.IdentifierType) (bool, error) {
	key := fmt.Sprintf("otp:%s:%s", identifierType, identifier)
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, appErrors.ErrNotFound
		}
		return false, err
	}
	hash := utils.GenerateHash(otp)

	if hmac.Equal([]byte(hash), []byte(val)) {
		_ = s.cache.Delete(ctx, key)
		return true, nil
	}

	return false, nil
}
