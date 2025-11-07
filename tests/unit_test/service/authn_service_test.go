package service

import (
	"context"
	"testing"
	"time"

	"cp_service/internal/adapters/database/db"
	"cp_service/internal/adapters/token"
	"cp_service/internal/core/services"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for AuthN service tests
// See mocks.go for shared mock definitions and helper functions

func TestLogin(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	email := "user@example.com"

	tests := []struct {
		name      string
		setup     func(*MockUserRepository, *MockCredentialRepository, *MockPasswordHasher, *MockTokenGenerator, *MockEventProducer)
		wantError bool
		errMsg    string
	}{
		{
			name: "successful login",
			setup: func(mockUser *MockUserRepository, mockCred *MockCredentialRepository, mockHasher *MockPasswordHasher, mockToken *MockTokenGenerator, mockEvent *MockEventProducer) {
				user := db.User{
					ID:       testPgUUID(userID.String()),
					Email:    email,
					TenantID: testPgUUID(tenantID.String()),
					Status:   db.UserStatusACTIVE,
				}
				credential := db.Credential{UserID: testPgUUID(userID.String()), PasswordHash: "hashed"}
				
				mockUser.On("GetUserByEmail", ctx, email, tenantID).Return(user, nil)
				mockCred.On("GetCredentialByUserID", ctx, userID).Return(credential, nil)
				mockHasher.On("Compare", "hashed", "pass123").Return(nil)
				mockToken.On("GenerateAccessToken", userID.String(), tenantID.String(), email).Return("access", nil)
				mockToken.On("GenerateRefreshToken", userID.String(), tenantID.String(), email).Return("refresh", nil)
				mockUser.On("UpdateLastLogin", ctx, userID).Return(nil)
				mockEvent.On("PublishUserLogin", ctx, mock.AnythingOfType("events.UserLoginEvent")).Return(nil)
			},
			wantError: false,
		},
		{
			name: "inactive user",
			setup: func(mockUser *MockUserRepository, mockCred *MockCredentialRepository, mockHasher *MockPasswordHasher, mockToken *MockTokenGenerator, mockEvent *MockEventProducer) {
				user := db.User{Status: db.UserStatusSUSPENDED, TenantID: testPgUUID(tenantID.String())}
				mockUser.On("GetUserByEmail", ctx, email, tenantID).Return(user, nil)
			},
			wantError: true,
			errMsg:    "not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUser := new(MockUserRepository)
			mockCred := new(MockCredentialRepository)
			mockHasher := new(MockPasswordHasher)
			mockToken := new(MockTokenGenerator)
			mockEvent := new(MockEventProducer)

			tt.setup(mockUser, mockCred, mockHasher, mockToken, mockEvent)

			service := services.NewAuthnService(mockUser, mockCred, mockHasher, mockToken, new(MockTokenValidator), new(MockNotificationProducer), mockEvent)
			_, _, _, err := service.Login(ctx, email, "pass123", tenantID.String())

			if tt.wantError {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
			mockUser.AssertExpectations(t)
			mockCred.AssertExpectations(t)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockValidator := new(MockTokenValidator)
		mockGenerator := new(MockTokenGenerator)
		
		claims := &token.Claims{UserID: uuid.New().String(), TenantID: uuid.New().String(), Email: "test@example.com"}
		mockValidator.On("ValidateToken", "refresh-token").Return(claims, nil)
		mockGenerator.On("GenerateAccessToken", claims.UserID, claims.TenantID, claims.Email).Return("new-access", nil)

		service := services.NewAuthnService(new(MockUserRepository), new(MockCredentialRepository), new(MockPasswordHasher), mockGenerator, mockValidator, new(MockNotificationProducer), new(MockEventProducer))
		token, err := service.RefreshToken(ctx, "refresh-token")

		require.NoError(t, err)
		assert.Equal(t, "new-access", token)
	})
}

func TestValidateToken(t *testing.T) {
	ctx := context.Background()

	t.Run("valid token", func(t *testing.T) {
		mockValidator := new(MockTokenValidator)
		claims := &token.Claims{UserID: uuid.New().String()}
		mockValidator.On("ValidateToken", "token").Return(claims, nil)

		service := services.NewAuthnService(new(MockUserRepository), new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), mockValidator, new(MockNotificationProducer), new(MockEventProducer))
		result, err := service.ValidateToken(ctx, "token")

		require.NoError(t, err)
		assert.Equal(t, claims.UserID, result.UserID)
	})
}

func TestLogout(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)
		mockCred.On("DeleteAllUserTokens", ctx, userID).Return(nil)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.Logout(ctx, userID.String())

		require.NoError(t, err)
	})
}

