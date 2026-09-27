package domain

import (
	"encoding/json"
	"time"
)

const (
	OrderConfirmedEventType = "order.confirmed"
	OrderCancelledEventType = "order.cancelled"
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

type OrderEventPayload struct {
	OrderID   string `json:"orderId"`
	PaymentID string `json:"paymentId"`
	Reason    string `json:"reason,omitempty"`
}
