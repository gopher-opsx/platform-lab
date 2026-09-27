package outbox

import (
	"context"
	"log"
	"time"

	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/repository"
)

type MessagePublisher interface {
	Publish(ctx context.Context, topic, key string, payload []byte) error
}

type Publisher struct {
	outbox    repository.OutboxRepository
	messages  MessagePublisher
	poll      time.Duration
	batchSize int
}

func NewPublisher(outbox repository.OutboxRepository, messages MessagePublisher, poll time.Duration, batchSize int) *Publisher {
	return &Publisher{outbox: outbox, messages: messages, poll: poll, batchSize: batchSize}
}

func (p *Publisher) Run(ctx context.Context) {
	ticker := time.NewTicker(p.poll)
	defer ticker.Stop()
	p.publishBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.publishBatch(ctx)
		}
	}
}

func (p *Publisher) publishBatch(ctx context.Context) {
	events, err := p.outbox.LoadBatch(ctx, p.batchSize)
	if err != nil {
		log.Printf("payment outbox: load batch failed: %v", err)
		return
	}
	for _, event := range events {
		if err := p.messages.Publish(ctx, event.Topic, event.EventKey, event.Payload); err != nil {
			log.Printf("payment outbox: publish %s failed: %v", event.ID, err)
			_ = p.outbox.RecordFailure(ctx, event.ID)
			continue
		}
		if err := p.outbox.MarkPublished(ctx, event.ID); err != nil {
			log.Printf("payment outbox: mark published %s failed: %v", event.ID, err)
			continue
		}
		log.Printf("payment outbox: published %s (%s)", event.ID, event.EventType)
	}
}
