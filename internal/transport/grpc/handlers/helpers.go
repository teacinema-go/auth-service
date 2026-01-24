package handlers

import (
	"github.com/teacinema-go/auth-service/internal/auth/valueobject"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
)

func getIdentifierTypeFromProto(protoIdentifierType authv1.IdentifierType) (valueobject.IdentifierType, error) {
	var identifierType valueobject.IdentifierType
	return identifierType.FromProto(protoIdentifierType)
}

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
