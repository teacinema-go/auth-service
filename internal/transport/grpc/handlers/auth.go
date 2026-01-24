package handlers

import (
	"context"
	"errors"

	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
	"github.com/teacinema-go/core/logger"
	"github.com/teacinema-go/passport"
)

type AuthHandler struct {
	authService AuthService
	authv1.UnimplementedAuthServiceServer
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) SendOtp(ctx context.Context, req *authv1.SendOtpRequest) (*authv1.SendOtpResponse, error) {
	log := logger.With(
		"method", "SendOtp",
	)

	log.Info("send otp request received")

	identifierType, err := valueobject.NewIdentifierTypeFromProto(req.IdentifierType)
	if err != nil {
		return sendErrorSendOtpResponse(authv1.SendOtpResponse_INVALID_IDENTIFIER_TYPE)
	}

	identifier := valueobject.Identifier(req.Identifier)
	err = identifier.Validate(identifierType)
	if err != nil {
		return sendErrorSendOtpResponse(authv1.SendOtpResponse_INVALID_IDENTIFIER)
	}

	log = log.With("identifier_type", identifierType)

	exists, err := h.authService.AccountExists(ctx, identifier, identifierType)
	if err != nil {
		log.Error("failed at AccountExists()", "error", err)
		return sendErrorSendOtpResponse(authv1.SendOtpResponse_INTERNAL_ERROR)
	}

	if exists {
		return sendErrorSendOtpResponse(authv1.SendOtpResponse_ACCOUNT_ALREADY_EXISTS)
	}

	otp, err := h.authService.GenerateOtp(ctx, identifier, identifierType)
	if err != nil {
		log.Error("failed at GenerateOtp()", "error", err)
		return sendErrorSendOtpResponse(authv1.SendOtpResponse_INTERNAL_ERROR)
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

	identifierType, err := valueobject.NewIdentifierTypeFromProto(req.IdentifierType)
	if err != nil {
		return sendErrorVerifyOtpResponse(authv1.VerifyOtpResponse_INVALID_IDENTIFIER_TYPE)
	}

	identifier := valueobject.Identifier(req.Identifier)
	err = identifier.Validate(identifierType)
	if err != nil {
		return sendErrorVerifyOtpResponse(authv1.VerifyOtpResponse_INVALID_IDENTIFIER)
	}

	log = log.With("identifier_type", identifierType)

	isValid, err := h.authService.VerifyOtp(ctx, req.Otp, identifier, identifierType)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			log.Warn("invalid or expired otp")
			return sendErrorVerifyOtpResponse(authv1.VerifyOtpResponse_EXPIRED_OTP)
		}
		log.Error("failed at VerifyOtp()", "error", err)
		return sendErrorVerifyOtpResponse(authv1.VerifyOtpResponse_INTERNAL_ERROR)
	}

	if !isValid {
		log.Warn("invalid otp")
		return sendErrorVerifyOtpResponse(authv1.VerifyOtpResponse_INVALID_OTP)
	}

	log.Info("otp verified")

	res, err := h.authService.CreateAccountWithTokens(ctx, identifier, identifierType)
	if err != nil {
		if errors.Is(err, appErrors.ErrAccountAlreadyExists) {
			return sendErrorVerifyOtpResponse(authv1.VerifyOtpResponse_ACCOUNT_ALREADY_EXISTS)
		}
		log.Error("failed at CreateAccountWithTokens()", "error", err)
		return &authv1.VerifyOtpResponse{
			Success:      false,
			ErrorCode:    authv1.VerifyOtpResponse_INTERNAL_ERROR,
			ErrorMessage: "failed to create account",
		}, nil
	}

	log.Info("verification completed")

	return &authv1.VerifyOtpResponse{
		Success: true,
		Tokens: &authv1.VerifyOtpResponse_AuthTokens{
			AccessToken:      res.AccessToken,
			RefreshToken:     res.RefreshToken,
			ExpiresInSeconds: res.ExpiresIn,
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

	verified := h.authService.VerifyToken(oldToken)
	if !verified {
		log.Warn("invalid token signature")
		return &authv1.RefreshResponse{
			Success:      false,
			ErrorCode:    authv1.RefreshResponse_INVALID_REFRESH_TOKEN,
			ErrorMessage: "invalid refresh token",
		}, nil
	}

	res, err := h.authService.RotateRefreshToken(ctx, oldToken)
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
			AccessToken:      res.AccessToken,
			RefreshToken:     res.RefreshToken,
			ExpiresInSeconds: res.ExpiresIn,
		},
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	log := logger.With(
		"method", "Logout",
	)

	log.Info("logout request received")

	err := h.authService.Logout(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, appErrors.ErrInvalidRefreshToken) {
			log.Warn("refresh token not found in database")
			return &authv1.LogoutResponse{
				Success:      false,
				ErrorCode:    authv1.LogoutResponse_INVALID_REFRESH_TOKEN,
				ErrorMessage: "invalid refresh token",
			}, nil
		}
		log.Error("failed at Logout()", "error", err)
		return &authv1.LogoutResponse{
			Success:      false,
			ErrorCode:    authv1.LogoutResponse_INTERNAL_ERROR,
			ErrorMessage: "failed to logout",
		}, nil
	}

	log.Info("logout successful")

	return &authv1.LogoutResponse{
		Success: true,
	}, nil
}
