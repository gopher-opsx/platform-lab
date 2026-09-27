package postgres

import (
	"context"
	"fmt"
	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepository struct{ pool *pgxpool.Pool }

func NewOutboxRepository(p *pgxpool.Pool) *OutboxRepository { return &OutboxRepository{pool: p} }
func (r *OutboxRepository) LoadBatch(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,topic,event_key,event_type,payload,created_at,published_at,attempts FROM outbox_events WHERE published_at IS NULL ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []domain.OutboxEvent{}
	for rows.Next() {
		var e domain.OutboxEvent
		if err := rows.Scan(&e.ID, &e.Topic, &e.EventKey, &e.EventType, &e.Payload, &e.CreatedAt, &e.PublishedAt, &e.Attempts); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
func (r *OutboxRepository) MarkPublished(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox_events SET published_at=NOW() WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("mark published: %w", err)
	}
	return nil
}
func (r *OutboxRepository) RecordFailure(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox_events SET attempts=attempts+1 WHERE id=$1`, id)
	return err
}
