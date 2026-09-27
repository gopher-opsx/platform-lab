package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gopher-opsx/platform-lab/services/order-service/internal/domain"
)

type paymentEventRepo struct {
	orderID   string
	paymentID string
	status    domain.OrderStatus
	reason    string
	topic     string
}

func (r *paymentEventRepo) Create(context.Context, domain.Order) (domain.Order, error) {
	return domain.Order{}, nil
}
func (r *paymentEventRepo) GetByID(context.Context, string) (domain.Order, error) {
	return domain.Order{}, nil
}
func (r *paymentEventRepo) ListByCustomer(context.Context, string) ([]domain.Order, error) {
	return nil, nil
}
func (r *paymentEventRepo) ApplyPaymentEvent(
	_ context.Context,
	_ domain.EventEnvelope,
	orderID string,
	paymentID string,
	status domain.OrderStatus,
	reason string,
	topic string,
) error {
	r.orderID = orderID
	r.paymentID = paymentID
	r.status = status
	r.reason = reason
	r.topic = topic
	return nil
}

func TestPaymentAuthorizedConfirmsOrder(t *testing.T) {
	payload, _ := json.Marshal(domain.PaymentAuthorizedPayload{
		OrderID:   "ord-001",
		PaymentID: "pay-001",
	})

	repo := &paymentEventRepo{}
	svc := NewPaymentEventService(repo, "orders")

	err := svc.HandleEvent(context.Background(), domain.EventEnvelope{
		EventID:   "evt-payment-001",
		EventType: domain.PaymentAuthorizedEventType,
		Payload:   payload,
	})
	if err != nil {
		t.Fatalf("HandleEvent() error = %v", err)
	}

	if repo.status != domain.OrderStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", repo.status)
	}
	if repo.orderID != "ord-001" || repo.paymentID != "pay-001" {
		t.Fatalf("unexpected identifiers: order=%s payment=%s", repo.orderID, repo.paymentID)
	}
}

func TestPaymentFailedCancelsOrder(t *testing.T) {
	payload, _ := json.Marshal(domain.PaymentFailedPayload{
		OrderID:   "ord-002",
		PaymentID: "pay-002",
		Reason:    "declined",
	})

	repo := &paymentEventRepo{}
	svc := NewPaymentEventService(repo, "orders")

	err := svc.HandleEvent(context.Background(), domain.EventEnvelope{
		EventID:   "evt-payment-002",
		EventType: domain.PaymentFailedEventType,
		Payload:   payload,
	})
	if err != nil {
		t.Fatalf("HandleEvent() error = %v", err)
	}

	if repo.status != domain.OrderStatusCancelled {
		t.Fatalf("expected cancelled, got %s", repo.status)
	}
	if repo.reason != "declined" {
		t.Fatalf("expected declined reason, got %s", repo.reason)
	}
}
