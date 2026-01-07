package handler

import (
	"context"
	"log/slog"

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

	_, err := h.s.GetOrCreateAccount(ctx, req.Identifier, identifierType)
	if err != nil {
		log.Error("failed at GetOrCreateAccount()", "error", err)
		return &authv1.SendOtpResponse{Ok: false}, status.Error(codes.Internal, "failed to get or create account")
	}

	code, err := h.s.GenerateCode(ctx, req.Identifier, identifierType)
	if err != nil {
		log.Error("failed at GenerateCode()", "error", err)
		return &authv1.SendOtpResponse{Ok: false}, status.Error(codes.Internal, "failed to generate code")
	}

	log.Info("code", "code", code)

	return &authv1.SendOtpResponse{Ok: true}, nil
}
