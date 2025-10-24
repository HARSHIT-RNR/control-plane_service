# Control Plane Microservice

A production-ready backend microservice for ERP system control plane featuring user management, RBAC, authentication, and authorization.

## 🏗️ Architecture

**Hexagonal Architecture (Ports & Adapters)**
```
├── Domain Layer (Business Entities)
│   ├── User, Role, Department, Designation
│   └── Pure business logic, framework-agnostic
│
├── Application Layer (Use Cases)
│   ├── UserService, AuthnService, AuthzService
│   └── Orchestrates business flows
│
├── Ports (Interfaces)
│   ├── Repository interfaces
│   └── Service interfaces
│
└── Adapters (Infrastructure)
    ├── Repositories (PostgreSQL via SQLC)
    ├── gRPC Handlers
    ├── Kafka Producers/Consumers
    ├── OPA Client (Authorization)
    └── JWT Token Manager
```

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

## 🛠️ Setup

### 1. Install Dependencies

```bash
go mod download
```

### 2. Setup Database

```bash
# Start PostgreSQL
docker-compose up -d postgres

# Run migrations
psql -U postgres -d cp_db -f internal/adapters/database/migration/001_init.up.sql
```

### 3. Generate Code

```bash
# Generate SQLC code
sqlc generate

# Generate gRPC code (if proto files changed)
protoc --go_out=. --go-grpc_out=. api/proto/*.proto
```

### 4. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 5. Run the Service

```bash
go run cmd/main_new.go
```

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
- **OPA Integration**: Policy-based authorization

## 📊 Database Schema

```sql
Tables:
- users (id, full_name, email, tenant_id, status, ...)
- roles (id, tenant_id, name, permissions[])
- user_roles (user_id, role_id, tenant_id)
- credentials (user_id, password_hash)
- tokens (hash, user_id, expiry, scope)
- departments (id, name, tenant_id)
- designations (id, name, tenant_id)
```

## 🎯 Design Patterns

- **Hexagonal Architecture**: Clean separation of concerns
- **Repository Pattern**: Data access abstraction
- **Dependency Injection**: All dependencies injected
- **Event Sourcing**: Domain events for audit trail
- **CQRS Lite**: Separate read/write optimizations

## 🔧 Development

### Running Tests
```bash
go test ./...
```

### Code Generation
```bash
# After modifying SQL queries
sqlc generate

# After modifying proto files
protoc --go_out=. --go-grpc_out=. api/proto/*.proto
```

### gRPC Testing
```bash
# Using grpcurl
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext -d '{"email":"admin@example.com","password":"pass123","tenant_identifier":"tenant-1"}' localhost:50051 cp.AuthnService/Login
```

## 📦 Project Structure

```
cp_service/
├── api/proto/              # gRPC protocol definitions
├── cmd/                    # Application entry points
├── internal/
│   ├── user/
│   │   ├── domain/        # User entities
│   │   ├── application/   # UserService
│   │   └── ports/         # Repository interfaces
│   ├── authn/             # Authentication domain
│   ├── authz/             # Authorization domain
│   ├── organization/      # Org structure domain
│   └── adapters/
│       ├── postgres/      # Repository implementations
│       ├── kafka/         # Event producers/consumers
│       ├── opa/           # OPA client
│       └── ports/grpc/    # gRPC handlers
├── adapters/
│   ├── password/          # Password hashing
│   └── token/             # JWT generation/validation
└── config/                # Configuration management
```

## 🚀 Deployment

### Docker
```bash
docker-compose up -d
```

### Kubernetes
```bash
kubectl apply -f k8s/
```

## 🔍 Monitoring & Observability

- Structured logging throughout
- gRPC reflection enabled
- Health checks on all services
- Kafka consumer lag monitoring

## 📝 License

Proprietary - Internal ERP System

## 👥 Team

Backend Control Plane Team
