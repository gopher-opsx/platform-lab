package repository

import (
	"context"

	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/domain"
)

type PaymentDecision struct {
	Authorized bool
	Reason     string
}

type PaymentRepository interface {
	ProcessPayment(
		ctx context.Context,
		sourceEvent domain.EventEnvelope,
		payment domain.Payment,
		decision PaymentDecision,
		paymentsTopic string,
	) error
}
