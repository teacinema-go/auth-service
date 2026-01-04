package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	teacinema "github.com/teacinema-go/auth-service/internal/database/sqlc"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	log := h.logger.With(
		"method", "SendOtp",
		"identifier_type", req.Type,
	)

	var err error
	var account teacinema.Account
	if req.Type == "phone" {
		account, err = h.queries.GetAccountByPhone(ctx, &req.Identifier)
	} else {
		account, err = h.queries.GetAccountByEmail(ctx, &req.Identifier)
	}

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Error("error getting account", "error", err)
		return &authv1.SendOtpResponse{Ok: false}, status.Error(codes.Internal, err.Error())
	}

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		uuidV7, err := uuid.NewV7()
		if err != nil {
			log.Error("failed to generate UUIDv7", "error", err)
			return &authv1.SendOtpResponse{Ok: false}, status.Error(codes.Internal, "failed to generate UUID")
		}
		accParams := teacinema.CreateAccountParams{
			ID: pgtype.UUID{
				Bytes: uuidV7,
				Valid: true,
			},
		}
		if req.Type == "phone" {
			accParams.Phone = &req.Identifier
		} else {
			accParams.Email = &req.Identifier
		}
		account, err = h.queries.CreateAccount(ctx, accParams)
		if err != nil {
			log.Error("failed to create account", "error", err)
			return &authv1.SendOtpResponse{Ok: false}, status.Error(codes.Internal, "failed to create account")
		}

	case err != nil:
		log.Error("failed to get account", "error", err)
		return &authv1.SendOtpResponse{Ok: false}, status.Error(codes.Internal, "failed to get account")
	}

	log.Info("otp sent successfully", "account", account)
	return &authv1.SendOtpResponse{Ok: true}, nil
}
