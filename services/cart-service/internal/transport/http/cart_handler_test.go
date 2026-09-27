package httptransport

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/service"
)

type handlerRepo struct {
	items map[string]domain.CartItem
}

func newHandlerRepo() *handlerRepo {
	return &handlerRepo{items: map[string]domain.CartItem{}}
}

func (r *handlerRepo) Get(_ context.Context, customerID string) (domain.Cart, error) {
	items := make([]domain.CartItem, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return domain.Cart{CustomerID: customerID, Items: items}, nil
}

func (r *handlerRepo) SetItem(_ context.Context, _ string, item domain.CartItem) error {
	r.items[item.ProductID] = item
	return nil
}

func (r *handlerRepo) RemoveItem(_ context.Context, _, productID string) error {
	delete(r.items, productID)
	return nil
}

func (r *handlerRepo) Clear(_ context.Context, _ string) error {
	r.items = map[string]domain.CartItem{}
	return nil
}

func TestAddItemRequiresCustomerHeader(t *testing.T) {
	handler := NewCartHandler(service.NewCartService(newHandlerRepo()))
	request := httptest.NewRequest(http.MethodPost, "/cart/items", strings.NewReader(`{"productId":"prod-001","quantity":1}`))
	response := httptest.NewRecorder()

	handler.AddItem(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestAddAndGetCart(t *testing.T) {
	repo := newHandlerRepo()
	handler := NewCartHandler(service.NewCartService(repo))

	addRequest := httptest.NewRequest(http.MethodPost, "/cart/items", bytes.NewBufferString(`{"productId":"prod-001","quantity":2}`))
	addRequest.Header.Set(customerIDHeader, "customer-001")
	addResponse := httptest.NewRecorder()
	handler.AddItem(addResponse, addRequest)

	if addResponse.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", addResponse.Code)
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/cart", nil)
	getRequest.Header.Set(customerIDHeader, "customer-001")
	getResponse := httptest.NewRecorder()
	handler.GetCart(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getResponse.Code)
	}

	if !strings.Contains(getResponse.Body.String(), `"productId":"prod-001"`) {
		t.Fatalf("unexpected cart body: %s", getResponse.Body.String())
	}
}
