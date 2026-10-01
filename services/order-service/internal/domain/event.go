package domain

import (
	"encoding/json"
	"time"
)

const (
	OrderCreatedEventType      = "order.created"
	OrderConfirmedEventType    = "order.confirmed"
	OrderCancelledEventType    = "order.cancelled"
	PaymentAuthorizedEventType = "payment.authorized"
	PaymentFailedEventType     = "payment.failed"
	InventoryRejectedEventType = "inventory.rejected"
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

type OrderCreatedPayload struct {
	OrderID    string      `json:"orderId"`
	CustomerID string      `json:"customerId"`
	Currency   string      `json:"currency"`
	TotalCents int64       `json:"totalCents"`
	Items      []OrderItem `json:"items"`
}

type PaymentAuthorizedPayload struct {
	OrderID     string `json:"orderId"`
	PaymentID   string `json:"paymentId"`
	AmountCents int64  `json:"amountCents"`
	Currency    string `json:"currency"`
}

type PaymentFailedPayload struct {
	OrderID     string `json:"orderId"`
	PaymentID   string `json:"paymentId"`
	AmountCents int64  `json:"amountCents"`
	Currency    string `json:"currency"`
	Reason      string `json:"reason"`
}

type InventoryRejectedPayload struct {
	OrderID string `json:"orderId"`
	Reason  string `json:"reason"`
}

type OrderConfirmedPayload struct {
	OrderID   string `json:"orderId"`
	PaymentID string `json:"paymentId"`
}

type OrderCancelledPayload struct {
	OrderID   string `json:"orderId"`
	PaymentID string `json:"paymentId"`
	Reason    string `json:"reason"`
}
