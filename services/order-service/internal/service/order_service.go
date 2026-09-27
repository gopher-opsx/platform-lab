package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gopher-opsx/platform-lab/services/order-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/order-service/internal/metrics"
	"github.com/gopher-opsx/platform-lab/services/order-service/internal/repository"
)

var (
	ErrCustomerRequired = errors.New("customer is required")
	ErrItemsRequired    = errors.New("at least one order item is required")
	ErrProductRequired  = errors.New("productId is required")
	ErrInvalidQuantity  = errors.New("quantity must be greater than zero")
	ErrInvalidPrice     = errors.New("unit price must not be negative")
	ErrCurrencyRequired = errors.New("currency is required")
)

type CreateOrderInput struct {
	CustomerID string
	Currency   string
	Items      []domain.OrderItem
}

type OrderService struct {
	orders repository.OrderRepository
	now    func() time.Time
	newID  func() (string, error)
}

func NewOrderService(orders repository.OrderRepository) *OrderService {
	return &OrderService{
		orders: orders,
		now:    time.Now,
		newID:  generateOrderID,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, input CreateOrderInput) (domain.Order, error) {
	input.CustomerID = strings.TrimSpace(input.CustomerID)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))

	if input.CustomerID == "" {
		return domain.Order{}, ErrCustomerRequired
	}
	if input.Currency == "" {
		return domain.Order{}, ErrCurrencyRequired
	}
	if len(input.Items) == 0 {
		return domain.Order{}, ErrItemsRequired
	}

	var total int64
	items := make([]domain.OrderItem, len(input.Items))
	for i, item := range input.Items {
		item.ProductID = strings.TrimSpace(item.ProductID)
		if item.ProductID == "" {
			return domain.Order{}, ErrProductRequired
		}
		if item.Quantity <= 0 {
			return domain.Order{}, ErrInvalidQuantity
		}
		if item.UnitPriceCents < 0 {
			return domain.Order{}, ErrInvalidPrice
		}
		total += int64(item.Quantity) * item.UnitPriceCents
		items[i] = item
	}

	id, err := s.newID()
	if err != nil {
		return domain.Order{}, err
	}

	now := s.now().UTC()
	order := domain.Order{
		ID:         id,
		CustomerID: input.CustomerID,
		Status:     domain.OrderStatusPending,
		Currency:   input.Currency,
		TotalCents: total,
		Items:      items,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	created, err := s.orders.Create(ctx, order)
	if err == nil {
		metrics.IncBusiness("platform_lab_orders_created_total")
	}
	return created, err
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (domain.Order, error) {
	return s.orders.GetByID(ctx, strings.TrimSpace(id))
}

func (s *OrderService) ListOrders(ctx context.Context, customerID string) ([]domain.Order, error) {
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return nil, ErrCustomerRequired
	}
	return s.orders.ListByCustomer(ctx, customerID)
}

func generateOrderID() (string, error) {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return "ord-" + hex.EncodeToString(raw[:]), nil
}
