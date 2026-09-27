package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/gopher-opsx/platform-lab/services/order-service/internal/domain"
)

type PaymentEventHandler interface {
	HandleEvent(ctx context.Context, envelope domain.EventEnvelope) error
}

type PaymentConsumer struct {
	reader  *kafkago.Reader
	handler PaymentEventHandler
}

func NewPaymentConsumer(
	brokers []string,
	topic string,
	groupID string,
	handler PaymentEventHandler,
) *PaymentConsumer {
	return &PaymentConsumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:        brokers,
			Dialer:         newKafkaDialer(),
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: 0,
		}),
		handler: handler,
	}
}

func (c *PaymentConsumer) Run(ctx context.Context) {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("order payment consumer: fetch failed: %v", err)
			continue
		}

		var envelope domain.EventEnvelope
		if err := json.Unmarshal(message.Value, &envelope); err != nil {
			log.Printf("order payment consumer: malformed event: %v", err)
			_ = c.reader.CommitMessages(ctx, message)
			continue
		}

		if envelope.EventType != domain.PaymentAuthorizedEventType &&
			envelope.EventType != domain.PaymentFailedEventType {
			_ = c.reader.CommitMessages(ctx, message)
			continue
		}

		if err := c.handler.HandleEvent(ctx, envelope); err != nil {
			log.Printf("order payment consumer: process %s failed: %v", envelope.EventID, err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			log.Printf("order payment consumer: commit %s failed: %v", envelope.EventID, err)
			continue
		}

		log.Printf("order payment consumer: processed %s", envelope.EventID)
	}
}

func (c *PaymentConsumer) Close() error {
	return c.reader.Close()
}
