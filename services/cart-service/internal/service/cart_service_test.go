package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/repository"
)

type fakeCartRepository struct {
	cart    domain.Cart
	setItem domain.CartItem
	removed string
	cleared bool
}

func (f *fakeCartRepository) Get(_ context.Context, customerID string) (domain.Cart, error) {
	result := f.cart
	result.CustomerID = customerID
	return result, nil
}

func (f *fakeCartRepository) SetItem(_ context.Context, _ string, item domain.CartItem) error {
	f.setItem = item
	return nil
}

func (f *fakeCartRepository) RemoveItem(_ context.Context, _, productID string) error {
	f.removed = productID
	return nil
}

func (f *fakeCartRepository) Clear(_ context.Context, _ string) error {
	f.cleared = true
	return nil
}

func TestAddItem(t *testing.T) {
	repo := &fakeCartRepository{}
	svc := NewCartService(repo)

	err := svc.AddItem(context.Background(), "customer-001", "prod-001", 2)
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}

	if repo.setItem.ProductID != "prod-001" || repo.setItem.Quantity != 2 {
		t.Fatalf("unexpected stored item: %+v", repo.setItem)
	}
}

func TestAddItemRejectsInvalidQuantity(t *testing.T) {
	repo := &fakeCartRepository{}
	svc := NewCartService(repo)

	err := svc.AddItem(context.Background(), "customer-001", "prod-001", 0)
	if !errors.Is(err, repository.ErrInvalidQuantity) {
		t.Fatalf("expected ErrInvalidQuantity, got %v", err)
	}
}

func TestRemoveItem(t *testing.T) {
	repo := &fakeCartRepository{}
	svc := NewCartService(repo)

	if err := svc.RemoveItem(context.Background(), "customer-001", "prod-002"); err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}

	if repo.removed != "prod-002" {
		t.Fatalf("expected prod-002 to be removed, got %q", repo.removed)
	}
}

func TestClearCart(t *testing.T) {
	repo := &fakeCartRepository{}
	svc := NewCartService(repo)

	if err := svc.ClearCart(context.Background(), "customer-001"); err != nil {
		t.Fatalf("ClearCart() error = %v", err)
	}

	if !repo.cleared {
		t.Fatal("expected repository Clear() to be called")
	}
}
