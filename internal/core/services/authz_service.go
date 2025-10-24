package services

import (
	"context"
	"fmt"
	"strings"

	"cp_service/internal/adapters/token"
	"cp_service/internal/core/repository"

	"github.com/google/uuid"
)

// OPAClient defines the interface for Open Policy Agent integration
type OPAClient interface {
	Evaluate(ctx context.Context, input map[string]interface{}) (bool, error)
}

// AuthzService handles authorization business logic
type AuthzService struct {
	roleRepo       repository.RoleRepository
	tokenValidator token.Validator
	opaClient      OPAClient
}

// NewAuthzService creates a new authorization service
func NewAuthzService(
	roleRepo repository.RoleRepository,
	tokenValidator token.Validator,
	opaClient OPAClient,
) *AuthzService {
	return &AuthzService{
		roleRepo:       roleRepo,
		tokenValidator: tokenValidator,
		opaClient:      opaClient,
	}
}

// CheckAccess checks if a user has permission to perform an action on a resource
func (s *AuthzService) CheckAccess(ctx context.Context, accessToken, action, resource string) (bool, string, error) {
	// Validate token and extract claims
	claims, err := s.tokenValidator.ValidateToken(accessToken)
	if err != nil {
		return false, "Invalid or expired token", fmt.Errorf("invalid token: %w", err)
	}

	// Parse user ID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return false, "Invalid user ID", fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user's roles
	roles, err := s.roleRepo.ListUserRoles(ctx, userID)
	if err != nil {
		return false, "Failed to retrieve user roles", fmt.Errorf("failed to get user roles: %w", err)
	}

	// Collect all permissions from all roles
	permissions := make(map[string]bool)
	for _, role := range roles {
		for _, perm := range role.Permissions {
			permissions[perm] = true
		}
	}

	// If OPA is configured, use it for complex policy evaluation
	if s.opaClient != nil {
		// Build OPA input
		permList := make([]string, 0, len(permissions))
		for perm := range permissions {
			permList = append(permList, perm)
		}

		input := map[string]interface{}{
			"user": map[string]interface{}{
				"id":          claims.UserID,
				"tenant_id":   claims.TenantID,
				"email":       claims.Email,
				"permissions": permList,
			},
			"action":   action,
			"resource": resource,
		}

		allowed, err := s.opaClient.Evaluate(ctx, input)
		if err != nil {
			// Fall back to simple permission check if OPA fails
			return s.checkSimplePermission(permissions, action, resource)
		}

		if allowed {
			return true, "Access granted by policy", nil
		}
		return false, "Access denied by policy", nil
	}

	// Simple permission-based check
	return s.checkSimplePermission(permissions, action, resource)
}

// checkSimplePermission performs a simple permission-based check
func (s *AuthzService) checkSimplePermission(permissions map[string]bool, action, resource string) (bool, string, error) {
	// Build the required permission string: "resource:action"
	requiredPermission := fmt.Sprintf("%s:%s", resource, action)

	// Check for exact match
	if permissions[requiredPermission] {
		return true, "Access granted", nil
	}

	// Check for wildcard resource permission: "resource:*"
	wildcardResourcePermission := fmt.Sprintf("%s:*", resource)
	if permissions[wildcardResourcePermission] {
		return true, "Access granted (wildcard resource)", nil
	}

	// Check for wildcard action permission: "*:action"
	wildcardActionPermission := fmt.Sprintf("*:%s", action)
	if permissions[wildcardActionPermission] {
		return true, "Access granted (wildcard action)", nil
	}

	// Check for super admin permission: "*:*"
	if permissions["*:*"] {
		return true, "Access granted (super admin)", nil
	}

	return false, "Access denied", nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *AuthzService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user's roles
	roles, err := s.roleRepo.ListUserRoles(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Collect unique permissions
	permissionMap := make(map[string]bool)
	for _, role := range roles {
		for _, perm := range role.Permissions {
			permissionMap[perm] = true
		}
	}

	// Convert to slice
	permissions := make([]string, 0, len(permissionMap))
	for perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (s *AuthzService) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	permissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// Parse permission into resource and action
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid permission format: %s", permission)
	}

	resource := parts[0]
	action := parts[1]

	// Check permissions
	for _, perm := range permissions {
		// Exact match
		if perm == permission {
			return true, nil
		}

		// Wildcard matches
		parts := strings.Split(perm, ":")
		if len(parts) == 2 {
			if (parts[0] == resource && parts[1] == "*") || // resource:*
				(parts[0] == "*" && parts[1] == action) || // *:action
				(parts[0] == "*" && parts[1] == "*") { // *:*
				return true, nil
			}
		}
	}

	return false, nil
}
