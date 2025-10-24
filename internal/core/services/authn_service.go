package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"cp_service/internal/adapters/database/db"
	"cp_service/internal/adapters/password"
	"cp_service/internal/adapters/token"
	"cp_service/internal/core/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// NotificationProducer defines the interface for sending notifications
type NotificationProducer interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// SAMLProvider defines the interface for SAML authentication (COMMENTED OUT FOR NOW)
// type SAMLProvider interface {
// 	ValidateSAMLResponse(ctx context.Context, samlResponse string) (email string, attributes map[string][]string, err error)
// 	GetSAMLAuthURL(ctx context.Context, relayState string) (string, error)
// }

// AuthnService handles authentication business logic
type AuthnService struct {
	userRepo         repository.UserRepository
	credentialRepo   repository.CredentialRepository
	passwordHasher   password.Hasher
	tokenGenerator   token.Generator
	tokenValidator   token.Validator
	notificationProd NotificationProducer
	// samlProvider     SAMLProvider // COMMENTED OUT
}

// NewAuthnService creates a new authentication service
func NewAuthnService(
	userRepo repository.UserRepository,
	credentialRepo repository.CredentialRepository,
	passwordHasher password.Hasher,
	tokenGen token.Generator,
	tokenVal token.Validator,
	notificationProd NotificationProducer,
	// samlProvider SAMLProvider, // COMMENTED OUT
) *AuthnService {
	return &AuthnService{
		userRepo:         userRepo,
		credentialRepo:   credentialRepo,
		passwordHasher:   passwordHasher,
		tokenGenerator:   tokenGen,
		tokenValidator:   tokenVal,
		notificationProd: notificationProd,
		// samlProvider:     samlProvider, // COMMENTED OUT
	}
}

// GeneratePasswordSetupToken generates a one-time password setup token (Step 12 in onboarding)
func (s *AuthnService) GeneratePasswordSetupToken(ctx context.Context, userID, tenantID, email string) (string, error) {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %w", err)
	}

	// Generate random token (32 bytes = 256 bits)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode token to base64 for URL safety
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash the token before storing (SHA-256)
	hash := sha256.Sum256(tokenBytes)

	// Store token hash in database with 24-hour expiry
	expiry := time.Now().Add(24 * time.Hour)
	if err := s.credentialRepo.CreateToken(ctx, db.CreateTokenParams{
		Hash:   hash[:],
		UserID: pgUUID(userID),
		Expiry: pgtype.Timestamptz{Time: expiry, Valid: true},
		Scope:  db.TokenScopePASSWORDRESET,
	}); err != nil {
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	// Send email notification with setup link
	setupURL := fmt.Sprintf("https://app.yourcompany.com/auth/setup-password?token=%s", token)
	subject := "Set Your Password - Welcome to the Platform"
	body := fmt.Sprintf(`
Hello,

Welcome! Please click the link below to set your password and activate your account:

%s

This link will expire in 24 hours.

If you didn't request this, please ignore this email.

Best regards,
Your Platform Team
`, setupURL)

	if err := s.notificationProd.SendEmail(ctx, email, subject, body); err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	return token, nil
}

