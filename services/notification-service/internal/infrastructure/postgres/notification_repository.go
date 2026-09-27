package postgres

import (
	"context"
	"fmt"

	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct{ pool *pgxpool.Pool }

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) StoreDelivered(ctx context.Context, event domain.EventEnvelope, n domain.Notification) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin notification transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `INSERT INTO processed_events (event_id, event_type, processed_at) VALUES ($1,$2,now()) ON CONFLICT (event_id) DO NOTHING`, event.EventID, event.EventType)
	if err != nil {
		return false, fmt.Errorf("record processed event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	_, err = tx.Exec(ctx, `INSERT INTO notifications (id,order_id,source_event_id,source_event_type,channel,recipient,subject,body,status,delivered_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, n.ID, n.OrderID, n.SourceEventID, n.SourceEventType, n.Channel, n.Recipient, n.Subject, n.Body, n.Status, n.DeliveredAt)
	if err != nil {
		return false, fmt.Errorf("insert notification: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO delivery_log (notification_id,attempt_number,status,provider,message,attempted_at) VALUES ($1,1,$2,'simulated',$3,$4)`, n.ID, n.Status, "simulated delivery accepted", n.DeliveredAt)
	if err != nil {
		return false, fmt.Errorf("insert delivery log: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit notification transaction: %w", err)
	}
	return true, nil
}
