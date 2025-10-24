package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

// ====================================================================
// NotificationProducer - Wrapper for sending email notifications
// ====================================================================

// NotificationProducer sends email notifications via Kafka
type NotificationProducer struct {
	writer *kafka.Writer
	topic  string
}

// NewNotificationProducer creates a new notification producer
func NewNotificationProducer(brokers, topic string) (*NotificationProducer, error) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	log.Printf("[KAFKA] Notification producer created for topic: %s", topic)

	return &NotificationProducer{
		writer: writer,
		topic:  topic,
	}, nil
}

// SendEmail sends an email notification by publishing to Kafka
func (np *NotificationProducer) SendEmail(ctx context.Context, to, subject, body string) error {
	event := EmailNotificationEvent{
		To:      to,
		Subject: subject,
		Body:    body,
	}

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal email event: %w", err)
	}

	err = np.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(to),
		Value: value,
	})

	if err != nil {
		return fmt.Errorf("failed to produce notification to %s: %w", np.topic, err)
	}

	log.Printf("[KAFKA] Email notification sent to %s", to)
	return nil
}

// Close closes the notification producer
func (np *NotificationProducer) Close() {
	if err := np.writer.Close(); err != nil {
		log.Printf("[KAFKA] Error closing notification producer: %v", err)
	}
}