// SetInitialPassword sets the password for a user using a setup token (Step 14 in onboarding)
func (s *AuthnService) SetInitialPassword(ctx context.Context, setupToken, newPassword string) error {
	// Decode and hash the token
	tokenBytes, err := base64.URLEncoding.DecodeString(setupToken)
	if err != nil {
		return fmt.Errorf("invalid token format: %w", err)
	}

	tokenHash := sha256.Sum256(tokenBytes)

	// Retrieve token from database
	storedToken, err := s.credentialRepo.GetToken(ctx, tokenHash[:])
	if err != nil {
		return fmt.Errorf("invalid or expired token: %w", err)
	}

	// Check if token is expired
	if storedToken.Expiry.Time.Before(time.Now()) {
		return fmt.Errorf("token has expired")
	}

	// Hash the new password
	passwordHash, err := s.passwordHasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create credential
	userID, _ := uuid.FromBytes(storedToken.UserID.Bytes[:])
	if err := s.credentialRepo.CreateCredential(ctx, db.CreateCredentialParams{
		UserID:       storedToken.UserID,
		PasswordHash: passwordHash,
	}); err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	// Update user status to ACTIVE
	if err := s.userRepo.UpdateUserStatus(ctx, userID, db.UserStatusACTIVE); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// Delete the used token
	if err := s.credentialRepo.DeleteToken(ctx, tokenHash[:]); err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

// Login authenticates a user with email/password
func (s *AuthnService) Login(ctx context.Context, email, password, tenantIdentifier string) (accessToken, refreshToken string, user db.User, err error) {
	// Parse tenant ID
	tenantID, err := uuid.Parse(tenantIdentifier)
	if err != nil {
		return "", "", db.User{}, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Get user by email and tenant
	user, err = s.userRepo.GetUserByEmail(ctx, email, tenantID)
	if err != nil {
		return "", "", db.User{}, fmt.Errorf("invalid credentials: %w", err)
	}

	// Check if user is active
	if user.Status != db.UserStatusACTIVE {
		return "", "", db.User{}, fmt.Errorf("user account is not active")
	}

	// Get credential
	userID, _ := uuid.FromBytes(user.ID.Bytes[:])
	credential, err := s.credentialRepo.GetCredentialByUserID(ctx, userID)
	if err != nil {
		return "", "", db.User{}, fmt.Errorf("invalid credentials: %w", err)
	}

	// Verify password
	if err := s.passwordHasher.CheckPassword(password, credential.PasswordHash); err != nil {
		return "", "", db.User{}, fmt.Errorf("invalid credentials: %w", err)
	}

	// Generate JWT tokens
	accessToken, err = s.tokenGenerator.GenerateAccessToken(userID.String(), tenantIdentifier, email)
	if err != nil {
		return "", "", db.User{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = s.tokenGenerator.GenerateRefreshToken(userID.String(), tenantIdentifier)
	if err != nil {
		return "", "", db.User{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login time
	if err := s.userRepo.UpdateLastLogin(ctx, userID); err != nil {
		// Non-critical error, log but don't fail
		fmt.Printf("warning: failed to update last login: %v\n", err)
	}

	return accessToken, refreshToken, user, nil
}

// LoginWithSAML authenticates a user using SAML SSO (COMMENTED OUT FOR NOW)
// func (s *AuthnService) LoginWithSAML(ctx context.Context, samlResponse, tenantIdentifier string) (accessToken, refreshToken string, user db.User, err error) {
// 	// Validate SAML response
// 	email, attributes, err := s.samlProvider.ValidateSAMLResponse(ctx, samlResponse)
// 	if err != nil {
// 		return "", "", db.User{}, fmt.Errorf("invalid SAML response: %w", err)
// 	}
//
// 	// Parse tenant ID
// 	tenantID, err := uuid.Parse(tenantIdentifier)
// 	if err != nil {
// 		return "", "", db.User{}, fmt.Errorf("invalid tenant ID: %w", err)
// 	}
//
// 	// Get or create user
// 	user, err = s.userRepo.GetUserByEmail(ctx, email, tenantID)
// 	if err != nil {
// 		// User doesn't exist - create from SAML attributes
// 		fullName := extractSAMLAttribute(attributes, "displayName", email)
//
// 		user, err = s.userRepo.CreateUser(ctx, db.CreateUserParams{
// 			FullName: fullName,
// 			Email:    email,
// 			TenantID: pgUUID(tenantIdentifier),
// 			Status:   db.UserStatusACTIVE,
// 		})
// 		if err != nil {
// 			return "", "", db.User{}, fmt.Errorf("failed to create user from SAML: %w", err)
// 		}
// 	}
//
// 	// Check if user is active
// 	if user.Status != db.UserStatusACTIVE {
// 		return "", "", db.User{}, fmt.Errorf("user account is not active")
// 	}
//
// 	// Generate JWT tokens
// 	userID, _ := uuid.FromBytes(user.ID.Bytes[:])
// 	accessToken, err = s.tokenGenerator.GenerateAccessToken(userID.String(), tenantIdentifier, email)
// 	if err != nil {
// 		return "", "", db.User{}, fmt.Errorf("failed to generate access token: %w", err)
// 	}
//
// 	refreshToken, err = s.tokenGenerator.GenerateRefreshToken(userID.String(), tenantIdentifier)
// 	if err != nil {
// 		return "", "", db.User{}, fmt.Errorf("failed to generate refresh token: %w", err)
// 	}
//
// 	// Update last login time
// 	if err := s.userRepo.UpdateLastLogin(ctx, userID); err != nil {
// 		fmt.Printf("warning: failed to update last login: %v\n", err)
// 	}
//
// 	return accessToken, refreshToken, user, nil
// }

// GetSAMLAuthURL returns the SAML SSO authentication URL (COMMENTED OUT FOR NOW)
// func (s *AuthnService) GetSAMLAuthURL(ctx context.Context, relayState string) (string, error) {
// 	return s.samlProvider.GetSAMLAuthURL(ctx, relayState)
// }

// RefreshToken generates a new access token using a refresh token
func (s *AuthnService) RefreshToken(ctx context.Context, refreshTokenStr string) (accessToken string, err error) {
	// Validate refresh token
	claims, err := s.tokenValidator.ValidateToken(refreshTokenStr)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new access token
	accessToken, err = s.tokenGenerator.GenerateAccessToken(claims.UserID, claims.TenantID, claims.Email)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthnService) ValidateToken(ctx context.Context, tokenStr string) (*token.Claims, error) {
	claims, err := s.tokenValidator.ValidateToken(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}

// ForgotPassword initiates the password reset process
func (s *AuthnService) ForgotPassword(ctx context.Context, email, tenantIdentifier string) error {
	tenantID, err := uuid.Parse(tenantIdentifier)
	if err != nil {
		return fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetUserByEmail(ctx, email, tenantID)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// Generate password reset token
	userID, _ := uuid.FromBytes(user.ID.Bytes[:])
	_, err = s.GeneratePasswordSetupToken(ctx, userID.String(), tenantIdentifier, email)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	return nil
}

// ResetPassword resets a user's password using a reset token
func (s *AuthnService) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	// Use the same logic as SetInitialPassword
	return s.SetInitialPassword(ctx, resetToken, newPassword)
}

// RegisterInvitedUser allows an invited user to register and set their password
// Returns access token and refresh token to automatically log them in (as per proto definition)
func (s *AuthnService) RegisterInvitedUser(ctx context.Context, invitationToken, fullName, password string) (accessToken, refreshToken string, user db.User, err error) {
	// Decode and hash the token
	tokenBytes, decodeErr := base64.URLEncoding.DecodeString(invitationToken)
	if decodeErr != nil {
		return "", "", db.User{}, fmt.Errorf("invalid token format: %w", decodeErr)
	}

	tokenHash := sha256.Sum256(tokenBytes)

	// Retrieve token
	storedToken, err := s.credentialRepo.GetToken(ctx, tokenHash[:])
	if err != nil {
		return "", "", db.User{}, fmt.Errorf("invalid or expired invitation: %w", err)
	}

	// Check expiry
	if storedToken.Expiry.Time.Before(time.Now()) {
		return "", "", db.User{}, fmt.Errorf("invitation has expired")
	}

	// Hash password
	passwordHash, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return "", "", db.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create credential
	userID, _ := uuid.FromBytes(storedToken.UserID.Bytes[:])
	if err := s.credentialRepo.CreateCredential(ctx, db.CreateCredentialParams{
		UserID:       storedToken.UserID,
		PasswordHash: passwordHash,
	}); err != nil {
		return "", "", db.User{}, fmt.Errorf("failed to create credential: %w", err)
	}

	// Activate invited user (status: PENDING_INVITE -> USER_ACTIVE)
	if err := s.userRepo.ActivateInvitedUser(ctx, db.ActivateInvitedUserParams{
		ID:       storedToken.UserID,
		FullName: fullName,
	}); err != nil {
		return "", "", db.User{}, fmt.Errorf("failed to activate user: %w", err)
	}

	// Get the updated user with tenant info
	user, err = s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", db.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate JWT tokens (AUTOMATICALLY LOG THEM IN - admin already set their role)
	tenantID, _ := uuid.FromBytes(user.TenantID.Bytes[:])
	accessToken, err = s.tokenGenerator.GenerateAccessToken(userID.String(), tenantID.String(), user.Email)
	if err != nil {
		return "", "", db.User{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = s.tokenGenerator.GenerateRefreshToken(userID.String(), tenantID.String())
	if err != nil {
		return "", "", db.User{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login time
	if err := s.userRepo.UpdateLastLogin(ctx, userID); err != nil {
		// Non-critical error, log but don't fail
		fmt.Printf("warning: failed to update last login: %v\n", err)
	}

	// Delete token
	if err := s.credentialRepo.DeleteToken(ctx, tokenHash[:]); err != nil {
		return "", "", db.User{}, fmt.Errorf("failed to delete token: %w", err)
	}

	return accessToken, refreshToken, user, nil
}

// Logout invalidates a user's tokens
func (s *AuthnService) Logout(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Delete all user tokens
	if err := s.credentialRepo.DeleteAllUserTokens(ctx, uid); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}

// ConfirmPassword verifies if a password is correct for a user
func (s *AuthnService) ConfirmPassword(ctx context.Context, userID, password string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	credential, err := s.credentialRepo.GetCredentialByUserID(ctx, uid)
	if err != nil {
		return fmt.Errorf("credential not found: %w", err)
	}

	if err := s.passwordHasher.CheckPassword(password, credential.PasswordHash); err != nil {
		return fmt.Errorf("incorrect password: %w", err)
	}

	return nil
}

// Helper function to extract SAML attribute (COMMENTED OUT FOR NOW)
// func extractSAMLAttribute(attributes map[string][]string, key, defaultValue string) string {
// 	if values, ok := attributes[key]; ok && len(values) > 0 {
// 		return values[0]
// 	}
// 	return defaultValue
// }
