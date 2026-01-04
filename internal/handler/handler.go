package handler

import (
	"context"
	"log/slog"

	"github.com/go-playground/validator/v10"
	teacinema "github.com/teacinema-go/auth-service/internal/database/sqlc"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
)

type Handler struct {
	l  *slog.Logger
	v  *validator.Validate
	db *teacinema.Queries
	authv1.UnimplementedAuthServiceServer
}

func NewHandler(l *slog.Logger, db *teacinema.Queries) *Handler {
	v := validator.New()
	return &Handler{
		l:  l,
		v:  v,
		db: db,
	}
}

func (h *Handler) SendOtp(ctx context.Context, req *authv1.SendOtpRequest) (*authv1.SendOtpResponse, error) {
	return &authv1.SendOtpResponse{Ok: true}, nil
}
