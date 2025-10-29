package services

import (
	"context"
	"fmt"

	"cp_service/internal/adapters/database/db"
	events "cp_service/internal/adapters/kafka"
	"cp_service/internal/core/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// EventProducer defines the interface for publishing domain events
type EventProducer interface {
	PublishUserCreated(ctx context.Context, event events.UserCreatedEvent) error
	PublishUserInvited(ctx context.Context, event events.UserInvitedEvent) error
	PublishUserUpdated(ctx context.Context, event events.UserUpdatedEvent) error
	PublishUserDeleted(ctx context.Context, event events.UserDeletedEvent) error
	PublishRoleAssigned(ctx context.Context, event events.RoleAssignedEvent) error
	PublishRoleRevoked(ctx context.Context, event events.RoleRevokedEvent) error
	PublishUserStatusChanged(ctx context.Context, event events.UserStatusChangedEvent) error
	PublishPasswordChanged(ctx context.Context, event events.PasswordChangedEvent) error
	PublishUserLogin(ctx context.Context, event events.UserLoginEvent) error
}

// UserService handles user-related business logic
type UserService struct {
	userRepo      repository.UserRepository
	roleRepo      repository.RoleRepository
	eventProducer EventProducer
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	eventProducer EventProducer,
) *UserService {
	return &UserService{
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		eventProducer: eventProducer,
	}
}

// CreateInitialAdmin creates the first admin user for a new tenant (Step 10 in onboarding)
func (s *UserService) CreateInitialAdmin(ctx context.Context, tenantID, email, fullName string) (db.User, error) {
	// Validate tenant ID
	if _, err := uuid.Parse(tenantID); err != nil {
		return db.User{}, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Create admin user with PENDING_SETUP status
	user, err := s.userRepo.CreateInitialAdmin(ctx, db.CreateInitialAdminParams{
		FullName: fullName,
		Email:    email,
		TenantID: pgUUID(tenantID),
	})
	if err != nil {
		return db.User{}, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create default "Tenant Admin" role for this tenant
	adminRole, err := s.roleRepo.CreateRole(ctx, db.CreateRoleParams{
		ID:          pgUUID(uuid.New().String()),
		TenantID:    pgUUID(tenantID),
		Name:        "Tenant Admin",
		Description: pgText("Full administrative access to tenant"),
		Permissions: []string{
			"users:create", "users:read", "users:update", "users:delete",
			"roles:create", "roles:read", "roles:update", "roles:delete",
			"departments:create", "departments:read", "departments:update", "departments:delete",
			"designations:create", "designations:read", "designations:update", "designations:delete",
		},
	})
	if err != nil {
		return db.User{}, fmt.Errorf("failed to create admin role: %w", err)
	}

	// Assign admin role to user
	userID, _ := uuid.FromBytes(user.ID.Bytes[:])
	
	if err := s.roleRepo.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID:   user.ID,
		TenantID: user.TenantID,
		RoleID:   adminRole.ID,
	}); err != nil {
		return db.User{}, fmt.Errorf("failed to assign admin role: %w", err)
	}

	// Publish UserCreated event to Kafka (Step 11 in onboarding flow)
	event := events.UserCreatedEvent{
		UserID:         userID.String(),
		TenantID:       tenantID,
		Email:          email,
		IsInitialAdmin: true,
	}

	if err := s.eventProducer.PublishUserCreated(ctx, event); err != nil {
		return db.User{}, fmt.Errorf("failed to publish user created event: %w", err)
	}

	return user, nil
}

// CreateUser creates a new user (regular employee)
func (s *UserService) CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error) {
	// Create user
	user, err := s.userRepo.CreateUser(ctx, params)
	if err != nil {
		return db.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	// Extract UUID from pgtype.UUID
	userID, _ := uuid.FromBytes(user.ID.Bytes[:])
	tenantID, _ := uuid.FromBytes(user.TenantID.Bytes[:])

	// Publish UserCreated event
	event := events.UserCreatedEvent{
		UserID:         userID.String(),
		TenantID:       tenantID.String(),
		Email:          user.Email,
		IsInitialAdmin: false,
	}

	if err := s.eventProducer.PublishUserCreated(ctx, event); err != nil {
		return db.User{}, fmt.Errorf("failed to publish user created event: %w", err)
	}

	return user, nil
}

// InviteUser creates a user with PENDING_INVITE status
func (s *UserService) InviteUser(ctx context.Context, params db.CreateUserParams, roleIDs []string) (db.User, error) {
	// Extract tenant ID for validation
	tenantID, _ := uuid.FromBytes(params.TenantID.Bytes[:])
	if tenantID == uuid.Nil {
		return db.User{}, fmt.Errorf("invalid tenant ID")
	}

	// Set status to PENDING_INVITE
	params.Status = db.UserStatusPENDINGINVITE

	// Create user with PENDING_INVITE status
	user, err := s.userRepo.CreateUser(ctx, params)
	if err != nil {
		return db.User{}, fmt.Errorf("failed to create invited user: %w", err)
	}

	// Assign roles
	for _, roleID := range roleIDs {
		// Validate role ID
		if _, err := uuid.Parse(roleID); err != nil {
			return db.User{}, fmt.Errorf("invalid role ID: %w", err)
		}

		if err := s.roleRepo.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
			UserID:   user.ID,
			TenantID: user.TenantID,
			RoleID:   pgUUID(roleID),
		}); err != nil {
			return db.User{}, fmt.Errorf("failed to assign role: %w", err)
		}
	}

	// Extract UUIDs
	userID, _ := uuid.FromBytes(user.ID.Bytes[:])

	// Publish UserInvited event
	event := events.UserInvitedEvent{
		UserID:   userID.String(),
		TenantID: tenantID.String(),
		Email:    params.Email,
		FullName: params.FullName,
	}

	if err := s.eventProducer.PublishUserInvited(ctx, event); err != nil {
		return db.User{}, fmt.Errorf("failed to publish user invited event: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID string) (db.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return db.User{}, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return db.User{}, err
	}

	return user, nil
}

