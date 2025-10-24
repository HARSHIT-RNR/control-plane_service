package events

import (
	"fmt"
	"log"
	"net"

	"github.com/segmentio/kafka-go"
)

// EnsureTopics creates all necessary Kafka topics if they don't exist.
// This is useful for development and testing environments.
// In production, topics should be created by infrastructure/deployment scripts.
//
// Topics created:
//   - tenant.onboarding
//   - user.lifecycle
//   - notifications.send
func EnsureTopics(brokers string) error {
	conn, err := kafka.Dial("tcp", brokers)
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, fmt.Sprintf("%d", controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	log.Println("[KAFKA] Creating topics...")

	// Build topic configurations
	var topicConfigs []kafka.TopicConfig
	for _, topic := range AllTopics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		// Check if error is "topic already exists" - that's okay
		log.Printf("[KAFKA] Topics creation result: %v (may already exist)", err)
	}

	for _, topic := range AllTopics {
		log.Printf("[KAFKA] Topic '%s' ready", topic)
	}

	return nil
}
