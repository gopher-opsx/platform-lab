package outbox

import (
	"context"
	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/repository"
	"log"
	"time"
)

type MessagePublisher interface {
	Publish(context.Context, string, string, []byte) error
}
type Publisher struct {
	repo repository.OutboxRepository
	pub  MessagePublisher
	poll time.Duration
}

func New(repo repository.OutboxRepository, pub MessagePublisher, poll time.Duration) *Publisher {
	return &Publisher{repo: repo, pub: pub, poll: poll}
}
func (p *Publisher) Run(ctx context.Context) {
	t := time.NewTicker(p.poll)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.once(ctx)
		}
	}
}
func (p *Publisher) once(ctx context.Context) {
	events, err := p.repo.LoadBatch(ctx, 50)
	if err != nil {
		log.Printf("inventory outbox load: %v", err)
		return
	}
	for _, e := range events {
		if err := p.pub.Publish(ctx, e.Topic, e.EventKey, e.Payload); err != nil {
			_ = p.repo.RecordFailure(ctx, e.ID)
			continue
		}
		_ = p.repo.MarkPublished(ctx, e.ID)
		log.Printf("inventory outbox published %s (%s)", e.ID, e.EventType)
	}
}
