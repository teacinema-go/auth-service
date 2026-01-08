package handler

import (
	"context"
	"errors"
	"log/slog"

	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	"github.com/teacinema-go/auth-service/internal/service"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
)

type Handler struct {
	l *slog.Logger
	s *service.Service
	authv1.UnimplementedAuthServiceServer
}

func NewHandler(l *slog.Logger, s *service.Service) *Handler {
	return &Handler{
		l: l,
		s: s,
	}
}

func (h *Handler) SendOtp(ctx context.Context, req *authv1.SendOtpRequest) (*authv1.SendOtpResponse, error) {
	log := h.l.With(
		"method", "SendOtp",
	)

	log.Info("send otp request received")

	identifierType := req.GetIdentifierType()
	if identifierType == authv1.IdentifierType_IDENTIFIER_TYPE_UNSPECIFIED {
		log.Warn("invalid identifier type")
		return &authv1.SendOtpResponse{
			Success:      false,
			ErrorCode:    authv1.SendOtpResponse_INVALID_IDENTIFIER_TYPE,
			ErrorMessage: "invalid identifier type",
		}, nil
	}

	log = log.With("identifier_type", identifierType)

	_, err := h.s.GetAccount(ctx, req.Identifier, identifierType)
	if err != nil && !errors.Is(err, appErrors.ErrNotFound) {
		log.Error("failed at GetAccount()", "error", err)
		return &authv1.SendOtpResponse{
			Success:      false,
			ErrorCode:    authv1.SendOtpResponse_INTERNAL_ERROR,
			ErrorMessage: "failed to get account",
		}, nil
	}

	if errors.Is(err, appErrors.ErrNotFound) {
		_, err = h.s.CreateAccount(ctx, req.Identifier, identifierType)
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

	otp, err := h.s.GenerateOtp(ctx, req.Identifier, identifierType)
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

func (h *Handler) VerifyOtp(ctx context.Context, req *authv1.VerifyOtpRequest) (*authv1.VerifyOtpResponse, error) {
	log := h.l.With(
		"method", "VerifyOtp",
	)

	log.Info("verify otp request received")

	identifierType := req.GetIdentifierType()
	if identifierType == authv1.IdentifierType_IDENTIFIER_TYPE_UNSPECIFIED {
		log.Warn("invalid identifier type")
		return &authv1.VerifyOtpResponse{
			Success:      false,
			ErrorCode:    authv1.VerifyOtpResponse_INVALID_IDENTIFIER_TYPE,
			ErrorMessage: "invalid identifier type",
		}, nil
	}

	log = log.With("identifier_type", identifierType)

	isValid, err := h.s.VerifyOtp(ctx, req.Otp, req.Identifier, identifierType)
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

	acc, err := h.s.GetAccount(ctx, req.Identifier, identifierType)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
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

	log.Info("account found")

	err = h.s.VerifyAccountByIdentifierType(ctx, acc, identifierType)
	if err != nil {
		log.Error("failed at VerifyAccountByIdentifierType()", "error", err)
		return &authv1.VerifyOtpResponse{
			Success:      false,
			ErrorCode:    authv1.VerifyOtpResponse_INTERNAL_ERROR,
			ErrorMessage: "failed to verify account",
		}, nil
	}

	log.Info("account verified")

	return &authv1.VerifyOtpResponse{
		Success: true,
		Tokens: &authv1.VerifyOtpResponse_AuthTokens{
			AccessToken:      "123",
			RefreshToken:     "1234",
			ExpiresInSeconds: 2,
		},
	}, nil
}
