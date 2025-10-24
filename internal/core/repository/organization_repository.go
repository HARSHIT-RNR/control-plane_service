package repository

import (
	"context"

	"cp_service/internal/adapters/database/db"

	"github.com/google/uuid"
)

// OrganizationRepository defines the interface for organization structure operations
type OrganizationRepository interface {
	// Department operations
	CreateDepartment(ctx context.Context, params db.CreateDepartmentParams) (db.Department, error)
	GetDepartmentByID(ctx context.Context, id uuid.UUID) (db.Department, error)
	ListDepartments(ctx context.Context, tenantID uuid.UUID) ([]db.Department, error)
	UpdateDepartment(ctx context.Context, params db.UpdateDepartmentParams) (db.Department, error)
	DeleteDepartment(ctx context.Context, id uuid.UUID) error

	// Designation operations
	CreateDesignation(ctx context.Context, params db.CreateDesignationParams) (db.Designation, error)
	GetDesignationByID(ctx context.Context, id uuid.UUID) (db.Designation, error)
	ListDesignations(ctx context.Context, tenantID uuid.UUID) ([]db.Designation, error)
	UpdateDesignation(ctx context.Context, params db.UpdateDesignationParams) (db.Designation, error)
	DeleteDesignation(ctx context.Context, id uuid.UUID) error
}
