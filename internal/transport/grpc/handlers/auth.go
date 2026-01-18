package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
	"github.com/teacinema-go/core/logger"
	"github.com/teacinema-go/passport"
)

const (
	accessTokenTTL  = 40 * time.Minute
	refreshTokenTTL = 14 * 24 * time.Hour
)

type AuthHandler struct {
	authService AuthService
	secretKey   string
	authv1.UnimplementedAuthServiceServer
}

func NewAuthHandler(authService AuthService, secretKey string) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		secretKey:   secretKey,
	}
}

func (h *AuthHandler) SendOtp(ctx context.Context, req *authv1.SendOtpRequest) (*authv1.SendOtpResponse, error) {
	log := logger.With(
		"method", "SendOtp",
	)

	log.Info("send otp request received")
	var identifierType valueobject.IdentifierType
	identifierType, err := identifierType.FromProto(req.IdentifierType)
	if err != nil {
		log.Warn("invalid identifier type")
		return &authv1.SendOtpResponse{
			Success:      false,
			ErrorCode:    authv1.SendOtpResponse_INVALID_IDENTIFIER_TYPE,
			ErrorMessage: "invalid identifier type",
		}, nil
	}

	identifier := valueobject.Identifier(req.Identifier)
	err = identifier.Validate(identifierType)
	if err != nil {
		log.Warn("invalid identifier format")
		return &authv1.SendOtpResponse{
			Success:      false,
			ErrorCode:    authv1.SendOtpResponse_INVALID_IDENTIFIER,
			ErrorMessage: "invalid identifier format",
		}, nil
	}

	log = log.With("identifier_type", identifierType)

	_, err = h.authService.GetAccount(ctx, identifier, identifierType)
	if err != nil && !errors.Is(err, appErrors.ErrAccountNotFound) {
		log.Error("failed at GetAccount()", "error", err)
		return &authv1.SendOtpResponse{
			Success:      false,
			ErrorCode:    authv1.SendOtpResponse_INTERNAL_ERROR,
			ErrorMessage: "failed to get account",
		}, nil
	}

	if errors.Is(err, appErrors.ErrAccountNotFound) {
		err = h.authService.CreateAccount(ctx, identifier, identifierType)
		if err != nil {
			log.Error("failed at CreateAccount()", "error", err)
			return &authv1.SendOtpResponse{
				Success:      false,
				ErrorCode:    authv1.SendOtpResponse_INTERNAL_ERROR,
				ErrorMessage: "failed to create account",
			}, nil
		}
	}

	log.Info("account found or created")

	otp, err := h.authService.GenerateOtp(ctx, identifier, identifierType)
	if err != nil {
		log.Error("failed at GenerateOtp()", "error", err)
		return &authv1.SendOtpResponse{
			Success:      false,
			ErrorCode:    authv1.SendOtpResponse_INTERNAL_ERROR,
			ErrorMessage: "failed to generate otp",
		}, nil
	}

	log.Info("otp generated")
	log.Debug(otp)

	return &authv1.SendOtpResponse{
		Success: true,
		OtpInfo: &authv1.SendOtpResponse_OtpInfo{
			ExpiresInSeconds: 300,
		},
	}, nil
}

