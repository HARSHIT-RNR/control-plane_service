package service

import (
	"context"

	"cp_service/internal/adapters/database/db"
	events "cp_service/internal/adapters/kafka"
	"cp_service/internal/adapters/token"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
)

// Helper functions for pgtype conversions
func testPgUUID(id string) pgtype.UUID {
	uid, _ := uuid.Parse(id)
	return pgtype.UUID{Bytes: uid, Valid: true}
}

func testPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// User Repository Mock
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepository) CreateInitialAdmin(ctx context.Context, params db.CreateInitialAdminParams) (db.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string, tenantID uuid.UUID) (db.User, error) {
	args := m.Called(ctx, email, tenantID)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepository) ListUsers(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]db.User, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, params db.UpdateUserParams) (db.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status db.UserStatus) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	args := m.Called(ctx, userID, verified)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ActivateInvitedUser(ctx context.Context, params db.ActivateInvitedUserParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

// Credential Repository Mock
type MockCredentialRepository struct {
	mock.Mock
}

func (m *MockCredentialRepository) CreateCredential(ctx context.Context, params db.CreateCredentialParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockCredentialRepository) GetCredentialByUserID(ctx context.Context, userID uuid.UUID) (db.Credential, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Credential), args.Error(1)
}

func (m *MockCredentialRepository) UpdateCredential(ctx context.Context, params db.UpdateCredentialParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockCredentialRepository) DeleteCredential(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCredentialRepository) CreateToken(ctx context.Context, params db.CreateTokenParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockCredentialRepository) GetToken(ctx context.Context, hash []byte) (db.Token, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(db.Token), args.Error(1)
}

func (m *MockCredentialRepository) DeleteToken(ctx context.Context, hash []byte) error {
	args := m.Called(ctx, hash)
	return args.Error(0)
}

func (m *MockCredentialRepository) DeleteAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Role Repository Mock
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) CreateRole(ctx context.Context, params db.CreateRoleParams) (db.Role, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Role), args.Error(1)
}

func (m *MockRoleRepository) GetRoleByID(ctx context.Context, roleID uuid.UUID) (db.Role, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).(db.Role), args.Error(1)
}

func (m *MockRoleRepository) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]db.Role, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]db.Role), args.Error(1)
}

func (m *MockRoleRepository) UpdateRole(ctx context.Context, params db.UpdateRoleParams) (db.Role, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Role), args.Error(1)
}

func (m *MockRoleRepository) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	args := m.Called(ctx, roleID)
	return args.Error(0)
}

func (m *MockRoleRepository) AssignRoleToUser(ctx context.Context, params db.AssignRoleToUserParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockRoleRepository) RevokeRoleFromUser(ctx context.Context, params db.RevokeRoleFromUserParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockRoleRepository) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]db.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Role), args.Error(1)
}

// Organization Repository Mock
type MockOrganizationRepository struct {
	mock.Mock
}

func (m *MockOrganizationRepository) CreateDepartment(ctx context.Context, params db.CreateDepartmentParams) (db.Department, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Department), args.Error(1)
}

func (m *MockOrganizationRepository) GetDepartmentByID(ctx context.Context, id uuid.UUID) (db.Department, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Department), args.Error(1)
}

func (m *MockOrganizationRepository) ListDepartments(ctx context.Context, tenantID uuid.UUID) ([]db.Department, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]db.Department), args.Error(1)
}

func (m *MockOrganizationRepository) UpdateDepartment(ctx context.Context, params db.UpdateDepartmentParams) (db.Department, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Department), args.Error(1)
}

func (m *MockOrganizationRepository) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrganizationRepository) CreateDesignation(ctx context.Context, params db.CreateDesignationParams) (db.Designation, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Designation), args.Error(1)
}

func (m *MockOrganizationRepository) GetDesignationByID(ctx context.Context, id uuid.UUID) (db.Designation, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Designation), args.Error(1)
}

func (m *MockOrganizationRepository) ListDesignations(ctx context.Context, tenantID uuid.UUID) ([]db.Designation, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]db.Designation), args.Error(1)
}

func (m *MockOrganizationRepository) UpdateDesignation(ctx context.Context, params db.UpdateDesignationParams) (db.Designation, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Designation), args.Error(1)
}

func (m *MockOrganizationRepository) DeleteDesignation(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Password Hasher Mock
type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) Compare(hash, password string) error {
	args := m.Called(hash, password)
	return args.Error(0)
}

// Token Generator Mock
type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockTokenGenerator) GenerateAccessToken(userID, tenantID, email string) (string, error) {
	args := m.Called(userID, tenantID, email)
	return args.String(0), args.Error(1)
}

func (m *MockTokenGenerator) GenerateRefreshToken(userID, tenantID, email string) (string, error) {
	args := m.Called(userID, tenantID, email)
	return args.String(0), args.Error(1)
}

// Token Validator Mock
type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateToken(tokenStr string) (*token.Claims, error) {
	args := m.Called(tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*token.Claims), args.Error(1)
}

// Notification Producer Mock
type MockNotificationProducer struct {
	mock.Mock
}

func (m *MockNotificationProducer) SendEmail(ctx context.Context, to, subject, body string) error {
	args := m.Called(ctx, to, subject, body)
	return args.Error(0)
}

// Event Producer Mock
type MockEventProducer struct {
	mock.Mock
}

func (m *MockEventProducer) PublishUserCreated(ctx context.Context, event events.UserCreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventProducer) PublishUserInvited(ctx context.Context, event events.UserInvitedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventProducer) PublishUserUpdated(ctx context.Context, event events.UserUpdatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventProducer) PublishUserDeleted(ctx context.Context, event events.UserDeletedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventProducer) PublishRoleAssigned(ctx context.Context, event events.RoleAssignedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventProducer) PublishRoleRevoked(ctx context.Context, event events.RoleRevokedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventProducer) PublishUserStatusChanged(ctx context.Context, event events.UserStatusChangedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventProducer) PublishPasswordChanged(ctx context.Context, event events.PasswordChangedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventProducer) PublishUserLogin(ctx context.Context, event events.UserLoginEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// OPA Client Mock
type MockOPAClient struct {
	mock.Mock
}

func (m *MockOPAClient) Evaluate(ctx context.Context, input map[string]interface{}) (bool, error) {
	args := m.Called(ctx, input)
	return args.Bool(0), args.Error(1)
}
