package service

import (
	"context"
	"testing"

	"cp_service/internal/adapters/database/db"
	events "cp_service/internal/adapters/kafka"
	"cp_service/internal/core/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateInitialAdmin(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New().String()
	userID := uuid.New()
	roleID := uuid.New()

	tests := []struct {
		name      string
		tenantID  string
		setup     func(*MockUserRepository, *MockRoleRepository, *MockEventProducer)
		wantError bool
		errMsg    string
	}{
		{
			name:     "successful admin creation",
			tenantID: tenantID,
			setup: func(userRepo *MockUserRepository, roleRepo *MockRoleRepository, eventProducer *MockEventProducer) {
				userRepo.On("CreateInitialAdmin", ctx, mock.AnythingOfType("db.CreateInitialAdminParams")).Return(db.User{
					ID:       testPgUUID(userID.String()),
					Email:    "admin@example.com",
					FullName: "Admin User",
					TenantID: testPgUUID(tenantID),
					Status:   db.UserStatusPENDINGSETUP,
				}, nil)

				roleRepo.On("CreateRole", ctx, mock.AnythingOfType("db.CreateRoleParams")).Return(db.Role{
					ID:       testPgUUID(roleID.String()),
					TenantID: testPgUUID(tenantID),
					Name:     "Tenant Admin",
				}, nil)

				roleRepo.On("AssignRoleToUser", ctx, mock.AnythingOfType("db.AssignRoleToUserParams")).Return(nil)
				eventProducer.On("PublishUserCreated", ctx, mock.MatchedBy(func(event events.UserCreatedEvent) bool {
					return event.IsInitialAdmin == true
				})).Return(nil)
			},
			wantError: false,
		},
		{
			name:     "invalid tenant ID",
			tenantID: "invalid-uuid",
			setup:    func(userRepo *MockUserRepository, roleRepo *MockRoleRepository, eventProducer *MockEventProducer) {},
			wantError: true,
			errMsg:    "invalid tenant ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			roleRepo := new(MockRoleRepository)
			eventProducer := new(MockEventProducer)

			tt.setup(userRepo, roleRepo, eventProducer)

			service := services.NewUserService(userRepo, roleRepo, eventProducer)
			_, err := service.CreateInitialAdmin(ctx, tt.tenantID, "admin@example.com", "Admin User")

			if tt.wantError {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
			roleRepo.AssertExpectations(t)
			eventProducer.AssertExpectations(t)
		})
	}
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		eventProducer := new(MockEventProducer)

		userRepo.On("CreateUser", ctx, mock.AnythingOfType("db.CreateUserParams")).Return(db.User{
			ID:       testPgUUID(userID.String()),
			Email:    "user@example.com",
			FullName: "Test User",
		}, nil)

		eventProducer.On("PublishUserCreated", ctx, mock.AnythingOfType("events.UserCreatedEvent")).Return(nil)

		service := services.NewUserService(userRepo, new(MockRoleRepository), eventProducer)
		user, err := service.CreateUser(ctx, db.CreateUserParams{Email: "user@example.com", FullName: "Test User"})

		require.NoError(t, err)
		assert.Equal(t, "user@example.com", user.Email)
	})
}

func TestInviteUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()

	t.Run("successful invitation", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		roleRepo := new(MockRoleRepository)
		eventProducer := new(MockEventProducer)

		tenantID := uuid.New()
		userRepo.On("CreateUser", ctx, mock.MatchedBy(func(params db.CreateUserParams) bool {
			return params.Status == db.UserStatusPENDINGINVITE
		})).Return(db.User{
			ID:       testPgUUID(userID.String()),
			TenantID: testPgUUID(tenantID.String()),
			Status:   db.UserStatusPENDINGINVITE,
		}, nil)

		roleRepo.On("AssignRoleToUser", ctx, mock.AnythingOfType("db.AssignRoleToUserParams")).Return(nil)
		eventProducer.On("PublishUserInvited", ctx, mock.AnythingOfType("events.UserInvitedEvent")).Return(nil)

		service := services.NewUserService(userRepo, roleRepo, eventProducer)
		_, err := service.InviteUser(ctx, db.CreateUserParams{TenantID: testPgUUID(tenantID.String())}, []string{roleID.String()})

		require.NoError(t, err)
		userRepo.AssertExpectations(t)
		roleRepo.AssertExpectations(t)
		eventProducer.AssertExpectations(t)
	})
}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name      string
		userID    string
		setup     func(*MockUserRepository)
		wantError bool
	}{
		{
			name:   "success",
			userID: userID.String(),
			setup: func(userRepo *MockUserRepository) {
				userRepo.On("GetUserByID", ctx, userID).Return(db.User{
					ID:       testPgUUID(userID.String()),
					Email:    "user@example.com",
					FullName: "Test User",
				}, nil)
			},
			wantError: false,
		},
		{
			name:      "invalid ID",
			userID:    "invalid-uuid",
			setup:     func(userRepo *MockUserRepository) {},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			tt.setup(userRepo)

			service := services.NewUserService(userRepo, new(MockRoleRepository), new(MockEventProducer))
			_, err := service.GetUser(ctx, tt.userID)

			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	t.Run("success", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		users := []db.User{
			{ID: testPgUUID(uuid.New().String())},
			{ID: testPgUUID(uuid.New().String())},
		}

		userRepo.On("ListUsers", ctx, tenantID, int32(10), int32(0)).Return(users, nil)

		service := services.NewUserService(userRepo, new(MockRoleRepository), new(MockEventProducer))
		result, err := service.ListUsers(ctx, tenantID.String(), 10, 0)

		require.NoError(t, err)
		assert.Equal(t, 2, len(result))
	})
}

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		eventProducer := new(MockEventProducer)

		userRepo.On("UpdateUser", ctx, mock.AnythingOfType("db.UpdateUserParams")).Return(db.User{
			ID:       testPgUUID(userID.String()),
			FullName: "Updated Name",
		}, nil)

		eventProducer.On("PublishUserUpdated", ctx, mock.AnythingOfType("events.UserUpdatedEvent")).Return(nil)

		service := services.NewUserService(userRepo, new(MockRoleRepository), eventProducer)
		user, err := service.UpdateUser(ctx, db.UpdateUserParams{ID: testPgUUID(userID.String())})

		require.NoError(t, err)
		assert.Equal(t, "Updated Name", user.FullName)
	})
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("success", func(t *testing.T) {
		userRepo := new(MockUserRepository)
		eventProducer := new(MockEventProducer)

		userRepo.On("GetUserByID", ctx, userID).Return(db.User{
			ID:       testPgUUID(userID.String()),
			TenantID: testPgUUID(tenantID.String()),
		}, nil)

		userRepo.On("DeleteUser", ctx, userID).Return(nil)
		eventProducer.On("PublishUserDeleted", ctx, mock.AnythingOfType("events.UserDeletedEvent")).Return(nil)

		service := services.NewUserService(userRepo, new(MockRoleRepository), eventProducer)
		err := service.DeleteUser(ctx, userID.String())

		require.NoError(t, err)
	})
}

func TestAssignRoleToUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()
	tenantID := uuid.New().String()
	roleID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		eventProducer := new(MockEventProducer)

		roleRepo.On("AssignRoleToUser", ctx, mock.AnythingOfType("db.AssignRoleToUserParams")).Return(nil)
		eventProducer.On("PublishRoleAssigned", ctx, mock.AnythingOfType("events.RoleAssignedEvent")).Return(nil)

		service := services.NewUserService(new(MockUserRepository), roleRepo, eventProducer)
		err := service.AssignRoleToUser(ctx, userID, tenantID, roleID)

		require.NoError(t, err)
	})
}

func TestRevokeRoleFromUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()
	tenantID := uuid.New().String()
	roleID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		eventProducer := new(MockEventProducer)

		roleRepo.On("RevokeRoleFromUser", ctx, mock.AnythingOfType("db.RevokeRoleFromUserParams")).Return(nil)
		eventProducer.On("PublishRoleRevoked", ctx, mock.AnythingOfType("events.RoleRevokedEvent")).Return(nil)

		service := services.NewUserService(new(MockUserRepository), roleRepo, eventProducer)
		err := service.RevokeRoleFromUser(ctx, userID, tenantID, roleID)

		require.NoError(t, err)
	})
}

func TestListUserRoles(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		roles := []db.Role{
			{ID: testPgUUID(uuid.New().String()), Name: "Admin"},
			{ID: testPgUUID(uuid.New().String()), Name: "User"},
		}

		roleRepo.On("ListUserRoles", ctx, userID).Return(roles, nil)

		service := services.NewUserService(new(MockUserRepository), roleRepo, new(MockEventProducer))
		result, err := service.ListUserRoles(ctx, userID.String())

		require.NoError(t, err)
		assert.Equal(t, 2, len(result))
	})
}

func TestCreateRole(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		roleRepo.On("CreateRole", ctx, mock.AnythingOfType("db.CreateRoleParams")).Return(db.Role{
			ID:   testPgUUID(uuid.New().String()),
			Name: "Custom Role",
		}, nil)

		service := services.NewUserService(new(MockUserRepository), roleRepo, new(MockEventProducer))
		role, err := service.CreateRole(ctx, tenantID, "Custom Role", "Description", []string{"users:read"})

		require.NoError(t, err)
		assert.Equal(t, "Custom Role", role.Name)
	})
}

func TestUpdateRole(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		roleRepo.On("UpdateRole", ctx, mock.AnythingOfType("db.UpdateRoleParams")).Return(db.Role{
			ID:   testPgUUID(roleID),
			Name: "Updated Role",
		}, nil)

		service := services.NewUserService(new(MockUserRepository), roleRepo, new(MockEventProducer))
		role, err := service.UpdateRole(ctx, roleID, "Updated Role", "New Description", []string{"users:write"})

		require.NoError(t, err)
		assert.Equal(t, "Updated Role", role.Name)
	})
}

