package repository

import (
	"context"
	"database/sql"
	"fmt"

	"cp_service/internal/adapters/database/db"
	"cp_service/internal/core/repository"

	"github.com/google/uuid"
)

type roleRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// NewRoleRepository creates a new role repository implementation
func NewRoleRepository(database *sql.DB, queries *db.Queries) repository.RoleRepository {
	return &roleRepository{
		db:      database,
		queries: queries,
	}
}

func (r *roleRepository) CreateRole(ctx context.Context, params db.CreateRoleParams) (db.Role, error) {
	role, err := r.queries.CreateRole(ctx, params)
	if err != nil {
		return db.Role{}, fmt.Errorf("failed to create role: %w", err)
	}
	return role, nil
}

func (r *roleRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (db.Role, error) {
	role, err := r.queries.GetRoleByID(ctx, pgUUID(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return db.Role{}, fmt.Errorf("role not found")
		}
		return db.Role{}, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

func (r *roleRepository) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]db.Role, error) {
	roles, err := r.queries.ListRoles(ctx, pgUUID(tenantID))
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

func (r *roleRepository) UpdateRole(ctx context.Context, params db.UpdateRoleParams) (db.Role, error) {
	role, err := r.queries.UpdateRole(ctx, params)
	if err != nil {
		return db.Role{}, fmt.Errorf("failed to update role: %w", err)
	}
	return role, nil
}

func (r *roleRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteRole(ctx, pgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (r *roleRepository) AssignRoleToUser(ctx context.Context, params db.AssignRoleToUserParams) error {
	err := r.queries.AssignRoleToUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}
	return nil
}

func (r *roleRepository) RevokeRoleFromUser(ctx context.Context, params db.RevokeRoleFromUserParams) error {
	err := r.queries.RevokeRoleFromUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to revoke role from user: %w", err)
	}
	return nil
}

func (r *roleRepository) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]db.Role, error) {
	roles, err := r.queries.ListUserRoles(ctx, pgUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to list user roles: %w", err)
	}
	return roles, nil
}
