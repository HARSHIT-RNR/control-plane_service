package repository

import (
	"context"
	"fmt"

	"cp_service/internal/adapters/database/db"
	"cp_service/internal/core/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type organizationRepository struct {
	queries *db.Queries
}

// NewOrganizationRepository creates a new organization repository implementation
func NewOrganizationRepository(queries *db.Queries) repository.OrganizationRepository {
	return &organizationRepository{
		queries: queries,
	}
}

// Department operations

func (r *organizationRepository) CreateDepartment(ctx context.Context, params db.CreateDepartmentParams) (db.Department, error) {
	dept, err := r.queries.CreateDepartment(ctx, params)
	if err != nil {
		return db.Department{}, fmt.Errorf("failed to create department: %w", err)
	}
	return dept, nil
}

func (r *organizationRepository) GetDepartmentByID(ctx context.Context, id uuid.UUID) (db.Department, error) {
	dept, err := r.queries.GetDepartment(ctx, pgUUID(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.Department{}, fmt.Errorf("department not found")
		}
		return db.Department{}, fmt.Errorf("failed to get department: %w", err)
	}
	return dept, nil
}

func (r *organizationRepository) ListDepartments(ctx context.Context, tenantID uuid.UUID) ([]db.Department, error) {
	depts, err := r.queries.ListDepartments(ctx, db.ListDepartmentsParams{
		TenantID: pgUUID(tenantID),
		Limit:    100,
		Offset:   0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}
	return depts, nil
}

func (r *organizationRepository) UpdateDepartment(ctx context.Context, params db.UpdateDepartmentParams) (db.Department, error) {
	dept, err := r.queries.UpdateDepartment(ctx, params)
	if err != nil {
		return db.Department{}, fmt.Errorf("failed to update department: %w", err)
	}
	return dept, nil
}

func (r *organizationRepository) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteDepartment(ctx, pgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}
	return nil
}

// Designation operations

func (r *organizationRepository) CreateDesignation(ctx context.Context, params db.CreateDesignationParams) (db.Designation, error) {
	desig, err := r.queries.CreateDesignation(ctx, params)
	if err != nil {
		return db.Designation{}, fmt.Errorf("failed to create designation: %w", err)
	}
	return desig, nil
}

func (r *organizationRepository) GetDesignationByID(ctx context.Context, id uuid.UUID) (db.Designation, error) {
	desig, err := r.queries.GetDesignation(ctx, pgUUID(id))
	if err != nil {
		if err == pgx.ErrNoRows {
			return db.Designation{}, fmt.Errorf("designation not found")
		}
		return db.Designation{}, fmt.Errorf("failed to get designation: %w", err)
	}
	return desig, nil
}

func (r *organizationRepository) ListDesignations(ctx context.Context, tenantID uuid.UUID) ([]db.Designation, error) {
	desigs, err := r.queries.ListDesignations(ctx, db.ListDesignationsParams{
		TenantID: pgUUID(tenantID),
		Limit:    100,
		Offset:   0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list designations: %w", err)
	}
	return desigs, nil
}

func (r *organizationRepository) UpdateDesignation(ctx context.Context, params db.UpdateDesignationParams) (db.Designation, error) {
	desig, err := r.queries.UpdateDesignation(ctx, params)
	if err != nil {
		return db.Designation{}, fmt.Errorf("failed to update designation: %w", err)
	}
	return desig, nil
}

func (r *organizationRepository) DeleteDesignation(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteDesignation(ctx, pgUUID(id))
	if err != nil {
		return fmt.Errorf("failed to delete designation: %w", err)
	}
	return nil
}
