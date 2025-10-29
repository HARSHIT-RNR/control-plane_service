package main

import (
	"context"
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
	"cp_service/internal/adapters/password"
	"cp_service/internal/adapters/repository"
	"cp_service/internal/adapters/token"

	// "cp_service/internal/adapters/saml" // COMMENTED OUT
	"cp_service/internal/config"
	"cp_service/internal/core/services"
	grpcserver "cp_service/internal/ports/grpc_server"

	"github.com/jackc/pgx/v5/pgxpool"
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

// UserInvitedEvent represents user invitation event
type UserInvitedEvent struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
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

	// Database connection (using pgxpool for SQLC)
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		logger.Fatal("Failed to ping database: %v", err)
	}
	logger.Info("✓ Database connection established")

	// Initialize SQLC queries
	queries := db.New(pool)

	// --- Adapters (Infrastructure) ---

	// Password & Token utilities
	passwordHasher := password.NewBcryptHasher()
	tokenGenerator := token.NewJWTGenerator(cfg.TokenSymmetricKey, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
	tokenValidator := token.NewJWTValidator(cfg.TokenSymmetricKey)
	logger.Info("✓ Security adapters initialized")

	// Kafka Event Producer
	eventProducer, err := events.NewProducer(cfg.KafkaBrokers)
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

	// OPA Client (optional) - Disabled for now
	// TODO: Fix OPA client integration
	var opaClient services.OPAClient = nil
	// if cfg.OPAEnabled {
	// 	opaClient = opa.NewClient(cfg.OPAURL, "")
	// 	logger.Info("✓ OPA client initialized")
	// }

	// --- Repositories ---
	userRepo := repository.NewUserRepository(queries)
	roleRepo := repository.NewRoleRepository(queries)
	credentialRepo := repository.NewCredentialRepository(queries)
	organizationRepo := repository.NewOrganizationRepository(queries)
	logger.Info("✓ Repositories initialized")

	// --- Application Services ---
	userService := services.NewUserService(userRepo, roleRepo, eventProducer)
	authnService := services.NewAuthnService(userRepo, credentialRepo, passwordHasher, tokenGenerator, tokenValidator, notificationProducer, eventProducer)
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
			token, err := authnService.GeneratePasswordSetupToken(ctx, event.UserID, event.TenantID, event.Email)
			if err != nil {
				return fmt.Errorf("failed to generate password setup token: %w", err)
			}

			logger.Info("✓ Password setup token generated and email sent to: %s", event.Email)
			logger.Info("🔑 SETUP TOKEN: %s", token)
			logger.Info("📋 Copy this token and use it in SetInitialPassword RPC")
			return nil
		},
	)
	if err != nil {
		logger.Fatal("Failed to create user lifecycle consumer: %v", err)
	}
	defer userLifecycleConsumer.Close()
	logger.Info("✓ User lifecycle consumer initialized")

	// Consumer 3: User Invited (Sends invitation emails to invited users)
	userInvitedConsumer, err := events.NewConsumer(
		cfg.KafkaBrokers,
		"cp-user-invited-group",
		cfg.KafkaTopicUserLifecycle,
		func(ctx context.Context, key, value []byte) error {
			var event UserInvitedEvent
			if err := json.Unmarshal(value, &event); err != nil {
				// Try to parse as UserInvitedEvent from kafka producer
				var inviteEvent events.UserInvitedEvent
				if err := json.Unmarshal(value, &inviteEvent); err != nil {
					return fmt.Errorf("failed to unmarshal event: %w", err)
				}
				event.UserID = inviteEvent.UserID
				event.TenantID = inviteEvent.TenantID
				event.Email = inviteEvent.Email
				event.FullName = inviteEvent.FullName
			}

			logger.Info("Processing user invitation for: %s", event.Email)

			// Generate invitation token (similar to password setup)
			token, err := authnService.GeneratePasswordSetupToken(ctx, event.UserID, event.TenantID, event.Email)
			if err != nil {
				return fmt.Errorf("failed to generate invitation token: %w", err)
			}

			// Publish notification event for the notification service
			notificationEvent := events.EmailNotificationEvent{
				To:      event.Email,
				Subject: "Invitation to Join",
				Body:    fmt.Sprintf("You have been invited to join. Use this token: %s", token),
			}
			
			if err := eventProducer.PublishEmailNotification(ctx, notificationEvent); err != nil {
				logger.Error("Failed to publish email notification: %v", err)
				// Don't fail the whole process if notification fails
			}

			logger.Info("✓ Invitation email sent to: %s", event.Email)
			logger.Info("🔑 INVITATION TOKEN: %s", token)
			logger.Info("📋 Copy this token and use it in RegisterInvitedUser RPC")
			return nil
		},
	)
	if err != nil {
		logger.Fatal("Failed to create user invited consumer: %v", err)
	}
	defer userInvitedConsumer.Close()
	logger.Info("✓ User invited consumer initialized")

	// Create cancellable context for consumers
	ctx, cancel := context.WithCancel(context.Background())

	// Start consumers in background with proper context
	go func() {
		if err := tenantOnboardingConsumer.Start(ctx); err != nil && ctx.Err() == nil {
			logger.Error("Tenant onboarding consumer error: %v", err)
		}
	}()

	go func() {
		if err := userLifecycleConsumer.Start(ctx); err != nil && ctx.Err() == nil {
			logger.Error("User lifecycle consumer error: %v", err)
		}
	}()

	go func() {
		if err := userInvitedConsumer.Start(ctx); err != nil && ctx.Err() == nil {
			logger.Error("User invited consumer error: %v", err)
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
	logger.Info("🔄 Kafka consumers active: iam.create-initial-admin, user.lifecycle (created), user.lifecycle (invited)")

	// if cfg.SAMLEnabled && samlProvider != nil {
	// 	logger.Info("🔐 SAML SSO: ENABLED")
	// }

	logger.Info("Ready to accept requests...")

	// Start gRPC server in background
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
	
	// Cancel context to stop consumers
	cancel()
	
	// Stop gRPC server
	grpcServer.GracefulStop()
	
	// Close event producer
	eventProducer.Close()
	
	logger.Info("✓ All components stopped gracefully")
}