func TestDeleteRole(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.New()

	t.Run("success", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		roleRepo.On("DeleteRole", ctx, roleID).Return(nil)

		service := services.NewUserService(new(MockUserRepository), roleRepo, new(MockEventProducer))
		err := service.DeleteRole(ctx, roleID.String())

		require.NoError(t, err)
	})
}

func TestGetRole(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.New()
	tenantID := uuid.New()

	t.Run("success - retrieves role by ID", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		expectedRole := db.Role{
			ID:          testPgUUID(roleID.String()),
			TenantID:    testPgUUID(tenantID.String()),
			Name:        "Admin",
			Description: testPgText("Administrator role"),
			Permissions: []string{"users:read", "users:write", "roles:*"},
		}

		roleRepo.On("GetRoleByID", ctx, roleID).Return(expectedRole, nil)

		service := services.NewUserService(new(MockUserRepository), roleRepo, new(MockEventProducer))
		role, err := service.GetRole(ctx, roleID.String())

		require.NoError(t, err)
		assert.Equal(t, expectedRole.ID.Bytes, role.ID.Bytes)
		assert.Equal(t, expectedRole.Name, role.Name)
		assert.Equal(t, expectedRole.Description.String, role.Description.String)
		assert.Equal(t, expectedRole.Permissions, role.Permissions)
		roleRepo.AssertExpectations(t)
	})

	t.Run("error - invalid role ID", func(t *testing.T) {
		service := services.NewUserService(new(MockUserRepository), new(MockRoleRepository), new(MockEventProducer))
		_, err := service.GetRole(ctx, "invalid-uuid")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role ID")
	})

	t.Run("error - role not found", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		roleRepo.On("GetRoleByID", ctx, roleID).Return(db.Role{}, assert.AnError)

		service := services.NewUserService(new(MockUserRepository), roleRepo, new(MockEventProducer))
		_, err := service.GetRole(ctx, roleID.String())

		require.Error(t, err)
		roleRepo.AssertExpectations(t)
	})
}

func TestListRoles(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	t.Run("success - lists all roles for tenant", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		role1ID := uuid.New()
		role2ID := uuid.New()
		role3ID := uuid.New()
		
		expectedRoles := []db.Role{
			{
				ID:          testPgUUID(role1ID.String()),
				TenantID:    testPgUUID(tenantID.String()),
				Name:        "Admin",
				Description: testPgText("Administrator role"),
				Permissions: []string{"*:*"},
			},
			{
				ID:          testPgUUID(role2ID.String()),
				TenantID:    testPgUUID(tenantID.String()),
				Name:        "Manager",
				Description: testPgText("Manager role"),
				Permissions: []string{"users:read", "projects:*"},
			},
			{
				ID:          testPgUUID(role3ID.String()),
				TenantID:    testPgUUID(tenantID.String()),
				Name:        "User",
				Description: testPgText("Standard user role"),
				Permissions: []string{"projects:read", "tasks:read"},
			},
		}

		roleRepo.On("ListRoles", ctx, tenantID).Return(expectedRoles, nil)

		service := services.NewUserService(new(MockUserRepository), roleRepo, new(MockEventProducer))
		roles, err := service.ListRoles(ctx, tenantID.String())

		require.NoError(t, err)
		assert.Len(t, roles, 3)
		assert.Equal(t, "Admin", roles[0].Name)
		assert.Equal(t, "Manager", roles[1].Name)
		assert.Equal(t, "User", roles[2].Name)
		roleRepo.AssertExpectations(t)
	})

	t.Run("success - empty list when no roles exist", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		roleRepo.On("ListRoles", ctx, tenantID).Return([]db.Role{}, nil)

		service := services.NewUserService(new(MockUserRepository), roleRepo, new(MockEventProducer))
		roles, err := service.ListRoles(ctx, tenantID.String())

		require.NoError(t, err)
		assert.Empty(t, roles)
		roleRepo.AssertExpectations(t)
	})

	t.Run("error - invalid tenant ID", func(t *testing.T) {
		service := services.NewUserService(new(MockUserRepository), new(MockRoleRepository), new(MockEventProducer))
		_, err := service.ListRoles(ctx, "invalid-uuid")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tenant ID")
	})

	t.Run("error - database error", func(t *testing.T) {
		roleRepo := new(MockRoleRepository)
		roleRepo.On("ListRoles", ctx, tenantID).Return([]db.Role(nil), assert.AnError)

		service := services.NewUserService(new(MockUserRepository), roleRepo, new(MockEventProducer))
		_, err := service.ListRoles(ctx, tenantID.String())

		require.Error(t, err)
		roleRepo.AssertExpectations(t)
	})
}
