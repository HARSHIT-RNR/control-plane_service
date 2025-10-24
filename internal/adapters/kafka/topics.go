package events

// This file centralizes all Kafka topic names used across the control plane.
// Topics are aligned with the ERP onboarding flow.

const (
	// ====================================================================
	// TOPIC 1: iam.create-initial-admin
	// ====================================================================
	// External tenant-service produces events when a new tenant is created.
	// Control Plane consumes these events to create the initial admin user.
	// 
	// Flow: Tenant Service → Kafka → Control Plane (CreateInitialAdmin)
	TopicIAMCreateInitialAdmin = "iam.create-initial-admin"

	// ====================================================================
	// TOPIC 2: user.lifecycle
	// ====================================================================
	// Control Plane produces and consumes user lifecycle events.
	// 
	// Events published to this topic:
	//   - UserCreated (with is_initial_admin flag)
	//   - UserInvited
	//   - UserUpdated
	//   - UserDeleted
	//   - RoleAssigned
	//   - RoleRevoked
	// 
	// Flow: Control Plane → Kafka → Control Plane (for initial admin password token)
	TopicUserLifecycle = "user.lifecycle"

	// ====================================================================
	// TOPIC 3: notification.send-password-setup
	// ====================================================================
	// Control Plane produces email notification requests.
	// External notification-service consumes these events to send emails.
	// 
	// Events published to this topic:
	//   - PasswordSetupTokenGenerated (send setup email)
	//   - InvitationTokenGenerated (send invitation email)
	//   - PasswordResetRequested (send reset email)
	// 
	// Flow: Control Plane → Kafka → Notification Service → Email
	TopicNotificationSendPasswordSetup = "notification.send-password-setup"
)

// AllTopics returns all topics used by the control plane service.
// This is used for automated topic creation at startup.
var AllTopics = []string{
	TopicIAMCreateInitialAdmin,
	TopicUserLifecycle,
	TopicNotificationSendPasswordSetup,
}