// TestGeneratePasswordSetupToken tests Step 11 of the onboarding flow
func TestGeneratePasswordSetupToken(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	email := "admin@example.com"

	t.Run("success - generates token and sends email", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)
		mockNotif := new(MockNotificationProducer)

		// Expect token to be stored in database
		mockCred.On("CreateToken", ctx, mock.MatchedBy(func(params db.CreateTokenParams) bool {
			return params.UserID.Bytes == testPgUUID(userID.String()).Bytes &&
				params.Scope == db.TokenScopePASSWORDRESET &&
				params.Expiry.Valid
		})).Return(nil)

		// Expect email notification to be sent
		mockNotif.On("SendEmail", ctx, email, mock.Anything, mock.Anything).Return(nil)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), mockNotif, new(MockEventProducer))
		token, err := service.GeneratePasswordSetupToken(ctx, userID.String(), tenantID.String(), email)

		require.NoError(t, err)
		assert.NotEmpty(t, token, "Token should not be empty")
		mockCred.AssertExpectations(t)
		mockNotif.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		service := services.NewAuthnService(new(MockUserRepository), new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		_, err := service.GeneratePasswordSetupToken(ctx, "invalid-uuid", tenantID.String(), email)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")
	})

	t.Run("error - failed to store token", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)
		mockCred.On("CreateToken", ctx, mock.Anything).Return(assert.AnError)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		_, err := service.GeneratePasswordSetupToken(ctx, userID.String(), tenantID.String(), email)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store token")
	})
}

// TestSetInitialPassword tests Step 12 of the onboarding flow
func TestSetInitialPassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	setupToken := "dGVzdC10b2tlbi0xMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTA=" // base64 encoded 32 bytes

	t.Run("success - sets password and activates user", func(t *testing.T) {
		mockUser := new(MockUserRepository)
		mockCred := new(MockCredentialRepository)
		mockHasher := new(MockPasswordHasher)
		mockEvent := new(MockEventProducer)

		// Mock token retrieval (valid token, not expired)
		futureTime := time.Now().Add(24 * time.Hour)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{
			UserID: testPgUUID(userID.String()),
			Expiry: pgtype.Timestamptz{Time: futureTime, Valid: true},
			Scope:  db.TokenScopePASSWORDRESET,
		}, nil)

		// Mock password hashing
		mockHasher.On("Hash", "newPassword123").Return("hashed_password", nil)

		// Mock credential creation
		mockCred.On("CreateCredential", ctx, mock.MatchedBy(func(params db.CreateCredentialParams) bool {
			return params.UserID.Bytes == testPgUUID(userID.String()).Bytes &&
				params.PasswordHash == "hashed_password"
		})).Return(nil)

		// Mock user retrieval for tenant info
		mockUser.On("GetUserByID", ctx, userID).Return(db.User{
			ID:       testPgUUID(userID.String()),
			TenantID: testPgUUID(tenantID.String()),
			Status:   db.UserStatusPENDINGSETUP,
		}, nil)

		// Mock status update - PENDING_SETUP -> ACTIVE (critical for onboarding)
		mockUser.On("UpdateUserStatus", ctx, userID, db.UserStatusACTIVE).Return(nil)

		// Mock event publishing
		mockEvent.On("PublishUserStatusChanged", ctx, mock.MatchedBy(func(event interface{}) bool {
			return true // Accept any status changed event
		})).Return(nil)

		mockEvent.On("PublishPasswordChanged", ctx, mock.MatchedBy(func(event interface{}) bool {
			return true // Accept any password changed event
		})).Return(nil)

		// Mock token deletion
		mockCred.On("DeleteToken", ctx, mock.Anything).Return(nil)

		service := services.NewAuthnService(mockUser, mockCred, mockHasher, new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), mockEvent)
		err := service.SetInitialPassword(ctx, setupToken, "newPassword123")

		require.NoError(t, err)
		mockUser.AssertExpectations(t)
		mockCred.AssertExpectations(t)
		mockHasher.AssertExpectations(t)
	})

	t.Run("error - invalid token format", func(t *testing.T) {
		service := services.NewAuthnService(new(MockUserRepository), new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.SetInitialPassword(ctx, "invalid-base64!", "newPassword123")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")
	})

	t.Run("error - token not found", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{}, assert.AnError)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.SetInitialPassword(ctx, setupToken, "newPassword123")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired token")
	})

	t.Run("error - token expired", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)
		
		// Return expired token (1 hour ago)
		pastTime := time.Now().Add(-1 * time.Hour)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{
			UserID: testPgUUID(userID.String()),
			Expiry: pgtype.Timestamptz{Time: pastTime, Valid: true},
			Scope:  db.TokenScopePASSWORDRESET,
		}, nil)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.SetInitialPassword(ctx, setupToken, "newPassword123")
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

