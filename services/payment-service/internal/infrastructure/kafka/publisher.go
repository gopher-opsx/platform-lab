package kafka

import (
	"context"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafkago.Writer
}

func NewPublisher(brokers []string) *Publisher {
	return &Publisher{
		writer: &kafkago.Writer{
			Addr:                   kafkago.TCP(brokers...),
			Transport:              newKafkaTransport(),
			Balancer:               &kafkago.Hash{},
			RequiredAcks:           kafkago.RequireAll,
			AllowAutoTopicCreation: false,
			BatchTimeout:           10 * time.Millisecond,
		},
	}
}

func (p *Publisher) Publish(ctx context.Context, topic, key string, payload []byte) error {
	if err := p.writer.WriteMessages(ctx, kafkago.Message{
		Topic: topic, Key: []byte(key), Value: payload, Time: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("publish kafka message: %w", err)
	}
	return nil
}

func (p *Publisher) Close() error {
	return p.writer.Close()
}
