package domain

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ProductID      string `json:"productId"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unitPriceCents"`
}

type Order struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customerId"`
	Status     OrderStatus `json:"status"`
	Currency   string      `json:"currency"`
	TotalCents int64       `json:"totalCents"`
	Items      []OrderItem `json:"items"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}
