package grpc_server

import (
	"context"

	pb "cp_service/api/proto/pb"
	"cp_service/internal/adapters/helpers"
	"cp_service/internal/core/services"

	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthnHandler struct {
	pb.UnimplementedAuthnServiceServer
	authnService *services.AuthnService
}

func NewAuthnHandler(authnService *services.AuthnService) *AuthnHandler {
	return &AuthnHandler{authnService: authnService}
}

func (h *AuthnHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	accessToken, refreshToken, user, err := h.authnService.Login(ctx, req.Email, req.Password, req.TenantIdentifier)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         dbUserToProto(&user),
	}, nil
}

// LoginWithSAML - COMMENTED OUT FOR NOW
// func (h *AuthnHandler) LoginWithSAML(ctx context.Context, req *pb.LoginWithSAMLRequest) (*pb.LoginResponse, error) {
// 	accessToken, refreshToken, user, err := h.authnService.LoginWithSAML(ctx, req.SamlResponse, req.TenantIdentifier)
// 	if err != nil {
// 		return nil, helpers.ToGRPCError(err)
// 	}
//
// 	return &pb.LoginResponse{
// 		AccessToken:  accessToken,
// 		RefreshToken: refreshToken,
// 		User:         dbUserToProto(&user),
// 	}, nil
// }

// GetSAMLAuthURL - COMMENTED OUT FOR NOW
// func (h *AuthnHandler) GetSAMLAuthURL(ctx context.Context, req *pb.GetSAMLAuthURLRequest) (*pb.GetSAMLAuthURLResponse, error) {
// 	authURL, err := h.authnService.GetSAMLAuthURL(ctx, req.RelayState)
// 	if err != nil {
// 		return nil, helpers.ToGRPCError(err)
// 	}
//
// 	return &pb.GetSAMLAuthURLResponse{AuthUrl: authURL}, nil
// }

func (h *AuthnHandler) SetInitialPassword(ctx context.Context, req *pb.SetInitialPasswordRequest) (*emptypb.Empty, error) {
	if err := h.authnService.SetInitialPassword(ctx, req.SetupToken, req.NewPassword); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *AuthnHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	accessToken, err := h.authnService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.RefreshTokenResponse{AccessToken: accessToken}, nil
}

func (h *AuthnHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := h.authnService.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.ValidateTokenResponse{
		Valid:    true,
		UserId:   claims.UserID,
		TenantId: claims.TenantID,
		Email:    claims.Email,
	}, nil
}

func (h *AuthnHandler) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*emptypb.Empty, error) {
	if err := h.authnService.ForgotPassword(ctx, req.Email, req.TenantIdentifier); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *AuthnHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*emptypb.Empty, error) {
	if err := h.authnService.ResetPassword(ctx, req.ResetToken, req.NewPassword); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

// RegisterInvitedUser allows invited users to complete registration and get logged in automatically
// Returns LoginResponse with access/refresh tokens (admin already set their role during invitation)
func (h *AuthnHandler) RegisterInvitedUser(ctx context.Context, req *pb.RegisterInvitedUserRequest) (*pb.LoginResponse, error) {
	accessToken, refreshToken, user, err := h.authnService.RegisterInvitedUser(ctx, req.InvitationToken, req.FullName, req.Password)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         dbUserToProto(&user),
	}, nil
}

func (h *AuthnHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	if err := h.authnService.Logout(ctx, req.UserId); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *AuthnHandler) ConfirmPassword(ctx context.Context, req *pb.ConfirmPasswordRequest) (*pb.ConfirmPasswordResponse, error) {
	err := h.authnService.ConfirmPassword(ctx, req.UserId, req.Password)
	if err != nil {
		return &pb.ConfirmPasswordResponse{Valid: false}, nil
	}

	return &pb.ConfirmPasswordResponse{Valid: true}, nil
}

func (h *AuthnHandler) GeneratePasswordSetupToken(ctx context.Context, req *pb.GeneratePasswordSetupTokenRequest) (*pb.GeneratePasswordSetupTokenResponse, error) {
	token, err := h.authnService.GeneratePasswordSetupToken(ctx, req.UserId, req.TenantId, req.Email)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.GeneratePasswordSetupTokenResponse{Token: token}, nil
}
