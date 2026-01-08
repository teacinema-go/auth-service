package handler

import (
	"context"
	"errors"
	"log/slog"

	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	"github.com/teacinema-go/auth-service/internal/service"
	"github.com/teacinema-go/auth-service/pkg/enum"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		"identifier_type", req.Type,
	)
	identifierType := enum.IdentifierType(req.Type)
	if !identifierType.IsValid() {
		return &authv1.SendOtpResponse{Ok: false}, status.Error(codes.InvalidArgument, "invalid identifier type")
	}

	_, err := h.s.GetAccount(ctx, req.Identifier, identifierType)
	if err != nil && !errors.Is(err, appErrors.ErrNotFound) {
		log.Error("failed at GetAccount()", "error", err)
		return &authv1.SendOtpResponse{Ok: false}, status.Error(codes.Internal, "failed to get account")
	}

	if errors.Is(err, appErrors.ErrNotFound) {
		_, err = h.s.CreateAccount(ctx, req.Identifier, identifierType)
		if err != nil {
			log.Error("failed at CreateAccount()", "error", err)
			return &authv1.SendOtpResponse{Ok: false}, status.Error(codes.Internal, "failed to create account")
		}
	}

	otp, err := h.s.GenerateOtp(ctx, req.Identifier, identifierType)
	if err != nil {
		log.Error("failed at GenerateOtp()", "error", err)
		return &authv1.SendOtpResponse{Ok: false}, status.Error(codes.Internal, "failed to generate otp")
	}

	log.Info("otp", "otp", otp)

	return &authv1.SendOtpResponse{Ok: true}, nil
}

func (h *Handler) VerifyOtp(ctx context.Context, req *authv1.VerifyOtpRequest) (*authv1.VerifyOtpResponse, error) {
	log := h.l.With(
		"method", "VerifyOtp",
		"identifier_type", req.Type,
	)

	identifierType := enum.IdentifierType(req.Type)
	if !identifierType.IsValid() {
		return &authv1.VerifyOtpResponse{}, status.Error(codes.InvalidArgument, "invalid identifier type")
	}

	isValid, err := h.s.VerifyOtp(ctx, req.Otp, req.Identifier, identifierType)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			return &authv1.VerifyOtpResponse{}, status.Error(codes.NotFound, "invalid or expired otp")
		}
		log.Error("failed at VerifyOtp()", "error", err)
		return &authv1.VerifyOtpResponse{}, status.Error(codes.Internal, "failed to verify otp")
	}

	if !isValid {
		return &authv1.VerifyOtpResponse{}, status.Error(codes.InvalidArgument, "invalid otp")
	}

	acc, err := h.s.GetAccount(ctx, req.Identifier, identifierType)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			return &authv1.VerifyOtpResponse{}, status.Error(codes.NotFound, "account not found")
		}
		log.Error("failed at GetAccount()", "error", err)
		return &authv1.VerifyOtpResponse{}, status.Error(codes.Internal, "failed to get account")
	}

	err = h.s.VerifyAccountByIdentifierType(ctx, acc, identifierType)
	if err != nil {
		log.Error("failed at VerifyAccountByIdentifierType()", "error", err)
		return &authv1.VerifyOtpResponse{}, status.Error(codes.Internal, "failed to verify account")
	}

	return &authv1.VerifyOtpResponse{AccessToken: "123123", RefreshToken: "1231231"}, nil
}
