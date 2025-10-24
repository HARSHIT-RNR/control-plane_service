package services

import (
	"context"
	"fmt"

	"cp_service/internal/adapters/database/db"
	"cp_service/internal/core/repository"

	"github.com/google/uuid"
)

// OrganizationService handles organization structure business logic
type OrganizationService struct {
	orgRepo repository.OrganizationRepository
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(orgRepo repository.OrganizationRepository) *OrganizationService {
	return &OrganizationService{
		orgRepo: orgRepo,
	}
}

// Department operations

func (s *OrganizationService) CreateDepartment(ctx context.Context, params db.CreateDepartmentParams) (db.Department, error) {
	dept, err := s.orgRepo.CreateDepartment(ctx, params)
	if err != nil {
		return db.Department{}, fmt.Errorf("failed to create department: %w", err)
	}
	return dept, nil
}

func (s *OrganizationService) GetDepartment(ctx context.Context, id string) (db.Department, error) {
	deptID, err := uuid.Parse(id)
	if err != nil {
		return db.Department{}, fmt.Errorf("invalid department ID: %w", err)
	}

	dept, err := s.orgRepo.GetDepartmentByID(ctx, deptID)
	if err != nil {
		return db.Department{}, err
	}

	return dept, nil
}

func (s *OrganizationService) ListDepartments(ctx context.Context, tenantID string) ([]db.Department, error) {
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	depts, err := s.orgRepo.ListDepartments(ctx, tid)
	if err != nil {
		return nil, err
	}

	return depts, nil
}

func (s *OrganizationService) UpdateDepartment(ctx context.Context, params db.UpdateDepartmentParams) (db.Department, error) {
	dept, err := s.orgRepo.UpdateDepartment(ctx, params)
	if err != nil {
		return db.Department{}, fmt.Errorf("failed to update department: %w", err)
	}
	return dept, nil
}

func (s *OrganizationService) DeleteDepartment(ctx context.Context, id string) error {
	deptID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid department ID: %w", err)
	}

	if err := s.orgRepo.DeleteDepartment(ctx, deptID); err != nil {
		return err
	}

	return nil
}

// Designation operations

func (s *OrganizationService) CreateDesignation(ctx context.Context, params db.CreateDesignationParams) (db.Designation, error) {
	desig, err := s.orgRepo.CreateDesignation(ctx, params)
	if err != nil {
		return db.Designation{}, fmt.Errorf("failed to create designation: %w", err)
	}
	return desig, nil
}

func (s *OrganizationService) GetDesignation(ctx context.Context, id string) (db.Designation, error) {
	desigID, err := uuid.Parse(id)
	if err != nil {
		return db.Designation{}, fmt.Errorf("invalid designation ID: %w", err)
	}

	desig, err := s.orgRepo.GetDesignationByID(ctx, desigID)
	if err != nil {
		return db.Designation{}, err
	}

	return desig, nil
}

func (s *OrganizationService) ListDesignations(ctx context.Context, tenantID string) ([]db.Designation, error) {
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	desigs, err := s.orgRepo.ListDesignations(ctx, tid)
	if err != nil {
		return nil, err
	}

	return desigs, nil
}

func (s *OrganizationService) UpdateDesignation(ctx context.Context, params db.UpdateDesignationParams) (db.Designation, error) {
	desig, err := s.orgRepo.UpdateDesignation(ctx, params)
	if err != nil {
		return db.Designation{}, fmt.Errorf("failed to update designation: %w", err)
	}
	return desig, nil
}

func (s *OrganizationService) DeleteDesignation(ctx context.Context, id string) error {
	desigID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid designation ID: %w", err)
	}

	if err := s.orgRepo.DeleteDesignation(ctx, desigID); err != nil {
		return err
	}

	return nil
}
