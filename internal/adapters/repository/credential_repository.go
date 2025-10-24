package repository

import (
	"context"
	"database/sql"
	"fmt"

	"cp_service/internal/adapters/database/db"
	"cp_service/internal/core/repository"

	"github.com/google/uuid"
)

type credentialRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// NewCredentialRepository creates a new credential repository implementation
func NewCredentialRepository(database *sql.DB, queries *db.Queries) repository.CredentialRepository {
	return &credentialRepository{
		db:      database,
		queries: queries,
	}
}

func (r *credentialRepository) CreateCredential(ctx context.Context, params db.CreateCredentialParams) error {
	err := r.queries.CreateCredential(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}
	return nil
}

func (r *credentialRepository) GetCredentialByUserID(ctx context.Context, userID uuid.UUID) (db.Credential, error) {
	credential, err := r.queries.GetCredentialByUserID(ctx, pgUUID(userID))
	if err != nil {
		if err == sql.ErrNoRows {
			return db.Credential{}, fmt.Errorf("credential not found")
		}
		return db.Credential{}, fmt.Errorf("failed to get credential: %w", err)
	}
	return credential, nil
}

func (r *credentialRepository) UpdateCredential(ctx context.Context, params db.UpdateCredentialParams) error {
	err := r.queries.UpdateCredential(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}
	return nil
}

func (r *credentialRepository) DeleteCredential(ctx context.Context, userID uuid.UUID) error {
	err := r.queries.DeleteCredential(ctx, pgUUID(userID))
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	return nil
}

func (r *credentialRepository) CreateToken(ctx context.Context, params db.CreateTokenParams) error {
	err := r.queries.CreateToken(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}
	return nil
}

func (r *credentialRepository) GetToken(ctx context.Context, hash []byte) (db.Token, error) {
	token, err := r.queries.GetToken(ctx, hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.Token{}, fmt.Errorf("token not found")
		}
		return db.Token{}, fmt.Errorf("failed to get token: %w", err)
	}
	return token, nil
}

func (r *credentialRepository) DeleteToken(ctx context.Context, hash []byte) error {
	err := r.queries.DeleteToken(ctx, hash)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

func (r *credentialRepository) DeleteAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	err := r.queries.DeleteAllUserTokens(ctx, pgUUID(userID))
	if err != nil {
		return fmt.Errorf("failed to delete all user tokens: %w", err)
	}
	return nil
}
