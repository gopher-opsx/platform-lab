package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gopher-opsx/platform-lab/services/order-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/order-service/internal/repository"
)

type fakeOrderRepository struct{ created domain.Order }

func (f *fakeOrderRepository) Create(_ context.Context, o domain.Order) (domain.Order, error) {
	f.created = o
	return o, nil
}
func (f *fakeOrderRepository) GetByID(_ context.Context, _ string) (domain.Order, error) {
	return domain.Order{}, repository.ErrOrderNotFound
}
func (f *fakeOrderRepository) ListByCustomer(_ context.Context, _ string) ([]domain.Order, error) {
	return nil, nil
}
func TestCreateOrderCalculatesTotalAndStartsPending(t *testing.T) {
	r := &fakeOrderRepository{}
	s := NewOrderService(r)
	s.now = func() time.Time { return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC) }
	s.newID = func() (string, error) { return "ord-test", nil }
	o, e := s.CreateOrder(context.Background(), CreateOrderInput{CustomerID: "customer-001", Currency: "usd", Items: []domain.OrderItem{{ProductID: "prod-001", Quantity: 2, UnitPriceCents: 1000}, {ProductID: "prod-002", Quantity: 1, UnitPriceCents: 500}}})
	if e != nil {
		t.Fatal(e)
	}
	if o.TotalCents != 2500 || o.Status != domain.OrderStatusPending || o.Currency != "USD" {
		t.Fatalf("unexpected order: %+v", o)
	}
}
func TestCreateOrderRejectsEmptyItems(t *testing.T) {
	s := NewOrderService(&fakeOrderRepository{})
	_, e := s.CreateOrder(context.Background(), CreateOrderInput{CustomerID: "customer-001", Currency: "USD"})
	if !errors.Is(e, ErrItemsRequired) {
		t.Fatalf("expected ErrItemsRequired, got %v", e)
	}
}

func (r *fakeOrderRepository) ApplyPaymentEvent(
	context.Context,
	domain.EventEnvelope,
	string,
	string,
	domain.OrderStatus,
	string,
	string,
) error {
	return nil
}
