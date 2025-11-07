package service

import (
	"context"
	"testing"

	"cp_service/internal/adapters/database/db"
	"cp_service/internal/adapters/token"
	"cp_service/internal/core/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckAccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	tests := []struct {
		name         string
		action       string
		resource     string
		setup        func(*MockRoleRepository, *MockTokenValidator, *MockOPAClient)
		wantAllowed  bool
		wantReason   string
		wantError    bool
	}{
		{
			name:     "access granted - exact permission match",
			action:   "read",
			resource: "users",
			setup: func(mockRole *MockRoleRepository, mockToken *MockTokenValidator, mockOPA *MockOPAClient) {
				claims := &token.Claims{
					UserID:   userID.String(),
					TenantID: tenantID.String(),
					Email:    "user@example.com",
				}
				roles := []db.Role{
					{
						ID:          testPgUUID(uuid.New().String()),
						Name:        "Admin",
						Permissions: []string{"users:read", "users:write"},
					},
				}
				mockToken.On("ValidateToken", "access-token").Return(claims, nil)
				mockRole.On("ListUserRoles", ctx, userID).Return(roles, nil)
			},
			wantAllowed: true,
			wantReason:  "Access granted",
			wantError:   false,
		},
		{
			name:     "access granted - wildcard resource",
			action:   "delete",
			resource: "users",
			setup: func(mockRole *MockRoleRepository, mockToken *MockTokenValidator, mockOPA *MockOPAClient) {
				claims := &token.Claims{UserID: userID.String(), TenantID: tenantID.String()}
				roles := []db.Role{
					{ID: testPgUUID(uuid.New().String()), Name: "Admin", Permissions: []string{"users:*"}},
				}
				mockToken.On("ValidateToken", "access-token").Return(claims, nil)
				mockRole.On("ListUserRoles", ctx, userID).Return(roles, nil)
			},
			wantAllowed: true,
			wantReason:  "Access granted (wildcard resource)",
			wantError:   false,
		},
		{
			name:     "access granted - super admin",
			action:   "delete",
			resource: "users",
			setup: func(mockRole *MockRoleRepository, mockToken *MockTokenValidator, mockOPA *MockOPAClient) {
				claims := &token.Claims{UserID: userID.String(), TenantID: tenantID.String()}
				roles := []db.Role{
					{ID: testPgUUID(uuid.New().String()), Name: "SuperAdmin", Permissions: []string{"*:*"}},
				}
				mockToken.On("ValidateToken", "access-token").Return(claims, nil)
				mockRole.On("ListUserRoles", ctx, userID).Return(roles, nil)
			},
			wantAllowed: true,
			wantReason:  "Access granted (super admin)",
			wantError:   false,
		},
		{
			name:     "access denied - no matching permission",
			action:   "delete",
			resource: "users",
			setup: func(mockRole *MockRoleRepository, mockToken *MockTokenValidator, mockOPA *MockOPAClient) {
				claims := &token.Claims{UserID: userID.String(), TenantID: tenantID.String()}
				roles := []db.Role{
					{ID: testPgUUID(uuid.New().String()), Name: "Viewer", Permissions: []string{"users:read"}},
				}
				mockToken.On("ValidateToken", "access-token").Return(claims, nil)
				mockRole.On("ListUserRoles", ctx, userID).Return(roles, nil)
			},
			wantAllowed: false,
			wantReason:  "Access denied",
			wantError:   false,
		},
		{
			name:     "access granted - with OPA",
			action:   "read",
			resource: "sensitive-data",
			setup: func(mockRole *MockRoleRepository, mockToken *MockTokenValidator, mockOPA *MockOPAClient) {
				claims := &token.Claims{UserID: userID.String(), TenantID: tenantID.String(), Email: "user@example.com"}
				roles := []db.Role{
					{ID: testPgUUID(uuid.New().String()), Name: "Manager", Permissions: []string{"sensitive-data:read"}},
				}
				mockToken.On("ValidateToken", "access-token").Return(claims, nil)
				mockRole.On("ListUserRoles", ctx, userID).Return(roles, nil)
				mockOPA.On("Evaluate", ctx, mock.AnythingOfType("map[string]interface {}")).Return(true, nil)
			},
			wantAllowed: true,
			wantReason:  "Access granted by policy",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRole := new(MockRoleRepository)
			mockToken := new(MockTokenValidator)
			mockOPA := new(MockOPAClient)

			tt.setup(mockRole, mockToken, mockOPA)

			var service *services.AuthzService
			if tt.name == "access granted - with OPA" {
				service = services.NewAuthzService(mockRole, mockToken, mockOPA)
			} else {
				service = services.NewAuthzService(mockRole, mockToken, nil)
			}

			allowed, reason, err := service.CheckAccess(ctx, "access-token", tt.action, tt.resource)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantAllowed, allowed)
				assert.Contains(t, reason, tt.wantReason)
			}

			mockRole.AssertExpectations(t)
			mockToken.AssertExpectations(t)
			if tt.name == "access granted - with OPA" {
				mockOPA.AssertExpectations(t)
			}
		})
	}
}

