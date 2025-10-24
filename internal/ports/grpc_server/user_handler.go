package grpc_server

import (
	"context"

	pb "cp_service/api/proto/pb"
	"cp_service/internal/adapters/database/db"
	"cp_service/internal/adapters/helpers"
	"cp_service/internal/core/services"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// CreateInitialAdmin creates the first admin user for a new tenant
func (h *UserHandler) CreateInitialAdmin(ctx context.Context, req *pb.CreateInitialAdminRequest) (*pb.User, error) {
	user, err := h.userService.CreateInitialAdmin(ctx, req.TenantId, req.AdminEmail, req.AdminFullName)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbUserToProto(&user), nil
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	// Build params
	params := db.CreateUserParams{
		FullName: req.FullName,
		Email:    req.Email,
		TenantID: helpers.StringToPgUUID(req.TenantId),
		Status:   db.UserStatusACTIVE,
	}

	if req.EmployeeId != "" {
		params.EmployeeID = helpers.StringToPgText(req.EmployeeId)
	}
	if req.DepartmentId != "" {
		params.DepartmentID = helpers.StringToPgUUID(req.DepartmentId)
	}
	if req.DesignationId != "" {
		params.DesignationID = helpers.StringToPgUUID(req.DesignationId)
	}
	if req.JobTitle != "" {
		params.JobTitle = helpers.StringToPgText(req.JobTitle)
	}
	if req.PhoneNumber != "" {
		params.PhoneNumber = helpers.StringToPgText(req.PhoneNumber)
	}

	user, err := h.userService.CreateUser(ctx, params)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbUserToProto(&user), nil
}

func (h *UserHandler) InviteUser(ctx context.Context, req *pb.InviteUserRequest) (*pb.User, error) {
	user, err := h.userService.InviteUser(ctx, req.TenantId, req.Email, req.FullName, req.RoleIds)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbUserToProto(&user), nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	user, err := h.userService.GetUser(ctx, req.Id)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbUserToProto(&user), nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 50
	}

	users, err := h.userService.ListUsers(ctx, req.TenantId, limit, req.Offset)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = dbUserToProto(&user)
	}

	return &pb.ListUsersResponse{Users: pbUsers}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	params := db.UpdateUserParams{
		ID:       helpers.StringToPgUUID(req.Id),
		FullName: req.FullName,
	}

	if req.PhoneNumber != "" {
		params.PhoneNumber = helpers.StringToPgText(req.PhoneNumber)
	}
	if req.JobTitle != "" {
		params.JobTitle = helpers.StringToPgText(req.JobTitle)
	}
	if req.DepartmentId != "" {
		params.DepartmentID = helpers.StringToPgUUID(req.DepartmentId)
	}
	if req.DesignationId != "" {
		params.DesignationID = helpers.StringToPgUUID(req.DesignationId)
	}

	user, err := h.userService.UpdateUser(ctx, params)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbUserToProto(&user), nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := h.userService.DeleteUser(ctx, req.Id); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}

func (h *UserHandler) AssignRoleToUser(ctx context.Context, req *pb.AssignRoleToUserRequest) (*pb.AssignRoleToUserResponse, error) {
	if err := h.userService.AssignRoleToUser(ctx, req.UserId, req.TenantId, req.RoleId); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.AssignRoleToUserResponse{Success: true}, nil
}

func (h *UserHandler) RevokeRoleFromUser(ctx context.Context, req *pb.RevokeRoleFromUserRequest) (*pb.RevokeRoleFromUserResponse, error) {
	if err := h.userService.RevokeRoleFromUser(ctx, req.UserId, req.TenantId, req.RoleId); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &pb.RevokeRoleFromUserResponse{Success: true}, nil
}

func (h *UserHandler) ListUserRoles(ctx context.Context, req *pb.ListUserRolesRequest) (*pb.ListUserRolesResponse, error) {
	roles, err := h.userService.ListUserRoles(ctx, req.UserId)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	pbRoles := make([]*pb.Role, len(roles))
	for i, role := range roles {
		pbRoles[i] = dbRoleToProto(&role)
	}

	return &pb.ListUserRolesResponse{Roles: pbRoles}, nil
}

// Helper function to convert db.User to pb.User
func dbUserToProto(user *db.User) *pb.User {
	pbUser := &pb.User{
		Id:            helpers.PgUUIDToString(user.ID),
		FullName:      user.FullName,
		Email:         user.Email,
		TenantId:      helpers.PgUUIDToString(user.TenantID),
		EmailVerified: user.EmailVerified,
		Status:        string(user.Status),
	}

	if user.EmployeeID.Valid {
		pbUser.EmployeeId = user.EmployeeID.String
	}
	if user.DepartmentID.Valid {
		pbUser.DepartmentId = helpers.PgUUIDToString(user.DepartmentID)
	}
	if user.DesignationID.Valid {
		pbUser.DesignationId = helpers.PgUUIDToString(user.DesignationID)
	}
	if user.PhoneNumber.Valid {
		pbUser.PhoneNumber = user.PhoneNumber.String
	}
	if user.JobTitle.Valid {
		pbUser.JobTitle = user.JobTitle.String
	}
	if user.CreatedAt.Valid {
		pbUser.CreatedAt = timestamppb.New(user.CreatedAt.Time)
	}
	if user.UpdatedAt.Valid {
		pbUser.UpdatedAt = timestamppb.New(user.UpdatedAt.Time)
	}

	return pbUser
}

// Helper function to convert db.Role to pb.Role
func dbRoleToProto(role *db.Role) *pb.Role {
	pbRole := &pb.Role{
		Id:          helpers.PgUUIDToString(role.ID),
		TenantId:    helpers.PgUUIDToString(role.TenantID),
		Name:        role.Name,
		Permissions: role.Permissions,
	}

	if role.Description.Valid {
		pbRole.Description = role.Description.String
	}
	if role.CreatedAt.Valid {
		pbRole.CreatedAt = timestamppb.New(role.CreatedAt.Time)
	}
	if role.UpdatedAt.Valid {
		pbRole.UpdatedAt = timestamppb.New(role.UpdatedAt.Time)
	}

	return pbRole
}
