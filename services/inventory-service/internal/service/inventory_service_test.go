package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/domain"
)

type fakeRepo struct{ reserved, released bool }

func (f *fakeRepo) ReserveForOrder(context.Context, domain.EventEnvelope, domain.OrderCreatedPayload, string) error {
	f.reserved = true
	return nil
}
func (f *fakeRepo) ReleaseForOrder(context.Context, domain.EventEnvelope, domain.OrderCancelledPayload, string) error {
	f.released = true
	return nil
}
func TestHandleOrderCreated(t *testing.T) {
	raw, _ := json.Marshal(domain.OrderCreatedPayload{OrderID: "ord-1"})
	r := &fakeRepo{}
	s := New(r, "inventory")
	if err := s.HandleEvent(context.Background(), domain.EventEnvelope{EventType: domain.OrderCreated, Payload: raw}); err != nil {
		t.Fatal(err)
	}
	if !r.reserved {
		t.Fatal("expected repository call")
	}
}

func TestHandleOrderCancelled(t *testing.T) {
	raw, _ := json.Marshal(domain.OrderCancelledPayload{OrderID: "ord-1", Reason: "payment failed"})
	r := &fakeRepo{}
	s := New(r, "inventory")
	if err := s.HandleEvent(context.Background(), domain.EventEnvelope{EventType: domain.OrderCancelled, Payload: raw}); err != nil {
		t.Fatal(err)
	}
	if !r.released {
		t.Fatal("expected release repository call")
	}
}

func TestIgnoreUnrelatedEvent(t *testing.T) {
	r := &fakeRepo{}
	s := New(r, "inventory")
	if err := s.HandleEvent(context.Background(), domain.EventEnvelope{EventType: "payment.authorized"}); err != nil {
		t.Fatal(err)
	}
	if r.reserved || r.released {
		t.Fatal("unexpected repository call")
	}
}
