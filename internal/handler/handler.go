package handler

import (
	"context"
	"log/slog"

	"github.com/go-playground/validator/v10"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
)

type Handler struct {
	l *slog.Logger
	v *validator.Validate
	authv1.UnimplementedAuthServiceServer
}

func NewHandler(l *slog.Logger) *Handler {
	v := validator.New()
	return &Handler{
		l: l,
		v: v,
	}
}

func (h *Handler) SendOtp(context.Context, *authv1.SendOtpRequest) (*authv1.SendOtpResponse, error) {
	return &authv1.SendOtpResponse{Ok: true}, nil
}
