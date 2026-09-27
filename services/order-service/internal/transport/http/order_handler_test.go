package httptransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gopher-opsx/platform-lab/services/order-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/order-service/internal/repository"
	"github.com/gopher-opsx/platform-lab/services/order-service/internal/service"
)

type handlerOrderRepository struct{ order domain.Order }

func (r *handlerOrderRepository) Create(_ context.Context, o domain.Order) (domain.Order, error) {
	r.order = o
	return o, nil
}
func (r *handlerOrderRepository) GetByID(_ context.Context, id string) (domain.Order, error) {
	if r.order.ID != id {
		return domain.Order{}, repository.ErrOrderNotFound
	}
	return r.order, nil
}
func (r *handlerOrderRepository) ListByCustomer(_ context.Context, id string) ([]domain.Order, error) {
	if r.order.CustomerID != id {
		return []domain.Order{}, nil
	}
	return []domain.Order{r.order}, nil
}
func TestCreateOrderRequiresCustomerHeader(t *testing.T) {
	h := NewOrderHandler(service.NewOrderService(&handlerOrderRepository{}))
	q := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"currency":"USD","items":[{"productId":"prod-001","quantity":1,"unitPriceCents":1000}]}`))
	w := httptest.NewRecorder()
	h.CreateOrder(w, q)
	if w.Code != 400 {
		t.Fatalf("expected 400 got %d", w.Code)
	}
}
func TestCreateOrderReturnsCreated(t *testing.T) {
	h := NewOrderHandler(service.NewOrderService(&handlerOrderRepository{}))
	q := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"currency":"USD","items":[{"productId":"prod-001","quantity":2,"unitPriceCents":1000}]}`))
	q.Header.Set(customerIDHeader, "customer-001")
	w := httptest.NewRecorder()
	h.CreateOrder(w, q)
	if w.Code != 201 {
		t.Fatalf("expected 201 got %d body=%s", w.Code, w.Body.String())
	}
}

func (r *handlerOrderRepository) ApplyPaymentEvent(
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
