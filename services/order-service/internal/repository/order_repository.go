package repository

import (
	"context"
	"errors"

	"github.com/gopher-opsx/platform-lab/services/order-service/internal/domain"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderRepository interface {
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
	GetByID(ctx context.Context, id string) (domain.Order, error)
	ListByCustomer(ctx context.Context, customerID string) ([]domain.Order, error)
	ApplyPaymentEvent(
		ctx context.Context,
		sourceEvent domain.EventEnvelope,
		orderID string,
		paymentID string,
		newStatus domain.OrderStatus,
		reason string,
		ordersTopic string,
	) error
}
