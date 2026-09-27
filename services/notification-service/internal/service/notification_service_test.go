package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/domain"
)

type fakeRepository struct {
	saved []domain.Notification
	seen  map[string]bool
}

func (f *fakeRepository) StoreDelivered(_ context.Context, event domain.EventEnvelope, notification domain.Notification) (bool, error) {
	if f.seen == nil {
		f.seen = map[string]bool{}
	}
	if f.seen[event.EventID] {
		return false, nil
	}
	f.seen[event.EventID] = true
	f.saved = append(f.saved, notification)
	return true, nil
}

func event(t *testing.T, eventType string, payload any) domain.EventEnvelope {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return domain.EventEnvelope{EventID: "evt-1", EventType: eventType, AggregateID: "ord-1", Payload: raw}
}

func TestConfirmedOrderCreatesDeliveredNotification(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewNotificationService(repo)
	fixed := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixed }
	err := svc.HandleEvent(context.Background(), event(t, domain.OrderConfirmedEventType, domain.OrderEventPayload{OrderID: "ord-1", PaymentID: "pay-1"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("saved = %d, want 1", len(repo.saved))
	}
	got := repo.saved[0]
	if got.Status != "delivered" || got.DeliveredAt != fixed || got.OrderID != "ord-1" {
		t.Fatalf("unexpected notification: %#v", got)
	}
}

func TestCancelledOrderUsesReason(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewNotificationService(repo)
	err := svc.HandleEvent(context.Background(), event(t, domain.OrderCancelledEventType, domain.OrderEventPayload{OrderID: "ord-1", Reason: "card declined"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.saved) != 1 || repo.saved[0].Subject == "" {
		t.Fatalf("unexpected notifications: %#v", repo.saved)
	}
}

func TestDuplicateEventIsIdempotent(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewNotificationService(repo)
	e := event(t, domain.OrderConfirmedEventType, domain.OrderEventPayload{OrderID: "ord-1"})
	if err := svc.HandleEvent(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	if err := svc.HandleEvent(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("saved = %d, want 1", len(repo.saved))
	}
}

func TestUnrelatedEventIsIgnored(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewNotificationService(repo)
	if err := svc.HandleEvent(context.Background(), event(t, "order.created", map[string]string{"orderId": "ord-1"})); err != nil {
		t.Fatal(err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("saved = %d, want 0", len(repo.saved))
	}
}