// ListUsers retrieves users for a tenant
func (s *UserService) ListUsers(ctx context.Context, tenantID string, limit, offset int32) ([]db.User, error) {
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	users, err := s.userRepo.ListUsers(ctx, id, limit, offset)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, params db.UpdateUserParams) (db.User, error) {
	user, err := s.userRepo.UpdateUser(ctx, params)
	if err != nil {
		return db.User{}, err
	}

	// Extract UUID
	userID, _ := uuid.FromBytes(user.ID.Bytes[:])
	tenantID, _ := uuid.FromBytes(user.TenantID.Bytes[:])

	// Publish UserUpdated event
	event := events.UserUpdatedEvent{
		UserID:   userID.String(),
		TenantID: tenantID.String(),
	}

	if err := s.eventProducer.PublishUserUpdated(ctx, event); err != nil {
		return db.User{}, fmt.Errorf("failed to publish user updated event: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user first to extract tenant ID for event
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.userRepo.DeleteUser(ctx, id); err != nil {
		return err
	}

	// Extract UUIDs
	tenantID, _ := uuid.FromBytes(user.TenantID.Bytes[:])

	// Publish UserDeleted event
	event := events.UserDeletedEvent{
		UserID:   userID,
		TenantID: tenantID.String(),
	}

	if err := s.eventProducer.PublishUserDeleted(ctx, event); err != nil {
		return fmt.Errorf("failed to publish user deleted event: %w", err)
	}

	return nil
}

// AssignRoleToUser assigns a role to a user
func (s *UserService) AssignRoleToUser(ctx context.Context, userID, tenantID, roleID string) error {
	if err := s.roleRepo.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID:   pgUUID(userID),
		TenantID: pgUUID(tenantID),
		RoleID:   pgUUID(roleID),
	}); err != nil {
		return err
	}

	// Publish RoleAssigned event
	event := events.RoleAssignedEvent{
		UserID:   userID,
		TenantID: tenantID,
		RoleID:   roleID,
	}

	if err := s.eventProducer.PublishRoleAssigned(ctx, event); err != nil {
		return fmt.Errorf("failed to publish role assigned event: %w", err)
	}

	return nil
}

// RevokeRoleFromUser revokes a role from a user
func (s *UserService) RevokeRoleFromUser(ctx context.Context, userID, tenantID, roleID string) error {
	if err := s.roleRepo.RevokeRoleFromUser(ctx, db.RevokeRoleFromUserParams{
		UserID: pgUUID(userID),
		RoleID: pgUUID(roleID),
	}); err != nil {
		return err
	}

	// Publish RoleRevoked event
	event := events.RoleRevokedEvent{
		UserID:   userID,
		TenantID: tenantID,
		RoleID:   roleID,
	}

	if err := s.eventProducer.PublishRoleRevoked(ctx, event); err != nil {
		return fmt.Errorf("failed to publish role revoked event: %w", err)
	}

	return nil
}

// ListUserRoles retrieves all roles assigned to a user
func (s *UserService) ListUserRoles(ctx context.Context, userID string) ([]db.Role, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	roles, err := s.roleRepo.ListUserRoles(ctx, id)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// CreateRole creates a new role
func (s *UserService) CreateRole(ctx context.Context, tenantID, name, description string, permissions []string) (db.Role, error) {
	// Validate tenant ID
	if _, err := uuid.Parse(tenantID); err != nil {
		return db.Role{}, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Create role
	role, err := s.roleRepo.CreateRole(ctx, db.CreateRoleParams{
		ID:          pgUUID(uuid.New().String()),
		TenantID:    pgUUID(tenantID),
		Name:        name,
		Description: pgText(description),
		Permissions: permissions,
	})
	if err != nil {
		return db.Role{}, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRole retrieves a role by ID
func (s *UserService) GetRole(ctx context.Context, roleID string) (db.Role, error) {
	id, err := uuid.Parse(roleID)
	if err != nil {
		return db.Role{}, fmt.Errorf("invalid role ID: %w", err)
	}

	role, err := s.roleRepo.GetRoleByID(ctx, id)
	if err != nil {
		return db.Role{}, err
	}

	return role, nil
}

// ListRoles retrieves all roles for a tenant
func (s *UserService) ListRoles(ctx context.Context, tenantID string) ([]db.Role, error) {
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	roles, err := s.roleRepo.ListRoles(ctx, id)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// UpdateRole updates a role
func (s *UserService) UpdateRole(ctx context.Context, roleID, name, description string, permissions []string) (db.Role, error) {
	// Validate role ID
	if _, err := uuid.Parse(roleID); err != nil {
		return db.Role{}, fmt.Errorf("invalid role ID: %w", err)
	}

	// Update role
	role, err := s.roleRepo.UpdateRole(ctx, db.UpdateRoleParams{
		ID:          pgUUID(roleID),
		Name:        name,
		Description: pgText(description),
		Permissions: permissions,
	})
	if err != nil {
		return db.Role{}, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// DeleteRole deletes a role
func (s *UserService) DeleteRole(ctx context.Context, roleID string) error {
	id, err := uuid.Parse(roleID)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	if err := s.roleRepo.DeleteRole(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// Helper function to convert string UUID to pgtype.UUID
func pgUUID(id string) pgtype.UUID {
	uid, _ := uuid.Parse(id)
	return pgtype.UUID{
		Bytes: uid,
		Valid: true,
	}
}

// Helper function to convert string to pgtype.Text
func pgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}
