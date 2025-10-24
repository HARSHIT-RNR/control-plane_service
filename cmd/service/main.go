package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "cp_service/api/proto/pb"
	"cp_service/internal/adapters/database/db"
	events "cp_service/internal/adapters/kafka"
	"cp_service/internal/adapters/logger"
	"cp_service/internal/adapters/opa"
	"cp_service/internal/adapters/password"
	"cp_service/internal/adapters/repository"
	"cp_service/internal/adapters/token"

	// "cp_service/internal/adapters/saml" // COMMENTED OUT
	"cp_service/internal/config"
	"cp_service/internal/core/services"
	grpcserver "cp_service/internal/ports/grpc_server"

	confluentkafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// CreateInitialAdminEvent represents the tenant onboarding event
type CreateInitialAdminEvent struct {
	TenantID      string `json:"tenant_id"`
	AdminEmail    string `json:"admin_email"`
	AdminFullName string `json:"admin_full_name"`
}

// UserCreatedEvent represents user lifecycle event
type UserCreatedEvent struct {
	UserID         string `json:"user_id"`
	TenantID       string `json:"tenant_id"`
	Email          string `json:"email"`
	IsInitialAdmin bool   `json:"is_initial_admin"`
}

func main() {
	// Initialize logger
	logger := logger.New()
	logger.Info("Starting Control Plane Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}
	logger.Info("✓ Configuration loaded")

	// Database connection
	database, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		logger.Fatal("Failed to ping database: %v", err)
	}
	logger.Info("✓ Database connection established")

	// Initialize SQLC queries
	queries := db.New(database)

	// --- Adapters (Infrastructure) ---

	// Password & Token utilities
	passwordHasher := password.NewBcryptHasher()
	tokenGenerator := token.NewJWTGenerator(cfg.TokenSymmetricKey, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
	tokenValidator := token.NewJWTValidator(cfg.TokenSymmetricKey)
	logger.Info("✓ Security adapters initialized")

	// Kafka Event Producer
	kafkaConfig := &confluentkafka.ConfigMap{"bootstrap.servers": cfg.KafkaBrokers}
	eventProducer, err := events.NewKafkaEventProducer(kafkaConfig, events.TopicConfig{
		GeneratePasswordToken: cfg.KafkaTopicUserLifecycle,
		UserCreated:           cfg.KafkaTopicUserLifecycle,
		UserInvited:           cfg.KafkaTopicUserLifecycle,
		UserUpdated:           cfg.KafkaTopicUserLifecycle,
		UserDeleted:           cfg.KafkaTopicUserLifecycle,
		RoleAssigned:          cfg.KafkaTopicUserLifecycle,
		RoleRevoked:           cfg.KafkaTopicUserLifecycle,
	})
	if err != nil {
		logger.Fatal("Failed to create kafka event producer: %v", err)
	}
	defer eventProducer.Close()
	logger.Info("✓ Kafka event producer initialized")

	// Notification Producer
	notificationProducer, err := events.NewNotificationProducer(cfg.KafkaBrokers, cfg.KafkaTopicNotificationPasswordSetup)
	if err != nil {
		logger.Fatal("Failed to create notification producer: %v", err)
	}
	defer notificationProducer.Close()
	logger.Info("✓ Notification producer initialized")

	// SAML Provider (COMMENTED OUT FOR NOW)
	// var samlProvider services.SAMLProvider
	// if cfg.SAMLEnabled {
	// 	samlProvider, err = saml.NewSAMLProvider(saml.Config{
	// 		EntityID:          cfg.SAMLEntityID,
	// 		SSOURL:            cfg.SAMLSSOURL,
	// 		IDPMetadataURL:    cfg.SAMLIDPMetadataURL,
	// 		CertificateFile:   cfg.SAMLCertificateFile,
	// 		PrivateKeyFile:    cfg.SAMLPrivateKeyFile,
	// 		RootURL:           cfg.SAMLRootURL,
	// 		AllowIDPInitiated: cfg.SAMLAllowIDPInitiated,
	// 	})
	// 	if err != nil {
	// 		logger.Warn("SAML provider initialization failed (continuing without SAML): %v", err)
	// 		samlProvider = nil
	// 	} else {
	// 		logger.Info("✓ SAML provider initialized")
	// 	}
	// }

	// OPA Client (optional)
	var opaClient services.OPAClient
	if cfg.OPAEnabled {
		opaClient = opa.NewClient(cfg.OPAURL)
		logger.Info("✓ OPA client initialized")
	}

	// --- Repositories ---
	userRepo := repository.NewUserRepository(database, queries)
	roleRepo := repository.NewRoleRepository(database, queries)
	credentialRepo := repository.NewCredentialRepository(database, queries)
	organizationRepo := repository.NewOrganizationRepository(database, queries)
	logger.Info("✓ Repositories initialized")

	// --- Application Services ---
	userService := services.NewUserService(userRepo, roleRepo, eventProducer)
	authnService := services.NewAuthnService(userRepo, credentialRepo, passwordHasher, tokenGenerator, tokenValidator, notificationProducer)
	authzService := services.NewAuthzService(roleRepo, tokenValidator, opaClient)
	organizationService := services.NewOrganizationService(organizationRepo)
	logger.Info("✓ Application services initialized")

	// --- gRPC Handlers ---
	userHandler := grpcserver.NewUserHandler(userService)
	authnHandler := grpcserver.NewAuthnHandler(authnService)
	authzHandler := grpcserver.NewAuthzHandler(authzService)
	organizationHandler := grpcserver.NewOrganizationHandler(organizationService)
	logger.Info("✓ gRPC handlers initialized")

	// --- Kafka Consumers for Onboarding Flow ---

	// Consumer 1: IAM Create Initial Admin (Step 9-10 - Creates initial admin)
	tenantOnboardingConsumer, err := events.NewConsumer(
		cfg.KafkaBrokers,
		"cp-iam-initial-admin-group",
		events.TopicIAMCreateInitialAdmin,
		func(ctx context.Context, key, value []byte) error {
			var event CreateInitialAdminEvent
			if err := json.Unmarshal(value, &event); err != nil {
				return fmt.Errorf("failed to unmarshal event: %w", err)
			}

			logger.Info("Processing tenant onboarding for tenant: %s", event.TenantID)

			// Create initial admin user (Step 2 in onboarding flow)
			_, err := userService.CreateInitialAdmin(ctx, event.TenantID, event.AdminEmail, event.AdminFullName)
			if err != nil {
				return fmt.Errorf("failed to create initial admin: %w", err)
			}

			logger.Info("✓ Initial admin created for tenant: %s", event.TenantID)
			return nil
		},
	)
	if err != nil {
		logger.Fatal("Failed to create tenant onboarding consumer: %v", err)
	}
	defer tenantOnboardingConsumer.Close()
	logger.Info("✓ Tenant onboarding consumer initialized")

	// Consumer 2: User Lifecycle (Step 3 - Generates password token for initial admin)
	userLifecycleConsumer, err := events.NewConsumer(
		cfg.KafkaBrokers,
		"cp-user-lifecycle-group",
		cfg.KafkaTopicUserLifecycle,
		func(ctx context.Context, key, value []byte) error {
			var event UserCreatedEvent
			if err := json.Unmarshal(value, &event); err != nil {
				return fmt.Errorf("failed to unmarshal event: %w", err)
			}

			// Only process initial admin users (Step 4 in onboarding flow)
			if !event.IsInitialAdmin {
				return nil
			}

			logger.Info("Generating password setup token for initial admin: %s", event.UserID)

			// Generate password setup token (Step 5 in onboarding flow)
			_, err := authnService.GeneratePasswordSetupToken(ctx, event.UserID, event.TenantID, event.Email)
			if err != nil {
				return fmt.Errorf("failed to generate password setup token: %w", err)
			}

			logger.Info("✓ Password setup token generated and email sent to: %s", event.Email)
			return nil
		},
	)
	if err != nil {
		logger.Fatal("Failed to create user lifecycle consumer: %v", err)
	}
	defer userLifecycleConsumer.Close()
	logger.Info("✓ User lifecycle consumer initialized")

	// Start consumers in background
	go func() {
		if err := tenantOnboardingConsumer.Start(context.Background()); err != nil {
			logger.Error("Tenant onboarding consumer error: %v", err)
		}
	}()

	go func() {
		if err := userLifecycleConsumer.Start(context.Background()); err != nil {
			logger.Error("User lifecycle consumer error: %v", err)
		}
	}()

	// --- gRPC Server Setup ---
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterUserServiceServer(grpcServer, userHandler)
	pb.RegisterAuthnServiceServer(grpcServer, authnHandler)
	pb.RegisterAuthzServiceServer(grpcServer, authzHandler)
	pb.RegisterOrganizationServiceServer(grpcServer, organizationHandler)

	// Enable reflection for grpcurl
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to create listener: %v", err)
	}

	logger.Info("🚀 Control Plane gRPC server starting on port: %d", cfg.GRPCPort)
	logger.Info("📡 Services registered: UserService, AuthnService, AuthzService, OrganizationService")
	logger.Info("🔄 Kafka consumers active: iam.create-initial-admin, %s", cfg.KafkaTopicUserLifecycle)

	// if cfg.SAMLEnabled && samlProvider != nil {
	// 	logger.Info("🔐 SAML SSO: ENABLED")
	// }

	logger.Info("Ready to accept requests...")

	// Graceful shutdown
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gracefully...")
	grpcServer.GracefulStop()
	logger.Info("Server stopped")
}