// TestRegisterInvitedUser tests the invited user registration flow
func TestRegisterInvitedUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	invitationToken := "aW52aXRhdGlvbi10b2tlbi0xMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTA=" // base64 encoded 32 bytes
	email := "invited@example.com"
	fullName := "John Doe"
	password := "SecurePass123"

	t.Run("success - user registers and gets logged in automatically", func(t *testing.T) {
		mockUser := new(MockUserRepository)
		mockCred := new(MockCredentialRepository)
		mockHasher := new(MockPasswordHasher)
		mockToken := new(MockTokenGenerator)

		// Mock token retrieval (invitation token, valid and not expired)
		futureTime := time.Now().Add(24 * time.Hour)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{
			UserID: testPgUUID(userID.String()),
			Expiry: pgtype.Timestamptz{Time: futureTime, Valid: true},
			Scope:  db.TokenScopePASSWORDRESET, // Invitation uses same scope
		}, nil)

		// Mock password hashing
		mockHasher.On("Hash", password).Return("hashed_password", nil)

		// Mock credential creation
		mockCred.On("CreateCredential", ctx, mock.MatchedBy(func(params db.CreateCredentialParams) bool {
			return params.UserID.Bytes == testPgUUID(userID.String()).Bytes
		})).Return(nil)

		// Mock user activation - PENDING_INVITE -> ACTIVE
		mockUser.On("ActivateInvitedUser", ctx, mock.MatchedBy(func(params db.ActivateInvitedUserParams) bool {
			return params.ID.Bytes == testPgUUID(userID.String()).Bytes &&
				params.FullName == fullName
		})).Return(nil)

		// Mock user retrieval (now activated)
		mockUser.On("GetUserByID", ctx, userID).Return(db.User{
			ID:       testPgUUID(userID.String()),
			TenantID: testPgUUID(tenantID.String()),
			Email:    email,
			FullName: fullName,
			Status:   db.UserStatusACTIVE, // Now ACTIVE after activation
		}, nil)

		// Mock token generation (auto-login)
		mockToken.On("GenerateAccessToken", userID.String(), tenantID.String(), email).Return("access_token", nil)
		mockToken.On("GenerateRefreshToken", userID.String(), tenantID.String(), email).Return("refresh_token", nil)

		// Mock last login update
		mockUser.On("UpdateLastLogin", ctx, userID).Return(nil)

		// Mock token deletion (invitation token consumed)
		mockCred.On("DeleteToken", ctx, mock.Anything).Return(nil)

		service := services.NewAuthnService(mockUser, mockCred, mockHasher, mockToken, new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		accessToken, refreshToken, user, err := service.RegisterInvitedUser(ctx, invitationToken, fullName, password)

		require.NoError(t, err)
		assert.Equal(t, "access_token", accessToken, "Access token should be returned for auto-login")
		assert.Equal(t, "refresh_token", refreshToken, "Refresh token should be returned for auto-login")
		assert.Equal(t, email, user.Email)
		assert.Equal(t, db.UserStatusACTIVE, user.Status, "User should be ACTIVE after registration")
		
		mockUser.AssertExpectations(t)
		mockCred.AssertExpectations(t)
		mockHasher.AssertExpectations(t)
		mockToken.AssertExpectations(t)
	})

	t.Run("error - invalid invitation token format", func(t *testing.T) {
		service := services.NewAuthnService(new(MockUserRepository), new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		_, _, _, err := service.RegisterInvitedUser(ctx, "invalid-token!", fullName, password)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")
	})

	t.Run("error - invitation not found", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{}, assert.AnError)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		_, _, _, err := service.RegisterInvitedUser(ctx, invitationToken, fullName, password)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired invitation")
	})

	t.Run("error - invitation expired", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)
		
		// Return expired invitation (1 hour ago)
		pastTime := time.Now().Add(-1 * time.Hour)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{
			UserID: testPgUUID(userID.String()),
			Expiry: pgtype.Timestamptz{Time: pastTime, Valid: true},
			Scope:  db.TokenScopePASSWORDRESET,
		}, nil)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		_, _, _, err := service.RegisterInvitedUser(ctx, invitationToken, fullName, password)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

