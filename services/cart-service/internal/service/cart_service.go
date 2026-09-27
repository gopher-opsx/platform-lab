package service

import (
	"context"
	"strings"

	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/repository"
)

type CartService struct {
	carts repository.CartRepository
}

func NewCartService(carts repository.CartRepository) *CartService {
	return &CartService{carts: carts}
}

func (s *CartService) GetCart(ctx context.Context, customerID string) (domain.Cart, error) {
	return s.carts.Get(ctx, strings.TrimSpace(customerID))
}

func (s *CartService) AddItem(ctx context.Context, customerID, productID string, quantity int) error {
	if quantity <= 0 {
		return repository.ErrInvalidQuantity
	}

	return s.carts.SetItem(ctx, strings.TrimSpace(customerID), domain.CartItem{
		ProductID: strings.TrimSpace(productID),
		Quantity:  quantity,
	})
}

func (s *CartService) UpdateItem(ctx context.Context, customerID, productID string, quantity int) error {
	return s.AddItem(ctx, customerID, productID, quantity)
}

func (s *CartService) RemoveItem(ctx context.Context, customerID, productID string) error {
	return s.carts.RemoveItem(ctx, strings.TrimSpace(customerID), strings.TrimSpace(productID))
}

func (s *CartService) ClearCart(ctx context.Context, customerID string) error {
	return s.carts.Clear(ctx, strings.TrimSpace(customerID))
}
