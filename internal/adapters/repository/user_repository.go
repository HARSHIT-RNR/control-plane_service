package repository

import (
	"context"
	"fmt"

	"cp_service/internal/adapters/database/db"
	"cp_service/internal/core/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type userRepository struct {
	queries *db.Queries
}

// NewUserRepository creates a new user repository implementation
func NewUserRepository(queries *db.Queries) repository.UserRepository {
	return &userRepository{
		queries: queries,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error) {
	user, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return db.User{}, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (r *userRepository) CreateInitialAdmin(ctx context.Context, params db.CreateInitialAdminParams) (db.User, error) {
	user, err := r.queries.CreateInitialAdmin(ctx, params)
	if err != nil {
		return db.User{}, fmt.Errorf("failed to create initial admin: %w", err)
	}
	return user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	user, err := r.queries.GetUserByID(ctx, pgUUID(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.User{}, fmt.Errorf("user not found")
		}
		return db.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string, tenantID uuid.UUID) (db.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, db.GetUserByEmailParams{
		Email:    email,
		TenantID: pgUUID(tenantID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.User{}, fmt.Errorf("user not found")
		}
		return db.User{}, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (r *userRepository) ListUsers(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]db.User, error) {
	users, err := r.queries.ListUsers(ctx, db.ListUsersParams{
		TenantID: pgUUID(tenantID),
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, params db.UpdateUserParams) (db.User, error) {
	user, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return db.User{}, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

func (r *userRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status db.UserStatus) error {
	err := r.queries.UpdateUserStatus(ctx, db.UpdateUserStatusParams{
		ID:     pgUUID(userID),
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	err := r.queries.UpdateLastLogin(ctx, pgUUID(userID))
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteUser(ctx, pgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *userRepository) ActivateInvitedUser(ctx context.Context, params db.ActivateInvitedUserParams) error {
	err := r.queries.ActivateInvitedUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to activate invited user: %w", err)
	}
	return nil
}
