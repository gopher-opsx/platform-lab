package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/repository"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) ProcessPayment(
	ctx context.Context,
	sourceEvent domain.EventEnvelope,
	payment domain.Payment,
	decision repository.PaymentDecision,
	paymentsTopic string,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin payment transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var alreadyProcessed bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM processed_events WHERE event_id = $1
		)
	`, sourceEvent.EventID).Scan(&alreadyProcessed)
	if err != nil {
		return fmt.Errorf("check processed payment event: %w", err)
	}
	if alreadyProcessed {
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit duplicate payment event transaction: %w", err)
		}
		return nil
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payments (
			id, order_id, amount_cents, currency, status, failure_reason, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`,
		payment.ID, payment.OrderID, payment.AmountCents, payment.Currency,
		payment.Status, payment.FailureReason, payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}

	eventType := domain.PaymentAuthorizedEventType
	var eventPayload any = domain.PaymentAuthorizedPayload{
		OrderID: payment.OrderID, PaymentID: payment.ID,
		AmountCents: payment.AmountCents, Currency: payment.Currency,
	}
	if !decision.Authorized {
		eventType = domain.PaymentFailedEventType
		eventPayload = domain.PaymentFailedPayload{
			OrderID: payment.OrderID, PaymentID: payment.ID,
			AmountCents: payment.AmountCents, Currency: payment.Currency,
			Reason: decision.Reason,
		}
	}

	payloadJSON, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("marshal payment event payload: %w", err)
	}

	outgoing := domain.EventEnvelope{
		EventID:       sourceEvent.EventID + "-payment",
		EventType:     eventType,
		EventVersion:  1,
		OccurredAt:    payment.CreatedAt,
		AggregateID:   payment.OrderID,
		CorrelationID: sourceEvent.CorrelationID,
		CausationID:   sourceEvent.EventID,
		TraceParent:   sourceEvent.TraceParent,
		Payload:       payloadJSON,
	}
	outgoingJSON, err := json.Marshal(outgoing)
	if err != nil {
		return fmt.Errorf("marshal payment event: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (
			id, topic, event_key, event_type, payload, created_at
		)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6)
	`,
		outgoing.EventID, paymentsTopic, payment.OrderID,
		eventType, outgoingJSON, payment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert payment outbox event: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO processed_events (event_id, event_type, processed_at)
		VALUES ($1,$2,$3)
	`, sourceEvent.EventID, sourceEvent.EventType, payment.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert payment processed event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit payment transaction: %w", err)
	}
	return nil
}
