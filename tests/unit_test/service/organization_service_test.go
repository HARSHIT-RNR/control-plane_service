package service

import (
	"context"
	"errors"
	"testing"

	"cp_service/internal/adapters/database/db"
	"cp_service/internal/core/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Department Tests

func TestCreateDepartment(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		dept := db.Department{ID: testPgUUID(uuid.New().String()), Name: "Engineering"}
		
		mockOrg.On("CreateDepartment", ctx, mock.AnythingOfType("db.CreateDepartmentParams")).Return(dept, nil)

		service := services.NewOrganizationService(mockOrg)
		result, err := service.CreateDepartment(ctx, db.CreateDepartmentParams{Name: "Engineering"})

		require.NoError(t, err)
		assert.Equal(t, "Engineering", result.Name)
		mockOrg.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		mockOrg.On("CreateDepartment", ctx, mock.AnythingOfType("db.CreateDepartmentParams")).Return(db.Department{}, errors.New("db error"))

		service := services.NewOrganizationService(mockOrg)
		_, err := service.CreateDepartment(ctx, db.CreateDepartmentParams{Name: "Engineering"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create department")
	})
}

func TestGetDepartment(t *testing.T) {
	ctx := context.Background()
	deptID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		dept := db.Department{ID: testPgUUID(deptID.String()), Name: "Engineering"}
		
		mockOrg.On("GetDepartmentByID", ctx, deptID).Return(dept, nil)

		service := services.NewOrganizationService(mockOrg)
		result, err := service.GetDepartment(ctx, deptID.String())

		require.NoError(t, err)
		assert.Equal(t, "Engineering", result.Name)
	})

	t.Run("invalid ID", func(t *testing.T) {
		service := services.NewOrganizationService(new(MockOrganizationRepository))
		_, err := service.GetDepartment(ctx, "invalid-uuid")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid department ID")
	})
}

func TestListDepartments(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		depts := []db.Department{
			{ID: testPgUUID(uuid.New().String()), Name: "Engineering"},
			{ID: testPgUUID(uuid.New().String()), Name: "Sales"},
		}
		
		mockOrg.On("ListDepartments", ctx, tenantID).Return(depts, nil)

		service := services.NewOrganizationService(mockOrg)
		result, err := service.ListDepartments(ctx, tenantID.String())

		require.NoError(t, err)
		assert.Equal(t, 2, len(result))
	})
}

func TestUpdateDepartment(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		dept := db.Department{ID: testPgUUID(uuid.New().String()), Name: "Updated"}
		
		mockOrg.On("UpdateDepartment", ctx, mock.AnythingOfType("db.UpdateDepartmentParams")).Return(dept, nil)

		service := services.NewOrganizationService(mockOrg)
		result, err := service.UpdateDepartment(ctx, db.UpdateDepartmentParams{Name: "Updated"})

		require.NoError(t, err)
		assert.Equal(t, "Updated", result.Name)
	})
}

func TestDeleteDepartment(t *testing.T) {
	ctx := context.Background()
	deptID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		mockOrg.On("DeleteDepartment", ctx, deptID).Return(nil)

		service := services.NewOrganizationService(mockOrg)
		err := service.DeleteDepartment(ctx, deptID.String())

		require.NoError(t, err)
	})
}

// Designation Tests

func TestCreateDesignation(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		desig := db.Designation{ID: testPgUUID(uuid.New().String()), Name: "Senior Engineer"}
		
		mockOrg.On("CreateDesignation", ctx, mock.AnythingOfType("db.CreateDesignationParams")).Return(desig, nil)

		service := services.NewOrganizationService(mockOrg)
		result, err := service.CreateDesignation(ctx, db.CreateDesignationParams{Name: "Senior Engineer"})

		require.NoError(t, err)
		assert.Equal(t, "Senior Engineer", result.Name)
	})
}

func TestGetDesignation(t *testing.T) {
	ctx := context.Background()
	desigID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		desig := db.Designation{ID: testPgUUID(desigID.String()), Name: "Manager"}
		
		mockOrg.On("GetDesignationByID", ctx, desigID).Return(desig, nil)

		service := services.NewOrganizationService(mockOrg)
		result, err := service.GetDesignation(ctx, desigID.String())

		require.NoError(t, err)
		assert.Equal(t, "Manager", result.Name)
	})

	t.Run("invalid ID", func(t *testing.T) {
		service := services.NewOrganizationService(new(MockOrganizationRepository))
		_, err := service.GetDesignation(ctx, "invalid-uuid")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid designation ID")
	})
}

func TestListDesignations(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		desigs := []db.Designation{
			{ID: testPgUUID(uuid.New().String()), Name: "Manager"},
			{ID: testPgUUID(uuid.New().String()), Name: "Senior Engineer"},
		}
		
		mockOrg.On("ListDesignations", ctx, tenantID).Return(desigs, nil)

		service := services.NewOrganizationService(mockOrg)
		result, err := service.ListDesignations(ctx, tenantID.String())

		require.NoError(t, err)
		assert.Equal(t, 2, len(result))
	})
}

func TestUpdateDesignation(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		desig := db.Designation{ID: testPgUUID(uuid.New().String()), Name: "Lead Engineer"}
		
		mockOrg.On("UpdateDesignation", ctx, mock.AnythingOfType("db.UpdateDesignationParams")).Return(desig, nil)

		service := services.NewOrganizationService(mockOrg)
		result, err := service.UpdateDesignation(ctx, db.UpdateDesignationParams{Name: "Lead Engineer"})

		require.NoError(t, err)
		assert.Equal(t, "Lead Engineer", result.Name)
	})
}

func TestDeleteDesignation(t *testing.T) {
	ctx := context.Background()
	desigID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockOrg := new(MockOrganizationRepository)
		mockOrg.On("DeleteDesignation", ctx, desigID).Return(nil)

		service := services.NewOrganizationService(mockOrg)
		err := service.DeleteDesignation(ctx, desigID.String())

		require.NoError(t, err)
	})
}
