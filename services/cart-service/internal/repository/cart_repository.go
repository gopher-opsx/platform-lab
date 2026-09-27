package repository

import (
	"context"
	"errors"

	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/domain"
)

var ErrInvalidQuantity = errors.New("quantity must be greater than zero")

type CartRepository interface {
	Get(ctx context.Context, customerID string) (domain.Cart, error)
	SetItem(ctx context.Context, customerID string, item domain.CartItem) error
	RemoveItem(ctx context.Context, customerID, productID string) error
	Clear(ctx context.Context, customerID string) error
}
