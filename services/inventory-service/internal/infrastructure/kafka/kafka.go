package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/domain"
	kafkago "github.com/segmentio/kafka-go"
)

type Handler interface {
	HandleEvent(context.Context, domain.EventEnvelope) error
}

type Consumer struct {
	r *kafkago.Reader
	h Handler
}

func NewConsumer(b []string, topic, group string, h Handler) *Consumer {
	return &Consumer{r: kafkago.NewReader(kafkago.ReaderConfig{Brokers: b, Dialer: newKafkaDialer(), Topic: topic, GroupID: group, CommitInterval: 0}), h: h}
}
func (c *Consumer) Run(ctx context.Context) {
	for {
		m, err := c.r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("inventory consumer fetch: %v", err)
			continue
		}
		var e domain.EventEnvelope
		if err := json.Unmarshal(m.Value, &e); err != nil {
			_ = c.r.CommitMessages(ctx, m)
			continue
		}
		if err := c.h.HandleEvent(ctx, e); err != nil {
			log.Printf("inventory process %s: %v", e.EventID, err)
			continue
		}
		if err := c.r.CommitMessages(ctx, m); err != nil {
			log.Printf("inventory commit %s: %v", e.EventID, err)
			continue
		}
		log.Printf("inventory processed %s", e.EventID)
	}
}
func (c *Consumer) Close() error { return c.r.Close() }

type Publisher struct{ w *kafkago.Writer }

func NewPublisher(b []string) *Publisher {
	return &Publisher{w: &kafkago.Writer{Addr: kafkago.TCP(b...), Transport: newKafkaTransport(), Balancer: &kafkago.Hash{}, RequiredAcks: kafkago.RequireAll, AllowAutoTopicCreation: false, BatchTimeout: 10 * time.Millisecond}}
}
func (p *Publisher) Publish(ctx context.Context, topic, key string, payload []byte) error {
	return p.w.WriteMessages(ctx, kafkago.Message{Topic: topic, Key: []byte(key), Value: payload, Time: time.Now().UTC()})
}
func (p *Publisher) Close() error { return p.w.Close() }
