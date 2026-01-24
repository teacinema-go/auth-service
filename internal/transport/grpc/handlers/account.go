package handlers

import (
	"context"
	"errors"

	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
	appErrors "github.com/teacinema-go/auth-service/internal/errors"
	accountv1 "github.com/teacinema-go/contracts/gen/go/account/v1"
	"github.com/teacinema-go/core/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AccountHandler struct {
	authService AuthService
	accountv1.UnimplementedAccountServiceServer
}

func NewAccountHandler(authService AuthService) *AccountHandler {
	return &AccountHandler{
		authService: authService,
	}
}

func (h *AccountHandler) GetAccount(ctx context.Context, req *accountv1.GetAccountRequest) (*accountv1.GetAccountResponse, error) {
	log := logger.With(
		"method", "GetAccount",
	)

	log.Info("get account request received")

	ID, err := valueobject.NewIDFromString(req.GetId())
	if err != nil {
		return sendErrorGetAccountResponse(accountv1.GetAccountResponse_INVALID_ID)
	}

	acc, err := h.authService.GetAccount(ctx, ID)
	if err != nil {
		if errors.Is(err, appErrors.ErrAccountNotFound) {
			return sendErrorGetAccountResponse(accountv1.GetAccountResponse_ACCOUNT_NOT_FOUND)
		}
		return sendErrorGetAccountResponse(accountv1.GetAccountResponse_INTERNAL_ERROR)
	}

	return &accountv1.GetAccountResponse{
		Success: true,
		Account: &accountv1.Account{
			Id:        acc.ID.ToString(),
			Phone:     *acc.Phone,
			Email:     *acc.Email,
			Role:      acc.Role.ToProto(),
			CreatedAt: timestamppb.New(acc.CreatedAt),
			UpdatedAt: timestamppb.New(acc.UpdatedAt),
		},
	}, nil
}