// TestForgotPassword tests password recovery initiation
func TestForgotPassword(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	email := "user@example.com"
	userID := uuid.New()

	t.Run("success - initiates password reset", func(t *testing.T) {
		mockUser := new(MockUserRepository)
		mockCred := new(MockCredentialRepository)
		mockNotif := new(MockNotificationProducer)

		// Mock user retrieval
		mockUser.On("GetUserByEmail", ctx, email, tenantID).Return(db.User{
			ID:       testPgUUID(userID.String()),
			Email:    email,
			TenantID: testPgUUID(tenantID.String()),
		}, nil)

		// Mock token creation (called by GeneratePasswordSetupToken)
		mockCred.On("CreateToken", ctx, mock.Anything).Return(nil)

		// Mock email notification
		mockNotif.On("SendEmail", ctx, email, mock.Anything, mock.Anything).Return(nil)

		service := services.NewAuthnService(mockUser, mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), mockNotif, new(MockEventProducer))
		err := service.ForgotPassword(ctx, email, tenantID.String())

		require.NoError(t, err)
		mockUser.AssertExpectations(t)
		mockCred.AssertExpectations(t)
		mockNotif.AssertExpectations(t)
	})

	t.Run("success - user not found (security - don't reveal)", func(t *testing.T) {
		mockUser := new(MockUserRepository)

		// User doesn't exist - should still return success (security best practice)
		mockUser.On("GetUserByEmail", ctx, email, tenantID).Return(db.User{}, assert.AnError)

		service := services.NewAuthnService(mockUser, new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.ForgotPassword(ctx, email, tenantID.String())

		require.NoError(t, err, "Should not reveal if user exists")
		mockUser.AssertExpectations(t)
	})

	t.Run("error - invalid tenant ID", func(t *testing.T) {
		service := services.NewAuthnService(new(MockUserRepository), new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.ForgotPassword(ctx, email, "invalid-uuid")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tenant ID")
	})
}

// TestResetPassword tests password reset with token
func TestResetPassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	resetToken := "cmVzZXQtdG9rZW4tMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=" // base64 encoded

	t.Run("success - resets password with valid token", func(t *testing.T) {
		mockUser := new(MockUserRepository)
		mockCred := new(MockCredentialRepository)
		mockHasher := new(MockPasswordHasher)
		mockEvent := new(MockEventProducer)

		// Mock token retrieval
		futureTime := time.Now().Add(24 * time.Hour)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{
			UserID: testPgUUID(userID.String()),
			Expiry: pgtype.Timestamptz{Time: futureTime, Valid: true},
			Scope:  db.TokenScopePASSWORDRESET,
		}, nil)

		// Mock password hashing
		mockHasher.On("Hash", "NewSecurePassword123").Return("new_hashed_password", nil)

		// Mock credential creation
		mockCred.On("CreateCredential", ctx, mock.Anything).Return(nil)

		// Mock user retrieval
		mockUser.On("GetUserByID", ctx, userID).Return(db.User{
			ID:       testPgUUID(userID.String()),
			TenantID: testPgUUID(tenantID.String()),
			Status:   db.UserStatusACTIVE, // User already active, just resetting password
		}, nil)

		// Mock status update
		mockUser.On("UpdateUserStatus", ctx, userID, db.UserStatusACTIVE).Return(nil)

		// Mock event publishing
		mockEvent.On("PublishUserStatusChanged", ctx, mock.Anything).Return(nil)
		mockEvent.On("PublishPasswordChanged", ctx, mock.Anything).Return(nil)

		// Mock token deletion
		mockCred.On("DeleteToken", ctx, mock.Anything).Return(nil)

		service := services.NewAuthnService(mockUser, mockCred, mockHasher, new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), mockEvent)
		err := service.ResetPassword(ctx, resetToken, "NewSecurePassword123")

		require.NoError(t, err)
		mockUser.AssertExpectations(t)
		mockCred.AssertExpectations(t)
		mockHasher.AssertExpectations(t)
	})

	t.Run("error - invalid reset token", func(t *testing.T) {
		service := services.NewAuthnService(new(MockUserRepository), new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.ResetPassword(ctx, "invalid-base64!", "NewPassword123")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")
	})

	t.Run("error - expired reset token", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)

		// Return expired token
		pastTime := time.Now().Add(-1 * time.Hour)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{
			UserID: testPgUUID(userID.String()),
			Expiry: pgtype.Timestamptz{Time: pastTime, Valid: true},
			Scope:  db.TokenScopePASSWORDRESET,
		}, nil)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.ResetPassword(ctx, resetToken, "NewPassword123")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

