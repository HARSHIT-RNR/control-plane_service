# Kafka Adapter - ERP Onboarding Flow

This folder contains all Kafka-related code for the Control Plane service, aligned with the ERP onboarding flow.

## 📁 Files (Only 4 Files)

```
kafka/
├── topics.go          ← Topic constants (3 topics)
├── producer.go        ← Event publisher
├── consumer.go        ← Event consumer
└── ensure_topics.go   ← Topic creation utility
```

---

## 🎯 Topics (3 Topics)

### 1. **tenant.onboarding**
- **Producer:** External Tenant Service
- **Consumer:** Control Plane (this service)
- **Purpose:** Create initial admin when tenant is provisioned
- **Event:**
  ```json
  {
    "tenant_id": "uuid",
    "admin_email": "admin@company.com",
    "admin_full_name": "John Admin"
  }
  ```

### 2. **user.lifecycle**
- **Producer:** Control Plane (this service)
- **Consumer:** Control Plane (this service)
- **Purpose:** User events (created, invited, updated, deleted)
- **Events:**
  ```json
  // UserCreated (with is_initial_admin flag)
  {
    "user_id": "uuid",
    "tenant_id": "uuid",
    "email": "user@company.com",
    "is_initial_admin": true  // ← Triggers password token generation
  }

  // UserInvited
  {
    "user_id": "uuid",
    "tenant_id": "uuid",
    "email": "user@company.com",
    "full_name": "Jane Doe"
  }
  ```

### 3. **notifications.send**
- **Producer:** Control Plane (this service)
- **Consumer:** External Notification Service
- **Purpose:** Send emails (password setup, invitations, etc.)
- **Event:**
  ```json
  {
    "to": "user@company.com",
    "subject": "Set Your Password",
    "body": "Click here: https://app.com/setup?token=..."
  }
  ```

---

## 🚀 Usage Examples

### **Example 1: Create Producer (in main.go)**

```go
import "cp_service/internal/adapters/kafka"

// Create producer
producer, err := kafka.NewProducer("localhost:9092")
if err != nil {
    log.Fatal(err)
}
defer producer.Close()

// Publish user created event
err = producer.PublishUserCreated(ctx, kafka.UserCreatedEvent{
    UserID:         "user-123",
    TenantID:       "tenant-456",
    Email:          "admin@company.com",
    IsInitialAdmin: true, // ← This triggers password token generation!
})
```

### **Example 2: Create Consumer (in main.go)**

```go
import "cp_service/internal/adapters/kafka"

// Define handler function
handler := func(ctx context.Context, key, value []byte) error {
    var event struct {
        TenantID      string `json:"tenant_id"`
        AdminEmail    string `json:"admin_email"`
        AdminFullName string `json:"admin_full_name"`
    }
    
    if err := json.Unmarshal(value, &event); err != nil {
        return err
    }
    
    log.Printf("Creating initial admin for tenant: %s", event.TenantID)
    
    // Call your service
    _, err := userService.CreateInitialAdmin(ctx, event.TenantID, event.AdminEmail, event.AdminFullName)
    return err
}

// Create consumer
consumer, err := kafka.NewConsumer(
    "localhost:9092",           // brokers
    "cp-tenant-onboarding-group", // groupID
    kafka.TopicTenantOnboarding,  // topic
    handler,                      // message handler
)
if err != nil {
    log.Fatal(err)
}
defer consumer.Close()

// Start consuming (blocking call - run in goroutine)
go func() {
    if err := consumer.Start(ctx); err != nil {
        log.Printf("Consumer error: %v", err)
    }
}()
```

### **Example 3: Ensure Topics at Startup**

```go
import "cp_service/internal/adapters/kafka"

// Create all topics if they don't exist
if err := kafka.EnsureTopics("localhost:9092"); err != nil {
    log.Printf("Warning: Failed to ensure topics: %v", err)
}
// Topics created:
// - tenant.onboarding
// - user.lifecycle
// - notifications.send
```

---

## 🔄 Complete Onboarding Flow

```
STEP 1: External Tenant Service
  ↓ Publishes to: tenant.onboarding
  
STEP 2: Control Plane Consumer
  ↓ Consumes: tenant.onboarding
  ↓ Action: CreateInitialAdmin (user created with PENDING_SETUP status)
  ↓ Publishes to: user.lifecycle (with is_initial_admin=true)
  
STEP 3: Control Plane Consumer (same service)
  ↓ Consumes: user.lifecycle
  ↓ Filters: is_initial_admin == true
  ↓ Action: GeneratePasswordSetupToken
  ↓ Publishes to: notifications.send
  
STEP 4: External Notification Service
  ↓ Consumes: notifications.send
  ↓ Action: Send email with password setup link
  
STEP 5: Admin clicks link and sets password
  ↓ gRPC call: SetInitialPassword
  ↓ Status: PENDING_SETUP → ACTIVE ✅
```

---

## 📝 Implementation in main.go

Here's how to wire everything together in your `cmd/service/main.go`:

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "cp_service/internal/adapters/kafka"
    "cp_service/internal/core/services"
)

