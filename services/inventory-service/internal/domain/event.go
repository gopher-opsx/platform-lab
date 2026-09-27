package domain

import (
	"encoding/json"
	"time"
)

const (
	OrderCreated      = "order.created"
	OrderCancelled    = "order.cancelled"
	InventoryReserved = "inventory.reserved"
	InventoryRejected = "inventory.rejected"
	InventoryReleased = "inventory.released"
)

type EventEnvelope struct {
	EventID       string          `json:"eventId"`
	EventType     string          `json:"eventType"`
	EventVersion  int             `json:"eventVersion"`
	OccurredAt    time.Time       `json:"occurredAt"`
	AggregateID   string          `json:"aggregateId"`
	CorrelationID string          `json:"correlationId,omitempty"`
	CausationID   string          `json:"causationId,omitempty"`
	TraceParent   string          `json:"traceparent,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

type OrderItem struct {
	ProductID      string `json:"productId"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unitPriceCents"`
}

type OrderCreatedPayload struct {
	OrderID    string      `json:"orderId"`
	CustomerID string      `json:"customerId"`
	Currency   string      `json:"currency"`
	TotalCents int64       `json:"totalCents"`
	Items      []OrderItem `json:"items"`
}

type InventoryReservedPayload struct {
	OrderID string      `json:"orderId"`
	Items   []OrderItem `json:"items"`
}

type OrderCancelledPayload struct {
	OrderID   string `json:"orderId"`
	PaymentID string `json:"paymentId"`
	Reason    string `json:"reason,omitempty"`
}

type InventoryReleasedPayload struct {
	OrderID string      `json:"orderId"`
	Items   []OrderItem `json:"items"`
	Reason  string      `json:"reason,omitempty"`
}
type InventoryRejectedPayload struct {
	OrderID string `json:"orderId"`
	Reason  string `json:"reason"`
}
