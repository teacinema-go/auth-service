package handlers

import (
	accountv1 "github.com/teacinema-go/contracts/gen/go/account/v1"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
)

func sendErrorSendOtpResponse(errorCode authv1.SendOtpResponse_ErrorCode) (*authv1.SendOtpResponse, error) {
	return &authv1.SendOtpResponse{
		Success:   false,
		ErrorCode: errorCode,
	}, nil
}

func sendErrorVerifyOtpResponse(errorCode authv1.VerifyOtpResponse_ErrorCode) (*authv1.VerifyOtpResponse, error) {
	return &authv1.VerifyOtpResponse{
		Success:   false,
		ErrorCode: errorCode,
	}, nil
}

func sendErrorGetAccountResponse(errorCode accountv1.GetAccountResponse_ErrorCode) (*accountv1.GetAccountResponse, error) {
	return &accountv1.GetAccountResponse{
		Success:   false,
		ErrorCode: errorCode,
	}, nil
}
