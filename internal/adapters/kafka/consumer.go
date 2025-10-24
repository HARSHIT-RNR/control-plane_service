package events

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

// ====================================================================
// MessageHandler is a function type for processing Kafka messages
// ====================================================================

type MessageHandler func(ctx context.Context, key, value []byte) error

// ====================================================================
// Consumer
// ====================================================================

// Consumer wraps a Kafka consumer with a message handler
type Consumer struct {
	reader  *kafka.Reader
	handler MessageHandler
	topic   string
}

// NewConsumer creates a new Kafka consumer for a specific topic
// 
// Parameters:
//   - brokers: Kafka broker addresses (e.g., "localhost:9092")
//   - groupID: Consumer group ID
//   - topic: Topic to subscribe to
//   - handler: Function to handle incoming messages
func NewConsumer(brokers, groupID, topic string, handler MessageHandler) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokers},
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	log.Printf("[KAFKA] Consumer created for topic: %s (group: %s)", topic, groupID)

	return &Consumer{
		reader:  reader,
		handler: handler,
		topic:   topic,
	}, nil
}

// Start begins consuming messages (blocking call)
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("[KAFKA] Consumer started for topic: %s", c.topic)
	defer log.Printf("[KAFKA] Consumer stopped for topic: %s", c.topic)

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				log.Printf("[KAFKA] Context cancelled, stopping consumer for topic: %s", c.topic)
				return nil
			}
			log.Printf("[KAFKA] Error reading message from %s: %v", c.topic, err)
			continue
		}

		// Process message
		log.Printf("[KAFKA] Received message from %s (partition %d, offset %d)",
			c.topic, msg.Partition, msg.Offset)

		if err := c.handler(ctx, msg.Key, msg.Value); err != nil {
			log.Printf("[KAFKA] Error handling message from %s: %v", c.topic, err)
			// TODO: Implement retry logic or dead-letter queue
			continue
		}

		// Commit the message after successful processing
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("[KAFKA] Error committing message: %v", err)
		}

		log.Printf("[KAFKA] Message processed successfully from %s", c.topic)
	}
}

// Close closes the consumer
func (c *Consumer) Close() {
	log.Printf("[KAFKA] Closing consumer for topic: %s", c.topic)
	if err := c.reader.Close(); err != nil {
		log.Printf("[KAFKA] Error closing consumer: %v", err)
	}
}
