package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/repository"
)

type fakePaymentRepository struct {
	payment  domain.Payment
	decision repository.PaymentDecision
	topic    string
}

func (f *fakePaymentRepository) ProcessPayment(
	_ context.Context,
	_ domain.EventEnvelope,
	payment domain.Payment,
	decision repository.PaymentDecision,
	topic string,
) error {
	f.payment = payment
	f.decision = decision
	f.topic = topic
	return nil
}

func TestHandleInventoryReservedAuthorizesPayment(t *testing.T) {
	payload, _ := json.Marshal(domain.InventoryReservedPayload{
		OrderID: "ord-001",
		Items: []domain.OrderItem{
			{ProductID: "prod-001", Quantity: 1, UnitPriceCents: 169900},
		},
	})
	repo := &fakePaymentRepository{}
	svc := NewPaymentService(repo, "payments", 500000)
	svc.now = func() time.Time { return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC) }
	svc.newID = func() (string, error) { return "pay-test", nil }

	err := svc.HandleEvent(context.Background(), domain.EventEnvelope{
		EventID: "evt-001", EventType: domain.InventoryReservedEventType,
		EventVersion: 1, AggregateID: "ord-001", Payload: payload,
	})
	if err != nil {
		t.Fatalf("HandleEvent() error = %v", err)
	}
	if !repo.decision.Authorized {
		t.Fatalf("expected authorization, reason=%s", repo.decision.Reason)
	}
	if repo.payment.Status != domain.PaymentStatusAuthorized {
		t.Fatalf("expected authorized, got %s", repo.payment.Status)
	}
	if repo.payment.AmountCents != 169900 {
		t.Fatalf("expected 169900, got %d", repo.payment.AmountCents)
	}
}

func TestHandleInventoryReservedFailsAboveLimit(t *testing.T) {
	payload, _ := json.Marshal(domain.InventoryReservedPayload{
		OrderID: "ord-002",
		Items: []domain.OrderItem{
			{ProductID: "prod-001", Quantity: 4, UnitPriceCents: 169900},
		},
	})
	repo := &fakePaymentRepository{}
	svc := NewPaymentService(repo, "payments", 500000)
	svc.newID = func() (string, error) { return "pay-test-fail", nil }

	err := svc.HandleEvent(context.Background(), domain.EventEnvelope{
		EventID: "evt-002", EventType: domain.InventoryReservedEventType,
		EventVersion: 1, AggregateID: "ord-002", Payload: payload,
	})
	if err != nil {
		t.Fatalf("HandleEvent() error = %v", err)
	}
	if repo.decision.Authorized {
		t.Fatal("expected payment failure")
	}
	if repo.payment.Status != domain.PaymentStatusFailed {
		t.Fatalf("expected failed, got %s", repo.payment.Status)
	}
}
