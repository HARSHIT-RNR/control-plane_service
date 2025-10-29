package grpc_server

import (
	"context"

	pb "cp_service/api/proto/pb"
	"cp_service/internal/adapters/database/db"
	"cp_service/internal/adapters/helpers"
	"cp_service/internal/core/services"

	"google.golang.org/protobuf/types/known/emptypb"
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
	tenantID, err := helpers.StringToPgUUID(req.TenantId)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}
	
	params := db.CreateUserParams{
		FullName: req.FullName,
		Email:    req.Email,
		TenantID: tenantID,
		Status:   db.UserStatusACTIVE,
	}

	if req.EmployeeId != "" {
		params.EmployeeID = helpers.StringToPgText(req.EmployeeId)
	}
	if req.DepartmentId != "" {
		deptID, err := helpers.StringToPgUUID(req.DepartmentId)
		if err != nil {
			return nil, helpers.ToGRPCError(err)
		}
		params.DepartmentID = deptID
	}
	if req.DesignationId != "" {
		desigID, err := helpers.StringToPgUUID(req.DesignationId)
		if err != nil {
			return nil, helpers.ToGRPCError(err)
		}
		params.DesignationID = desigID
	}
	if req.JobTitle != "" {
		params.JobTitle = helpers.StringToPgText(req.JobTitle)
	}

	user, err := h.userService.CreateUser(ctx, params)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbUserToProto(&user), nil
}

func (h *UserHandler) InviteUser(ctx context.Context, req *pb.InviteUserRequest) (*emptypb.Empty, error) {
	// Build params with all fields including ERP fields
	tenantID, err := helpers.StringToPgUUID(req.TenantId)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	params := db.CreateUserParams{
		FullName: req.FullName,
		Email:    req.Email,
		TenantID: tenantID,
	}

	// Add ERP fields if provided
	if req.EmployeeId != "" {
		params.EmployeeID = helpers.StringToPgText(req.EmployeeId)
	}
	if req.DepartmentId != "" {
		deptID, err := helpers.StringToPgUUID(req.DepartmentId)
		if err != nil {
			return nil, helpers.ToGRPCError(err)
		}
		params.DepartmentID = deptID
	}
	if req.DesignationId != "" {
		desigID, err := helpers.StringToPgUUID(req.DesignationId)
		if err != nil {
			return nil, helpers.ToGRPCError(err)
		}
		params.DesignationID = desigID
	}
	if req.JobTitle != "" {
		params.JobTitle = helpers.StringToPgText(req.JobTitle)
	}
	if req.PhoneNumber != "" {
		params.PhoneNumber = helpers.StringToPgText(req.PhoneNumber)
	}

	_, err = h.userService.InviteUser(ctx, params, req.RoleIds)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	user, err := h.userService.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbUserToProto(&user), nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	limit := req.PageSize
	if limit == 0 {
		limit = 50
	}
	
	// TODO: Handle page_token for pagination
	offset := int32(0)

	users, err := h.userService.ListUsers(ctx, req.TenantId, limit, offset)
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
	userID, err := helpers.StringToPgUUID(req.UserId)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}
	
	params := db.UpdateUserParams{
		ID:       userID,
		FullName: req.FullName,
	}

	if req.JobTitle != "" {
		params.JobTitle = helpers.StringToPgText(req.JobTitle)
	}
	if req.EmployeeId != "" {
		params.EmployeeID = helpers.StringToPgText(req.EmployeeId)
	}
	if req.DepartmentId != "" {
		deptID, err := helpers.StringToPgUUID(req.DepartmentId)
		if err != nil {
			return nil, helpers.ToGRPCError(err)
		}
		params.DepartmentID = deptID
	}
	if req.DesignationId != "" {
		desigID, err := helpers.StringToPgUUID(req.DesignationId)
		if err != nil {
			return nil, helpers.ToGRPCError(err)
		}
		params.DesignationID = desigID
	}

	user, err := h.userService.UpdateUser(ctx, params)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbUserToProto(&user), nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := h.userService.DeleteUser(ctx, req.UserId); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *UserHandler) AssignRoleToUser(ctx context.Context, req *pb.AssignRoleToUserRequest) (*emptypb.Empty, error) {
	if err := h.userService.AssignRoleToUser(ctx, req.UserId, req.TenantId, req.RoleId); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *UserHandler) RevokeRoleFromUser(ctx context.Context, req *pb.RevokeRoleFromUserRequest) (*emptypb.Empty, error) {
	if err := h.userService.RevokeRoleFromUser(ctx, req.UserId, req.TenantId, req.RoleId); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
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

func (h *UserHandler) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.Role, error) {
	role, err := h.userService.CreateRole(ctx, req.TenantId, req.Name, req.Description, req.Permissions)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbRoleToProto(&role), nil
}

func (h *UserHandler) GetRole(ctx context.Context, req *pb.GetRoleRequest) (*pb.Role, error) {
	role, err := h.userService.GetRole(ctx, req.RoleId)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbRoleToProto(&role), nil
}

func (h *UserHandler) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (*pb.ListRolesResponse, error) {
	roles, err := h.userService.ListRoles(ctx, req.TenantId)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	pbRoles := make([]*pb.Role, len(roles))
	for i, role := range roles {
		pbRoles[i] = dbRoleToProto(&role)
	}

	return &pb.ListRolesResponse{Roles: pbRoles}, nil
}

func (h *UserHandler) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.Role, error) {
	role, err := h.userService.UpdateRole(ctx, req.RoleId, req.Name, req.Description, req.Permissions)
	if err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return dbRoleToProto(&role), nil
}

func (h *UserHandler) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*emptypb.Empty, error) {
	if err := h.userService.DeleteRole(ctx, req.RoleId); err != nil {
		return nil, helpers.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

// Helper function to convert db.User to pb.User
func dbUserToProto(user *db.User) *pb.User {
	// Convert DB UserStatus to proto UserStatus
	var protoStatus pb.UserStatus
	switch user.Status {
	case db.UserStatusPENDINGSETUP:
		protoStatus = pb.UserStatus_PENDING_SETUP
	case db.UserStatusPENDINGINVITE:
		protoStatus = pb.UserStatus_PENDING_INVITE
	case db.UserStatusACTIVE:
		protoStatus = pb.UserStatus_USER_ACTIVE
	case db.UserStatusSUSPENDED:
		protoStatus = pb.UserStatus_USER_SUSPENDED
	default:
		protoStatus = pb.UserStatus_USER_STATUS_UNSPECIFIED
	}
	
	pbUser := &pb.User{
		UserId:        helpers.PgUUIDToString(user.ID),
		FullName:      user.FullName,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Status:        protoStatus,
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
	if user.LastLoginAt.Valid {
		pbUser.LastLoginAt = timestamppb.New(user.LastLoginAt.Time)
	}

	return pbUser
}

// Helper function to convert db.Role to pb.Role
func dbRoleToProto(role *db.Role) *pb.Role {
	pbRole := &pb.Role{
		RoleId:      helpers.PgUUIDToString(role.ID),
		TenantId:    helpers.PgUUIDToString(role.TenantID),
		Name:        role.Name,
		Permissions: role.Permissions,
	}

	if role.Description.Valid {
		pbRole.Description = role.Description.String
	}

	return pbRole
}
