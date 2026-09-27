package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/domain"
	kafkago "github.com/segmentio/kafka-go"
)

type EventHandler interface {
	HandleEvent(context.Context, domain.EventEnvelope) error
}

type Consumer struct {
	reader  *kafkago.Reader
	handler EventHandler
}

func NewConsumer(brokers []string, topic, groupID string, handler EventHandler) *Consumer {
	return &Consumer{reader: kafkago.NewReader(kafkago.ReaderConfig{Brokers: brokers, Dialer: newKafkaDialer(), Topic: topic, GroupID: groupID, MinBytes: 1, MaxBytes: 10e6, CommitInterval: 0}), handler: handler}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("notification consumer: fetch failed: %v", err)
			continue
		}
		var envelope domain.EventEnvelope
		if err := json.Unmarshal(message.Value, &envelope); err != nil {
			log.Printf("notification consumer: malformed event: %v", err)
			_ = c.reader.CommitMessages(ctx, message)
			continue
		}
		if envelope.EventType != domain.OrderConfirmedEventType && envelope.EventType != domain.OrderCancelledEventType {
			_ = c.reader.CommitMessages(ctx, message)
			continue
		}
		if err := c.handler.HandleEvent(ctx, envelope); err != nil {
			log.Printf("notification consumer: process %s failed: %v", envelope.EventID, err)
			continue
		}
		if err := c.reader.CommitMessages(ctx, message); err != nil {
			log.Printf("notification consumer: commit %s failed: %v", envelope.EventID, err)
			continue
		}
		log.Printf("notification consumer: delivered %s", envelope.EventID)
	}
}

func (c *Consumer) Close() error { return c.reader.Close() }