func TestGetUserPermissions(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name       string
		setup      func(*MockRoleRepository)
		wantPerms  int
		wantError  bool
	}{
		{
			name: "multiple roles with overlapping permissions",
			setup: func(mockRole *MockRoleRepository) {
				roles := []db.Role{
					{ID: testPgUUID(uuid.New().String()), Permissions: []string{"users:read", "users:write"}},
					{ID: testPgUUID(uuid.New().String()), Permissions: []string{"users:read", "roles:read"}},
				}
				mockRole.On("ListUserRoles", ctx, userID).Return(roles, nil)
			},
			wantPerms: 3, // users:read, users:write, roles:read (deduplicated)
			wantError: false,
		},
		{
			name: "no roles",
			setup: func(mockRole *MockRoleRepository) {
				mockRole.On("ListUserRoles", ctx, userID).Return([]db.Role{}, nil)
			},
			wantPerms: 0,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRole := new(MockRoleRepository)
			tt.setup(mockRole)

			service := services.NewAuthzService(mockRole, new(MockTokenValidator), nil)
			permissions, err := service.GetUserPermissions(ctx, userID.String())

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantPerms, len(permissions))
			}

			mockRole.AssertExpectations(t)
		})
	}
}

func TestHasPermission(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name       string
		permission string
		setup      func(*MockRoleRepository)
		wantHas    bool
		wantError  bool
	}{
		{
			name:       "has exact permission",
			permission: "users:read",
			setup: func(mockRole *MockRoleRepository) {
				roles := []db.Role{
					{ID: testPgUUID(uuid.New().String()), Permissions: []string{"users:read", "users:write"}},
				}
				mockRole.On("ListUserRoles", ctx, userID).Return(roles, nil)
			},
			wantHas:   true,
			wantError: false,
		},
		{
			name:       "has wildcard permission",
			permission: "users:delete",
			setup: func(mockRole *MockRoleRepository) {
				roles := []db.Role{
					{ID: testPgUUID(uuid.New().String()), Permissions: []string{"users:*"}},
				}
				mockRole.On("ListUserRoles", ctx, userID).Return(roles, nil)
			},
			wantHas:   true,
			wantError: false,
		},
		{
			name:       "no permission",
			permission: "roles:delete",
			setup: func(mockRole *MockRoleRepository) {
				roles := []db.Role{
					{ID: testPgUUID(uuid.New().String()), Permissions: []string{"users:read"}},
				}
				mockRole.On("ListUserRoles", ctx, userID).Return(roles, nil)
			},
			wantHas:   false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRole := new(MockRoleRepository)
			tt.setup(mockRole)

			service := services.NewAuthzService(mockRole, new(MockTokenValidator), nil)
			has, err := service.HasPermission(ctx, userID.String(), tt.permission)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantHas, has)
			}

			mockRole.AssertExpectations(t)
		})
	}
}