func main() {
    ctx := context.Background()
    
    // 1. Ensure topics exist
    brokers := "localhost:9092"
    if err := kafka.EnsureTopics(brokers); err != nil {
        log.Printf("Warning: %v", err)
    }
    
    // 2. Create producer
    producer, err := kafka.NewProducer(brokers)
    if err != nil {
        log.Fatal(err)
    }
    defer producer.Close()
    
    // 3. Initialize services (your business logic)
    userService := services.NewUserService(...)
    authnService := services.NewAuthnService(...)
    
    // 4. Consumer 1: tenant.onboarding → CreateInitialAdmin
    tenantConsumer, err := kafka.NewConsumer(
        brokers,
        "cp-tenant-onboarding-group",
        kafka.TopicTenantOnboarding,
        func(ctx context.Context, key, value []byte) error {
            var event struct {
                TenantID      string `json:"tenant_id"`
                AdminEmail    string `json:"admin_email"`
                AdminFullName string `json:"admin_full_name"`
            }
            if err := json.Unmarshal(value, &event); err != nil {
                return err
            }
            
            log.Printf("Creating initial admin for tenant: %s", event.TenantID)
            user, err := userService.CreateInitialAdmin(ctx, event.TenantID, event.AdminEmail, event.AdminFullName)
            if err != nil {
                return err
            }
            
            // Publish UserCreated event
            return producer.PublishUserCreated(ctx, kafka.UserCreatedEvent{
                UserID:         user.ID,
                TenantID:       event.TenantID,
                Email:          event.AdminEmail,
                IsInitialAdmin: true, // ← KEY FLAG
            })
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    defer tenantConsumer.Close()
    
    // 5. Consumer 2: user.lifecycle → GeneratePasswordToken
    userConsumer, err := kafka.NewConsumer(
        brokers,
        "cp-user-lifecycle-group",
        kafka.TopicUserLifecycle,
        func(ctx context.Context, key, value []byte) error {
            var event kafka.UserCreatedEvent
            if err := json.Unmarshal(value, &event); err != nil {
                return err
            }
            
            // Only process initial admin events
            if !event.IsInitialAdmin {
                return nil // Skip regular users
            }
            
            log.Printf("Generating password token for: %s", event.Email)
            token, err := authnService.GeneratePasswordSetupToken(ctx, event.UserID, event.TenantID, event.Email)
            if err != nil {
                return err
            }
            
            // Send notification
            setupURL := fmt.Sprintf("https://app.company.com/auth/setup?token=%s", token)
            return producer.PublishEmailNotification(ctx, kafka.EmailNotificationEvent{
                To:      event.Email,
                Subject: "Set Your Password - Welcome!",
                Body:    fmt.Sprintf("Click here to set your password: %s", setupURL),
            })
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    defer userConsumer.Close()
    
    // 6. Start consumers in background
    go func() {
        if err := tenantConsumer.Start(ctx); err != nil {
            log.Printf("Tenant consumer error: %v", err)
        }
    }()
    
    go func() {
        if err := userConsumer.Start(ctx); err != nil {
            log.Printf("User consumer error: %v", err)
        }
    }()
    
    log.Println("🚀 Control Plane service started")
    log.Println("📡 Consumers active: tenant.onboarding, user.lifecycle")
    
    // 7. Wait for shutdown signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down...")
}
```

---

## ✅ Key Points

### 1. **Simple API**
- `NewProducer(brokers)` - Create producer
- `NewConsumer(brokers, groupID, topic, handler)` - Create consumer
- `EnsureTopics(brokers)` - Create topics

### 2. **Event-Driven Flow**
- Produce events after database operations
- Consume events to trigger actions
- Loose coupling between services

### 3. **Critical Flag: is_initial_admin**
```go
// When creating initial admin
producer.PublishUserCreated(ctx, kafka.UserCreatedEvent{
    UserID:         userID,
    TenantID:       tenantID,
    Email:          email,
    IsInitialAdmin: true, // ← MUST be true for initial admin!
})

// In consumer
if event.IsInitialAdmin {
    // Generate password setup token
    authnService.GeneratePasswordSetupToken(...)
}
```

### 4. **Error Handling**
- Consumers log errors and continue
- Failed messages can be retried (TODO: implement retry logic)
- Consider implementing dead-letter queue for production

### 5. **Production Considerations**
- Increase partitions for scalability
- Set replication factor to 3+ for high availability
- Use monitoring (Kafka lag, consumer group status)
- Implement circuit breakers
- Add distributed tracing

---

## 🧪 Testing

### Manual Testing with Kafka CLI

```bash
# 1. Produce test event
docker exec -it erp_kafka kafka-console-producer \
  --broker-list localhost:9092 \
  --topic tenant.onboarding

# Paste this JSON and press Enter:
{"tenant_id":"test-123","admin_email":"admin@test.com","admin_full_name":"Test Admin"}

# 2. Check logs
# You should see:
# [KAFKA] Received message from tenant.onboarding
# Creating initial admin for tenant: test-123
# [KAFKA] Message delivered to user.lifecycle
# [KAFKA] Received message from user.lifecycle
# Generating password token for: admin@test.com
# [KAFKA] Message delivered to notifications.send
```

### Consume Messages

```bash
# Watch tenant.onboarding
docker exec -it erp_kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic tenant.onboarding \
  --from-beginning

# Watch user.lifecycle
docker exec -it erp_kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic user.lifecycle \
  --from-beginning

# Watch notifications.send
docker exec -it erp_kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic notifications.send \
  --from-beginning
```

---

## 📋 Summary

| File | Purpose | Lines of Code |
|------|---------|---------------|
| `topics.go` | Topic constants (3 topics) | ~54 |
| `producer.go` | Publish events | ~119 |
| `consumer.go` | Consume events | ~107 |
| `ensure_topics.go` | Create topics | ~59 |

**Total:** 4 files, ~339 lines

**Simple, focused, production-ready!** ✅

---

## 🎯 Next Steps

1. ✅ Kafka files are ready
2. ⏭️ Update `cmd/service/main.go` to use new Kafka API
3. ⏭️ Update service files to call `producer.PublishUserCreated()`
4. ⏭️ Test onboarding flow end-to-end
5. ⏭️ Add monitoring and observability

---

**Clean, simple, and aligned with your ERP onboarding flow!** 🚀