// TestConfirmPassword tests password confirmation for sensitive actions
func TestConfirmPassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success - correct password confirmed", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)
		mockHasher := new(MockPasswordHasher)

		// Mock credential retrieval
		mockCred.On("GetCredentialByUserID", ctx, userID).Return(db.Credential{
			UserID:       testPgUUID(userID.String()),
			PasswordHash: "stored_hashed_password",
		}, nil)

		// Mock password comparison (correct password)
		mockHasher.On("Compare", "stored_hashed_password", "CorrectPassword123").Return(nil)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, mockHasher, new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.ConfirmPassword(ctx, userID.String(), "CorrectPassword123")

		require.NoError(t, err)
		mockCred.AssertExpectations(t)
		mockHasher.AssertExpectations(t)
	})

	t.Run("error - incorrect password", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)
		mockHasher := new(MockPasswordHasher)

		mockCred.On("GetCredentialByUserID", ctx, userID).Return(db.Credential{
			UserID:       testPgUUID(userID.String()),
			PasswordHash: "stored_hashed_password",
		}, nil)

		// Mock password comparison (incorrect password)
		mockHasher.On("Compare", "stored_hashed_password", "WrongPassword123").Return(assert.AnError)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, mockHasher, new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.ConfirmPassword(ctx, userID.String(), "WrongPassword123")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect password")
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		service := services.NewAuthnService(new(MockUserRepository), new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.ConfirmPassword(ctx, "invalid-uuid", "SomePassword123")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")
	})

	t.Run("error - credential not found", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)

		mockCred.On("GetCredentialByUserID", ctx, userID).Return(db.Credential{}, assert.AnError)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.ConfirmPassword(ctx, userID.String(), "SomePassword123")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "credential not found")
	})
}