func (h *AuthHandler) VerifyOtp(ctx context.Context, req *authv1.VerifyOtpRequest) (*authv1.VerifyOtpResponse, error) {
	log := logger.With(
		"method", "VerifyOtp",
	)

	log.Info("verify otp request received")

	var identifierType valueobject.IdentifierType
	identifierType, err := identifierType.FromProto(req.IdentifierType)
	if err != nil {
		log.Warn("invalid identifier type")
		return &authv1.VerifyOtpResponse{
			Success:      false,
			ErrorCode:    authv1.VerifyOtpResponse_INVALID_IDENTIFIER_TYPE,
			ErrorMessage: "invalid identifier type",
		}, nil
	}

	identifier := valueobject.Identifier(req.Identifier)
	err = identifier.Validate(identifierType)
	if err != nil {
		log.Warn("invalid identifier format")
		return &authv1.VerifyOtpResponse{
			Success:      false,
			ErrorCode:    authv1.VerifyOtpResponse_INVALID_IDENTIFIER,
			ErrorMessage: "invalid identifier format",
		}, nil
	}

	log = log.With("identifier_type", identifierType)

	isValid, err := h.authService.VerifyOtp(ctx, req.Otp, identifier, identifierType)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			log.Warn("invalid or expired otp")
			return &authv1.VerifyOtpResponse{
				Success:      false,
				ErrorCode:    authv1.VerifyOtpResponse_EXPIRED_OTP,
				ErrorMessage: "invalid or expired otp",
			}, nil
		}
		log.Error("failed at VerifyOtp()", "error", err)
		return &authv1.VerifyOtpResponse{
			Success:      false,
			ErrorCode:    authv1.VerifyOtpResponse_INTERNAL_ERROR,
			ErrorMessage: "failed to verify otp",
		}, nil
	}

	if !isValid {
		log.Warn("invalid otp")
		return &authv1.VerifyOtpResponse{
			Success:      false,
			ErrorCode:    authv1.VerifyOtpResponse_INVALID_OTP,
			ErrorMessage: "invalid otp",
		}, nil
	}

	log.Info("otp verified")

	acc, err := h.authService.GetAccount(ctx, identifier, identifierType)
	if err != nil {
		if errors.Is(err, appErrors.ErrAccountNotFound) {
			log.Warn("account not found")
			return &authv1.VerifyOtpResponse{
				Success:      false,
				ErrorCode:    authv1.VerifyOtpResponse_ACCOUNT_NOT_FOUND,
				ErrorMessage: "account not found",
			}, nil
		}
		log.Error("failed at GetAccount()", "error", err)
		return &authv1.VerifyOtpResponse{
			Success:      false,
			ErrorCode:    authv1.VerifyOtpResponse_INTERNAL_ERROR,
			ErrorMessage: "failed to get account",
		}, nil
	}

	log = logger.With(
		"accountID", acc.ID.String(),
	)

	log.Info("account found")

	accessToken := passport.GenerateToken(h.secretKey, acc.ID.String(), accessTokenTTL)
	refreshToken := passport.GenerateToken(h.secretKey, acc.ID.String(), refreshTokenTTL)

	err = h.authService.CompleteAccountVerification(ctx, acc, refreshToken)
	if err != nil {
		log.Error("failed at CompleteVerification()", "error", err)
		return &authv1.VerifyOtpResponse{
			Success:      false,
			ErrorCode:    authv1.VerifyOtpResponse_INTERNAL_ERROR,
			ErrorMessage: "failed to complete verification",
		}, nil
	}

	log.Info("verification completed")

	return &authv1.VerifyOtpResponse{
		Success: true,
		Tokens: &authv1.VerifyOtpResponse_AuthTokens{
			AccessToken:      accessToken.Val,
			RefreshToken:     refreshToken.Val,
			ExpiresInSeconds: int32(accessTokenTTL.Seconds()),
		},
	}, nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	log := logger.With(
		"method", "Refresh",
	)

	log.Info("refresh token request received")

	oldToken, err := passport.ParseToken(req.RefreshToken)
	if err != nil {
		log.Warn("failed at ParseToken()", "error", err)
		errorCode := authv1.RefreshResponse_INTERNAL_ERROR
		errorMessage := "failed to parse refresh token"
		switch {
		case errors.Is(err, passport.ErrInvalidToken):
			errorCode = authv1.RefreshResponse_INVALID_REFRESH_TOKEN
			errorMessage = "invalid refresh token"
		case errors.Is(err, passport.ErrExpiredToken):
			errorCode = authv1.RefreshResponse_EXPIRED_REFRESH_TOKEN
			errorMessage = "expired refresh token"
		}

		return &authv1.RefreshResponse{
			Success:      false,
			ErrorCode:    errorCode,
			ErrorMessage: errorMessage,
		}, nil
	}

	verified := passport.VerifyToken(oldToken, h.secretKey)
	if !verified {
		log.Warn("invalid token signature")
		return &authv1.RefreshResponse{
			Success:      false,
			ErrorCode:    authv1.RefreshResponse_INVALID_REFRESH_TOKEN,
			ErrorMessage: "invalid refresh token",
		}, nil
	}

	newAccessToken := passport.GenerateToken(h.secretKey, oldToken.UserID, accessTokenTTL)
	newRefreshToken := passport.GenerateToken(h.secretKey, oldToken.UserID, refreshTokenTTL)

	err = h.authService.RotateRefreshToken(ctx, req.RefreshToken, newRefreshToken)
	if err != nil {
		if errors.Is(err, appErrors.ErrInvalidRefreshToken) {
			log.Warn("refresh token not found in database")
			return &authv1.RefreshResponse{
				Success:      false,
				ErrorCode:    authv1.RefreshResponse_INVALID_REFRESH_TOKEN,
				ErrorMessage: "invalid refresh token",
			}, nil
		}
		log.Error("failed at RotateRefreshToken()", "error", err)
		return &authv1.RefreshResponse{
			Success:      false,
			ErrorCode:    authv1.RefreshResponse_INTERNAL_ERROR,
			ErrorMessage: "failed to refresh token",
		}, nil
	}

	log.Info("tokens refreshed successfully")

	return &authv1.RefreshResponse{
		Success: true,
		Tokens: &authv1.RefreshResponse_AuthTokens{
			AccessToken:      newAccessToken.Val,
			RefreshToken:     newRefreshToken.Val,
			ExpiresInSeconds: int32(accessTokenTTL.Seconds()),
		},
	}, nil
}
