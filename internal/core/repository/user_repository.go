package repository

import (
	"context"

	"cp_service/internal/adapters/database/db"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// User CRUD operations
	CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error)
	CreateInitialAdmin(ctx context.Context, params db.CreateInitialAdminParams) (db.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error)
	GetUserByEmail(ctx context.Context, email string, tenantID uuid.UUID) (db.User, error)
	ListUsers(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]db.User, error)
	UpdateUser(ctx context.Context, params db.UpdateUserParams) (db.User, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status db.UserStatus) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ActivateInvitedUser(ctx context.Context, params db.ActivateInvitedUserParams) error
}
