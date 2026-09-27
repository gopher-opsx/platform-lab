package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/metrics"
	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/repository"
)

var ErrUnsupportedEvent = errors.New("unsupported event type")

type PaymentService struct {
	payments      repository.PaymentRepository
	paymentsTopic string
	maxAuthCents  int64
	now           func() time.Time
	newID         func() (string, error)
}

func NewPaymentService(payments repository.PaymentRepository, paymentsTopic string, maxAuthCents int64) *PaymentService {
	return &PaymentService{
		payments:      payments,
		paymentsTopic: paymentsTopic,
		maxAuthCents:  maxAuthCents,
		now:           time.Now,
		newID:         generatePaymentID,
	}
}

func (s *PaymentService) HandleEvent(ctx context.Context, envelope domain.EventEnvelope) error {
	if envelope.EventType != domain.InventoryReservedEventType {
		return ErrUnsupportedEvent
	}

	var reserved domain.InventoryReservedPayload
	if err := json.Unmarshal(envelope.Payload, &reserved); err != nil {
		return fmt.Errorf("decode inventory.reserved payload: %w", err)
	}

	var amount int64
	for _, item := range reserved.Items {
		if item.Quantity <= 0 || item.UnitPriceCents < 0 {
			return fmt.Errorf("invalid payment item for product %s", item.ProductID)
		}
		amount += int64(item.Quantity) * item.UnitPriceCents
	}

	paymentID, err := s.newID()
	if err != nil {
		return err
	}

	decision := repository.PaymentDecision{Authorized: true}
	status := domain.PaymentStatusAuthorized
	failureReason := ""

	if amount > s.maxAuthCents {
		decision.Authorized = false
		decision.Reason = fmt.Sprintf("simulated authorization limit exceeded: amount=%d limit=%d", amount, s.maxAuthCents)
		status = domain.PaymentStatusFailed
		failureReason = decision.Reason
	}

	now := s.now().UTC()
	payment := domain.Payment{
		ID:            paymentID,
		OrderID:       reserved.OrderID,
		AmountCents:   amount,
		Currency:      "USD",
		Status:        status,
		FailureReason: failureReason,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err = s.payments.ProcessPayment(ctx, envelope, payment, decision, s.paymentsTopic)
	if err == nil {
		if decision.Authorized {
			metrics.IncBusiness("platform_lab_payments_authorized_total")
		} else {
			metrics.IncBusiness("platform_lab_payments_failed_total")
		}
	}
	return err
}

func generatePaymentID() (string, error) {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return "pay-" + hex.EncodeToString(raw[:]), nil
}
