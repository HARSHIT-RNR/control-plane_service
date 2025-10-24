package repository

import (
	"context"

	"cp_service/internal/adapters/database/db"

	"github.com/google/uuid"
)

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	// Role CRUD operations
	CreateRole(ctx context.Context, params db.CreateRoleParams) (db.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (db.Role, error)
	ListRoles(ctx context.Context, tenantID uuid.UUID) ([]db.Role, error)
	UpdateRole(ctx context.Context, params db.UpdateRoleParams) (db.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error

	// User-Role assignment
	AssignRoleToUser(ctx context.Context, params db.AssignRoleToUserParams) error
	RevokeRoleFromUser(ctx context.Context, params db.RevokeRoleFromUserParams) error
	ListUserRoles(ctx context.Context, userID uuid.UUID) ([]db.Role, error)
}
