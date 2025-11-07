package handler

import (
	"context"

	"cp_service/internal/adapters/database/db"
	events "cp_service/internal/adapters/kafka"
	"cp_service/internal/adapters/token"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
)

// Shared mock types for handler integration tests

// MockUserRepo implements repository.UserRepository
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (db.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (db.User, error) {
	args := m.Called(ctx, email, tenantID)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepo) ListUsers(ctx context.Context, params db.ListUsersParams) ([]db.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockUserRepo) UpdateUser(ctx context.Context, params db.UpdateUserParams) (db.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepo) ActivateInvitedUser(ctx context.Context, params db.ActivateInvitedUserParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockUserRepo) CreateInitialAdmin(ctx context.Context, params db.CreateInitialAdminParams) (db.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.User), args.Error(1)
}

// MockCredentialRepo implements repository.CredentialRepository
type MockCredentialRepo struct {
	mock.Mock
}

func (m *MockCredentialRepo) CreateCredential(ctx context.Context, params db.CreateCredentialParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockCredentialRepo) GetCredentialByUserID(ctx context.Context, userID uuid.UUID) (db.Credential, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Credential), args.Error(1)
}

func (m *MockCredentialRepo) UpdateCredential(ctx context.Context, params db.UpdateCredentialParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockCredentialRepo) DeleteCredential(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCredentialRepo) CreateToken(ctx context.Context, params db.CreateTokenParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockCredentialRepo) GetToken(ctx context.Context, hash []byte) (db.Token, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(db.Token), args.Error(1)
}

func (m *MockCredentialRepo) DeleteToken(ctx context.Context, hash []byte) error {
	args := m.Called(ctx, hash)
	return args.Error(0)
}

func (m *MockCredentialRepo) DeleteAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockPasswordHash implements password.Hasher
type MockPasswordHash struct {
	mock.Mock
}

func (m *MockPasswordHash) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHash) Hash(password string) (string, error) {
	return m.HashPassword(password)
}

func (m *MockPasswordHash) ComparePassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

func (m *MockPasswordHash) Compare(hashedPassword, password string) error {
	return m.ComparePassword(hashedPassword, password)
}

// MockTokenGen implements token.Generator
type MockTokenGen struct {
	mock.Mock
}

func (m *MockTokenGen) GenerateAccessToken(userID, tenantID, email string) (string, error) {
	args := m.Called(userID, tenantID, email)
	return args.String(0), args.Error(1)
}

func (m *MockTokenGen) GenerateRefreshToken(userID, tenantID, email string) (string, error) {
	args := m.Called(userID, tenantID, email)
	return args.String(0), args.Error(1)
}

// MockTokenVal implements token.Validator
type MockTokenVal struct {
	mock.Mock
}

func (m *MockTokenVal) ValidateToken(tokenString string) (*token.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*token.Claims), args.Error(1)
}

// MockNotifProducer implements services.NotificationProducer
type MockNotifProducer struct {
	mock.Mock
}

func (m *MockNotifProducer) SendEmail(ctx context.Context, to, subject, body string) error {
	args := m.Called(ctx, to, subject, body)
	return args.Error(0)
}

// MockEvtProducer implements services.EventProducer
type MockEvtProducer struct {
	mock.Mock
}

func (m *MockEvtProducer) PublishUserCreated(ctx context.Context, event events.UserCreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEvtProducer) PublishUserUpdated(ctx context.Context, event events.UserUpdatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEvtProducer) PublishUserDeleted(ctx context.Context, event events.UserDeletedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEvtProducer) PublishUserInvited(ctx context.Context, event events.UserInvitedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEvtProducer) PublishRoleAssigned(ctx context.Context, event events.RoleAssignedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEvtProducer) PublishRoleRevoked(ctx context.Context, event events.RoleRevokedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEvtProducer) PublishPasswordChanged(ctx context.Context, userID, tenantID string) error {
	args := m.Called(ctx, userID, tenantID)
	return args.Error(0)
}

// MockRoleRepo implements repository.RoleRepository
type MockRoleRepo struct {
	mock.Mock
}

func (m *MockRoleRepo) CreateRole(ctx context.Context, params db.CreateRoleParams) (db.Role, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Role), args.Error(1)
}

func (m *MockRoleRepo) GetRoleByID(ctx context.Context, roleID uuid.UUID) (db.Role, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).(db.Role), args.Error(1)
}

func (m *MockRoleRepo) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]db.Role, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]db.Role), args.Error(1)
}

func (m *MockRoleRepo) UpdateRole(ctx context.Context, params db.UpdateRoleParams) (db.Role, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Role), args.Error(1)
}

func (m *MockRoleRepo) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	args := m.Called(ctx, roleID)
	return args.Error(0)
}

func (m *MockRoleRepo) AssignRoleToUser(ctx context.Context, params db.AssignRoleToUserParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockRoleRepo) RevokeRoleFromUser(ctx context.Context, params db.RevokeRoleFromUserParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockRoleRepo) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]db.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Role), args.Error(1)
}

// MockOPA implements OPA client interface
type MockOPA struct {
	mock.Mock
}

func (m *MockOPA) Evaluate(ctx context.Context, input map[string]interface{}) (bool, error) {
	args := m.Called(ctx, input)
	return args.Bool(0), args.Error(1)
}

// MockOrgRepo implements repository.OrganizationRepository
type MockOrgRepo struct {
	mock.Mock
}

func (m *MockOrgRepo) CreateDepartment(ctx context.Context, params db.CreateDepartmentParams) (db.Department, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Department), args.Error(1)
}

func (m *MockOrgRepo) GetDepartmentByID(ctx context.Context, id uuid.UUID) (db.Department, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Department), args.Error(1)
}

func (m *MockOrgRepo) ListDepartments(ctx context.Context, tenantID uuid.UUID) ([]db.Department, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]db.Department), args.Error(1)
}

func (m *MockOrgRepo) UpdateDepartment(ctx context.Context, params db.UpdateDepartmentParams) (db.Department, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Department), args.Error(1)
}

func (m *MockOrgRepo) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrgRepo) CreateDesignation(ctx context.Context, params db.CreateDesignationParams) (db.Designation, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Designation), args.Error(1)
}

func (m *MockOrgRepo) GetDesignationByID(ctx context.Context, id uuid.UUID) (db.Designation, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Designation), args.Error(1)
}

func (m *MockOrgRepo) ListDesignations(ctx context.Context, tenantID uuid.UUID) ([]db.Designation, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]db.Designation), args.Error(1)
}

func (m *MockOrgRepo) UpdateDesignation(ctx context.Context, params db.UpdateDesignationParams) (db.Designation, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(db.Designation), args.Error(1)
}

func (m *MockOrgRepo) DeleteDesignation(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func testPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}
