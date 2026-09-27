package domain

import "time"

type PaymentStatus string

const (
	PaymentStatusAuthorized PaymentStatus = "authorized"
	PaymentStatusFailed     PaymentStatus = "failed"
)

type Payment struct {
	ID            string        `json:"id"`
	OrderID       string        `json:"orderId"`
	AmountCents   int64         `json:"amountCents"`
	Currency      string        `json:"currency"`
	Status        PaymentStatus `json:"status"`
	FailureReason string        `json:"failureReason,omitempty"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}
