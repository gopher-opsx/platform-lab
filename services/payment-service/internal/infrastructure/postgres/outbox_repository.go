package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/domain"
)

type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

func (r *OutboxRepository) LoadBatch(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, topic, event_key, event_type, payload, created_at, published_at, attempts
		FROM outbox_events
		WHERE published_at IS NULL
		ORDER BY created_at
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("load payment outbox batch: %w", err)
	}
	defer rows.Close()

	events := make([]domain.OutboxEvent, 0)
	for rows.Next() {
		var event domain.OutboxEvent
		if err := rows.Scan(
			&event.ID, &event.Topic, &event.EventKey, &event.EventType,
			&event.Payload, &event.CreatedAt, &event.PublishedAt, &event.Attempts,
		); err != nil {
			return nil, fmt.Errorf("scan payment outbox event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payment outbox events: %w", err)
	}
	return events, nil
}

func (r *OutboxRepository) MarkPublished(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox_events SET published_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark payment outbox event published: %w", err)
	}
	return nil
}

func (r *OutboxRepository) RecordFailure(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox_events SET attempts = attempts + 1 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("record payment outbox failure: %w", err)
	}
	return nil
}