// TestRequestEmailVerification tests email verification request flow
func TestRequestEmailVerification(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	t.Run("success - sends verification email", func(t *testing.T) {
		mockUser := new(MockUserRepository)
		mockCred := new(MockCredentialRepository)
		mockNotif := new(MockNotificationProducer)

		// Mock user retrieval (email not verified yet)
		mockUser.On("GetUserByID", ctx, userID).Return(db.User{
			ID:            testPgUUID(userID.String()),
			TenantID:      testPgUUID(tenantID.String()),
			Email:         "user@example.com",
			EmailVerified: false,
		}, nil)

		// Mock token creation
		mockCred.On("CreateToken", ctx, mock.MatchedBy(func(params db.CreateTokenParams) bool {
			return params.UserID.Bytes == testPgUUID(userID.String()).Bytes &&
				params.Scope == db.TokenScopeEMAILVERIFICATION &&
				params.Expiry.Valid
		})).Return(nil)

		// Mock email notification
		mockNotif.On("SendEmail", ctx, "user@example.com", mock.Anything, mock.Anything).Return(nil)

		service := services.NewAuthnService(mockUser, mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), mockNotif, new(MockEventProducer))
		err := service.RequestEmailVerification(ctx, userID.String())

		require.NoError(t, err)
		mockUser.AssertExpectations(t)
		mockCred.AssertExpectations(t)
		mockNotif.AssertExpectations(t)
	})

	t.Run("error - invalid user ID", func(t *testing.T) {
		service := services.NewAuthnService(new(MockUserRepository), new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.RequestEmailVerification(ctx, "invalid-uuid")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")
	})

	t.Run("error - user not found", func(t *testing.T) {
		mockUser := new(MockUserRepository)
		mockUser.On("GetUserByID", ctx, userID).Return(db.User{}, assert.AnError)

		service := services.NewAuthnService(mockUser, new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.RequestEmailVerification(ctx, userID.String())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("error - email already verified", func(t *testing.T) {
		mockUser := new(MockUserRepository)

		// User with verified email
		mockUser.On("GetUserByID", ctx, userID).Return(db.User{
			ID:            testPgUUID(userID.String()),
			Email:         "user@example.com",
			EmailVerified: true,
		}, nil)

		service := services.NewAuthnService(mockUser, new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.RequestEmailVerification(ctx, userID.String())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "email already verified")
	})

	t.Run("error - failed to send email", func(t *testing.T) {
		mockUser := new(MockUserRepository)
		mockCred := new(MockCredentialRepository)
		mockNotif := new(MockNotificationProducer)

		mockUser.On("GetUserByID", ctx, userID).Return(db.User{
			ID:            testPgUUID(userID.String()),
			TenantID:      testPgUUID(tenantID.String()),
			Email:         "user@example.com",
			EmailVerified: false,
		}, nil)

		mockCred.On("CreateToken", ctx, mock.Anything).Return(nil)
		mockNotif.On("SendEmail", ctx, "user@example.com", mock.Anything, mock.Anything).Return(assert.AnError)

		service := services.NewAuthnService(mockUser, mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), mockNotif, new(MockEventProducer))
		err := service.RequestEmailVerification(ctx, userID.String())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send verification email")
	})
}

// TestVerifyEmail tests email verification with token
func TestVerifyEmail(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	verificationToken := "dmVyaWZpY2F0aW9uLXRva2VuLTEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEy" // base64 encoded

	t.Run("success - verifies email with valid token", func(t *testing.T) {
		mockUser := new(MockUserRepository)
		mockCred := new(MockCredentialRepository)
		mockEvent := new(MockEventProducer)

		// Mock token retrieval
		futureTime := time.Now().Add(24 * time.Hour)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{
			UserID: testPgUUID(userID.String()),
			Expiry: pgtype.Timestamptz{Time: futureTime, Valid: true},
			Scope:  db.TokenScopeEMAILVERIFICATION,
		}, nil)

		// Mock email verified update
		mockUser.On("UpdateEmailVerified", ctx, userID, true).Return(nil)

		// Mock user retrieval
		mockUser.On("GetUserByID", ctx, userID).Return(db.User{
			ID:       testPgUUID(userID.String()),
			TenantID: testPgUUID(tenantID.String()),
			Email:    "user@example.com",
		}, nil)

		// Mock token deletion
		mockCred.On("DeleteToken", ctx, mock.Anything).Return(nil)

		// Mock event publishing
		mockEvent.On("PublishUserUpdated", ctx, mock.Anything).Return(nil)

		service := services.NewAuthnService(mockUser, mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), mockEvent)
		err := service.VerifyEmail(ctx, verificationToken)

		require.NoError(t, err)
		mockCred.AssertExpectations(t)
		mockUser.AssertExpectations(t)
		mockEvent.AssertExpectations(t)
	})

	t.Run("error - invalid token format", func(t *testing.T) {
		service := services.NewAuthnService(new(MockUserRepository), new(MockCredentialRepository), new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.VerifyEmail(ctx, "invalid-base64!")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")
	})

	t.Run("error - token not found", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{}, assert.AnError)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.VerifyEmail(ctx, verificationToken)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired verification token")
	})

	t.Run("error - token expired", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)

		// Return expired token
		pastTime := time.Now().Add(-1 * time.Hour)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{
			UserID: testPgUUID(userID.String()),
			Expiry: pgtype.Timestamptz{Time: pastTime, Valid: true},
			Scope:  db.TokenScopeEMAILVERIFICATION,
		}, nil)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.VerifyEmail(ctx, verificationToken)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "verification token has expired")
	})

	t.Run("error - invalid token scope", func(t *testing.T) {
		mockCred := new(MockCredentialRepository)

		// Return token with wrong scope
		futureTime := time.Now().Add(24 * time.Hour)
		mockCred.On("GetToken", ctx, mock.Anything).Return(db.Token{
			UserID: testPgUUID(userID.String()),
			Expiry: pgtype.Timestamptz{Time: futureTime, Valid: true},
			Scope:  db.TokenScopePASSWORDRESET, // Wrong scope
		}, nil)

		service := services.NewAuthnService(new(MockUserRepository), mockCred, new(MockPasswordHasher), new(MockTokenGenerator), new(MockTokenValidator), new(MockNotificationProducer), new(MockEventProducer))
		err := service.VerifyEmail(ctx, verificationToken)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token scope")
	})
}
