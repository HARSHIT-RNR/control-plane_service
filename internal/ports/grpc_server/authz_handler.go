package grpc_server

import (
	"context"

	pb "cp_service/api/proto/pb"
	"cp_service/internal/adapters/helpers"
	"cp_service/internal/core/services"
)

type AuthzHandler struct {
	pb.UnimplementedAuthzServiceServer
	authzService *services.AuthzService
}

func NewAuthzHandler(authzService *services.AuthzService) *AuthzHandler {
	return &AuthzHandler{authzService: authzService}
}

func (h *AuthzHandler) CheckAccess(ctx context.Context, req *pb.CheckAccessRequest) (*pb.CheckAccessResponse, error) {
	allowed, reason, err := h.authzService.CheckAccess(ctx, req.AccessToken, req.Action, req.Resource)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.CheckAccessResponse{
		Allowed: allowed,
		Reason:  reason,
	}, nil
}

// GetUserPermissions and HasPermission are commented out as they're not in the proto definition
// Uncomment when/if these are added to the proto

// func (h *AuthzHandler) GetUserPermissions(ctx context.Context, req *pb.GetUserPermissionsRequest) (*pb.GetUserPermissionsResponse, error) {
// 	permissions, err := h.authzService.GetUserPermissions(ctx, req.UserId)
// 	if err != nil {
// 		return nil, helpers.ToGRPCError(err)
// 	}

// 	return &pb.GetUserPermissionsResponse{Permissions: permissions}, nil
// }

// func (h *AuthzHandler) HasPermission(ctx context.Context, req *pb.HasPermissionRequest) (*pb.HasPermissionResponse, error) {
// 	has, err := h.authzService.HasPermission(ctx, req.UserId, req.Permission)
// 	if err != nil {
// 		return nil, helpers.ToGRPCError(err)
// 	}

// 	return &pb.HasPermissionResponse{HasPermission: has}, nil
// }
