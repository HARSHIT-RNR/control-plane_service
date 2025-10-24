package repository

import (
	"context"

	"cp_service/internal/adapters/database/db"

	"github.com/google/uuid"
)

// CredentialRepository defines the interface for credential data operations
type CredentialRepository interface {
	CreateCredential(ctx context.Context, params db.CreateCredentialParams) error
	GetCredentialByUserID(ctx context.Context, userID uuid.UUID) (db.Credential, error)
	UpdateCredential(ctx context.Context, params db.UpdateCredentialParams) error
	DeleteCredential(ctx context.Context, userID uuid.UUID) error

	// Token operations
	CreateToken(ctx context.Context, params db.CreateTokenParams) error
	GetToken(ctx context.Context, hash []byte) (db.Token, error)
	DeleteToken(ctx context.Context, hash []byte) error
	DeleteAllUserTokens(ctx context.Context, userID uuid.UUID) error
}
