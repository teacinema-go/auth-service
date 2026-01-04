package handler

import (
	"context"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	teacinema "github.com/teacinema-go/auth-service/internal/database/sqlc"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
)

type Handler struct {
	logger    *slog.Logger
	validator *validator.Validate
	queries   *teacinema.Queries
	db        *pgxpool.Pool
	authv1.UnimplementedAuthServiceServer
}

func NewHandler(logger *slog.Logger, queries *teacinema.Queries, db *pgxpool.Pool) *Handler {
	v := validator.New()
	return &Handler{
		logger:    logger,
		validator: v,
		queries:   queries,
		db:        db,
	}
}

func (h *Handler) SendOtp(ctx context.Context, req *authv1.SendOtpRequest) (*authv1.SendOtpResponse, error) {
	return &authv1.SendOtpResponse{Ok: true}, nil
}
