package outbox

import (
	"context"
	"testing"
	"time"

	"github.com/gopher-opsx/platform-lab/services/order-service/internal/domain"
)

type fakeOutbox struct {
	events       []domain.OutboxEvent
	publishedIDs []string
}

func (f *fakeOutbox) ClaimBatch(_ context.Context, _ int) ([]domain.OutboxEvent, error) {
	events := f.events
	f.events = nil
	return events, nil
}
func (f *fakeOutbox) MarkPublished(_ context.Context, id string) error {
	f.publishedIDs = append(f.publishedIDs, id)
	return nil
}
func (f *fakeOutbox) RecordFailure(_ context.Context, _ string) error {
	return nil
}

type fakeMessages struct{}

func (f *fakeMessages) Publish(_ context.Context, _, _ string, _ []byte) error {
	return nil
}

func TestPublishBatchMarksSuccessfulEventPublished(t *testing.T) {
	store := &fakeOutbox{
		events: []domain.OutboxEvent{
			{ID: "evt-001", Topic: "orders", EventKey: "ord-001", Payload: []byte(`{}`)},
		},
	}

	publisher := NewPublisher(store, &fakeMessages{}, time.Second, 10)
	publisher.publishBatch(context.Background())

	if len(store.publishedIDs) != 1 || store.publishedIDs[0] != "evt-001" {
		t.Fatalf("unexpected published ids: %#v", store.publishedIDs)
	}
}
