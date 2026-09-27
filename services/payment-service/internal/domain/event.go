package domain

import (
	"encoding/json"
	"time"
)

const (
	InventoryReservedEventType = "inventory.reserved"
	PaymentAuthorizedEventType = "payment.authorized"
	PaymentFailedEventType     = "payment.failed"
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

type InventoryReservedPayload struct {
	OrderID string      `json:"orderId"`
	Items   []OrderItem `json:"items"`
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
