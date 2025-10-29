package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

// ====================================================================
// Event Payloads (aligned with ERP onboarding flow)
// ====================================================================

// UserCreatedEvent - Published to user.lifecycle after creating a user
type UserCreatedEvent struct {
	UserID         string `json:"user_id"`
	TenantID       string `json:"tenant_id"`
	Email          string `json:"email"`
	IsInitialAdmin bool   `json:"is_initial_admin"` // Critical flag for onboarding flow
}

// UserInvitedEvent - Published to user.lifecycle when a user is invited
type UserInvitedEvent struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// UserUpdatedEvent - Published when user info is updated
type UserUpdatedEvent struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

// UserDeletedEvent - Published when user is deleted
type UserDeletedEvent struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

// RoleAssignedEvent - Published when role is assigned to user
type RoleAssignedEvent struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	RoleID   string `json:"role_id"`
}

// RoleRevokedEvent - Published when role is revoked from user
type RoleRevokedEvent struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	RoleID   string `json:"role_id"`
}

// UserStatusChangedEvent - Published when user status changes
type UserStatusChangedEvent struct {
	UserID    string `json:"user_id"`
	TenantID  string `json:"tenant_id"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	ChangedBy string `json:"changed_by"` // user_id or "system"
	Reason    string `json:"reason"`
	Timestamp string `json:"timestamp"`
}

// PasswordChangedEvent - Published when password is changed
type PasswordChangedEvent struct {
	UserID    string `json:"user_id"`
	TenantID  string `json:"tenant_id"`
	ChangedBy string `json:"changed_by"` // "self" or admin_user_id
	Method    string `json:"method"`     // "reset", "initial_setup", "change"
	Timestamp string `json:"timestamp"`
}

// UserLoginEvent - Published on successful login
type UserLoginEvent struct {
	UserID    string `json:"user_id"`
	TenantID  string `json:"tenant_id"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Timestamp string `json:"timestamp"`
}

// EmailNotificationEvent - Published to notifications.send for email delivery
type EmailNotificationEvent struct {
	To      string `json:"email"` // Changed from "to" to "email" to match notification service
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// ====================================================================
// Producer
// ====================================================================

// Producer publishes events to Kafka topics
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers string) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers),
		Balancer: &kafka.LeastBytes{},
	}

	log.Printf("[KAFKA] Producer created for brokers: %s", brokers)

	return &Producer{writer: writer}, nil
}

// PublishUserCreated publishes a user created event to user.lifecycle topic
// If is_initial_admin=true, this triggers the password setup token generation
func (p *Producer) PublishUserCreated(ctx context.Context, event UserCreatedEvent) error {
	return p.publish(TopicUserLifecycle, event.UserID, event)
}

// PublishUserInvited publishes a user invited event to user.lifecycle topic
func (p *Producer) PublishUserInvited(ctx context.Context, event UserInvitedEvent) error {
	return p.publish(TopicUserLifecycle, event.UserID, event)
}

// PublishUserUpdated publishes a user updated event to user.lifecycle topic
func (p *Producer) PublishUserUpdated(ctx context.Context, event UserUpdatedEvent) error {
	return p.publish(TopicUserLifecycle, event.UserID, event)
}

// PublishUserDeleted publishes a user deleted event to user.lifecycle topic
func (p *Producer) PublishUserDeleted(ctx context.Context, event UserDeletedEvent) error {
	return p.publish(TopicUserLifecycle, event.UserID, event)
}

// PublishRoleAssigned publishes a role assigned event to user.lifecycle topic
func (p *Producer) PublishRoleAssigned(ctx context.Context, event RoleAssignedEvent) error {
	return p.publish(TopicUserLifecycle, event.UserID, event)
}

// PublishRoleRevoked publishes a role revoked event to user.lifecycle topic
func (p *Producer) PublishRoleRevoked(ctx context.Context, event RoleRevokedEvent) error {
	return p.publish(TopicUserLifecycle, event.UserID, event)
}

// PublishUserStatusChanged publishes a user status changed event to user.lifecycle topic
func (p *Producer) PublishUserStatusChanged(ctx context.Context, event UserStatusChangedEvent) error {
	return p.publish(TopicUserLifecycle, event.UserID, event)
}

// PublishPasswordChanged publishes a password changed event to user.lifecycle topic
func (p *Producer) PublishPasswordChanged(ctx context.Context, event PasswordChangedEvent) error {
	return p.publish(TopicUserLifecycle, event.UserID, event)
}

// PublishUserLogin publishes a user login event to user.lifecycle topic
func (p *Producer) PublishUserLogin(ctx context.Context, event UserLoginEvent) error {
	return p.publish(TopicUserLifecycle, event.UserID, event)
}

// PublishEmailNotification publishes an email notification request to notification.send-password-setup topic
func (p *Producer) PublishEmailNotification(ctx context.Context, event EmailNotificationEvent) error {
	return p.publish(TopicNotificationSendPasswordSetup, event.To, event)
}

// publish is a generic helper to publish any event to a given topic
func (p *Producer) publish(topic string, key string, payload interface{}) error {
	value, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	})

	if err != nil {
		return fmt.Errorf("failed to produce message to %s: %w", topic, err)
	}

	log.Printf("[KAFKA] Message published to %s", topic)
	return nil
}

// Close closes the producer
func (p *Producer) Close() {
	if err := p.writer.Close(); err != nil {
		log.Printf("[KAFKA] Error closing producer: %v", err)
	}
	log.Println("[KAFKA] Producer closed")
}
