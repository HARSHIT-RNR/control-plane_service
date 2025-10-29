# Control Plane Microservice

A production-ready backend microservice for ERP system control plane featuring user management, RBAC, authentication, and authorization.

## 🚀 Features

### 1. **User Management Service**
- ✅ Create/Read/Update/Delete users
- ✅ User invitation flow (PENDING_INVITE → ACTIVE)
- ✅ Initial admin creation for tenant onboarding
- ✅ Employee ID, department, designation tracking
- ✅ User status lifecycle management

### 2. **RBAC (Role-Based Access Control)**
- ✅ Dynamic role creation with custom permissions
- ✅ Role assignment/revocation
- ✅ Permission-based authorization
- ✅ Tenant-scoped roles

### 3. **Authentication Service (AuthN)**
- ✅ JWT-based authentication
- ✅ Password hashing with bcrypt
- ✅ Initial password setup flow
- ✅ Password reset flow
- ✅ Refresh token support
- ✅ Login with tenant context

### 4. **Authorization Service (AuthZ)**
- ✅ High-speed permission checks
- ✅ OPA (Open Policy Agent) integration
- ✅ Fallback to simple permission matching
- ✅ Resource-action-based access control

### 5. **Organization Structure**
- ✅ Department management
- ✅ Designation management
- ✅ Tenant-scoped entities

### 6. **Event-Driven Architecture**
- ✅ Kafka event producers for domain events
- ✅ Kafka consumers for async workflows
- ✅ User lifecycle events
- ✅ Notification events

## 📋 Prerequisites

- Go 1.24+
- PostgreSQL 15+
- Apache Kafka
- SQLC (for code generation)
- Protocol Buffers compiler

## 📡 gRPC Services

### UserService
- `CreateInitialAdmin` - Create first admin for tenant
- `CreateUser` - Create regular user
- `InviteUser` - Invite user (PENDING_INVITE state)
- `GetUser`, `ListUsers`, `UpdateUser`, `DeleteUser`
- `CreateRole`, `GetRole`, `ListRoles`, `UpdateRole`, `DeleteRole`
- `AssignRoleToUser`, `RevokeRoleFromUser`, `ListUserRoles`

### AuthnService
- `Login` - Authenticate user, return JWT tokens
- `Logout` - Invalidate token
- `RefreshToken` - Get new access token
- `SetInitialPassword` - Set password using setup token
- `RegisterInvitedUser` - Complete invitation registration
- `ForgotPassword`, `ResetPassword`
- `ValidateToken`, `ConfirmPassword`

### AuthzService
- `CheckAccess` - Verify user permissions

### OrganizationService
- `CreateDepartment`, `GetDepartment`, `ListDepartments`, `UpdateDepartment`, `DeleteDepartment`
- `CreateDesignation`, `GetDesignation`, `ListDesignations`, `UpdateDesignation`, `DeleteDesignation`

## 🔄 Onboarding Flow

**Tenant Onboarding Process:**

1. **External System** → Publishes `tenant.onboarding` event to Kafka
2. **Control Plane** → Consumes event, creates initial admin user (PENDING_SETUP)
3. **Control Plane** → Publishes `user.created` event
4. **AuthN Service** → Consumes event, generates password setup token
5. **Notification Service** → Sends setup email to admin
6. **Admin** → Clicks link, sets password (PENDING_SETUP → ACTIVE)
7. **Tenant is now active** ✅

## 🔐 Security

- **Password Hashing**: bcrypt with salt
- **JWT Tokens**: HS256 signing, short-lived access tokens
- **Token Refresh**: Separate refresh token flow
- **Permission Model**: Fine-grained resource:action permissions
// - **OPA Integration**: Policy-based authorization


## 🎯 Design Patterns

- **Hexagonal Architecture**: Clean separation of concerns
- **Repository Pattern**: Data access abstraction
- **Dependency Injection**: All dependencies injected
- **Event Sourcing**: Domain events for audit trail
- **CQRS Lite**: Separate read/write optimizations

