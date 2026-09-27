package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryRepository struct{ pool *pgxpool.Pool }

func NewInventoryRepository(p *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{pool: p}
}

func (r *InventoryRepository) ReserveForOrder(ctx context.Context, src domain.EventEnvelope, order domain.OrderCreatedPayload, topic string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM processed_events WHERE event_id=$1)`, src.EventID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return tx.Commit(ctx)
	}

	reserved := true
	reason := ""
	for _, item := range order.Items {
		var available int
		err := tx.QueryRow(ctx, `SELECT available_quantity FROM inventory WHERE product_id=$1 FOR UPDATE`, item.ProductID).Scan(&available)
		if errors.Is(err, pgx.ErrNoRows) {
			reserved = false
			reason = "inventory record not found for product " + item.ProductID
			break
		}
		if err != nil {
			return err
		}
		if available < item.Quantity {
			reserved = false
			reason = fmt.Sprintf("insufficient inventory for product %s: requested=%d available=%d", item.ProductID, item.Quantity, available)
			break
		}
	}
	if reserved {
		for _, item := range order.Items {
			if _, err := tx.Exec(ctx, `UPDATE inventory SET available_quantity=available_quantity-$2,reserved_quantity=reserved_quantity+$2,updated_at=NOW() WHERE product_id=$1`, item.ProductID, item.Quantity); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO inventory_reservations(order_id,product_id,quantity,status,reserved_at) VALUES($1,$2,$3,'reserved',NOW()) ON CONFLICT (order_id,product_id) DO NOTHING`, order.OrderID, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
	}

	now := time.Now().UTC()
	eventType := domain.InventoryReserved
	var p any = domain.InventoryReservedPayload{OrderID: order.OrderID, Items: order.Items}
	if !reserved {
		eventType = domain.InventoryRejected
		p = domain.InventoryRejectedPayload{OrderID: order.OrderID, Reason: reason}
	}
	raw, _ := json.Marshal(p)
	out := domain.EventEnvelope{EventID: src.EventID + "-inventory", EventType: eventType, EventVersion: 1, OccurredAt: now, AggregateID: order.OrderID, CorrelationID: src.CorrelationID, CausationID: src.EventID, TraceParent: src.TraceParent, Payload: raw}
	encoded, _ := json.Marshal(out)
	if _, err := tx.Exec(ctx, `INSERT INTO outbox_events(id,topic,event_key,event_type,payload,created_at) VALUES($1,$2,$3,$4,$5::jsonb,$6)`, out.EventID, topic, order.OrderID, eventType, encoded, now); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO processed_events(event_id,event_type,processed_at) VALUES($1,$2,$3)`, src.EventID, src.EventType, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *InventoryRepository) ReleaseForOrder(ctx context.Context, src domain.EventEnvelope, order domain.OrderCancelledPayload, topic string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM processed_events WHERE event_id=$1)`, src.EventID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return tx.Commit(ctx)
	}

	rows, err := tx.Query(ctx, `SELECT product_id,quantity FROM inventory_reservations WHERE order_id=$1 AND status='reserved' ORDER BY product_id FOR UPDATE`, order.OrderID)
	if err != nil {
		return err
	}
	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			rows.Close()
			return err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	now := time.Now().UTC()
	for _, item := range items {
		tag, err := tx.Exec(ctx, `UPDATE inventory SET available_quantity=available_quantity+$2,reserved_quantity=reserved_quantity-$2,updated_at=$3 WHERE product_id=$1 AND reserved_quantity >= $2`, item.ProductID, item.Quantity, now)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return fmt.Errorf("cannot release %d units of %s", item.Quantity, item.ProductID)
		}
	}
	if len(items) > 0 {
		if _, err := tx.Exec(ctx, `UPDATE inventory_reservations SET status='released',released_at=$2 WHERE order_id=$1 AND status='reserved'`, order.OrderID, now); err != nil {
			return err
		}
		payload, _ := json.Marshal(domain.InventoryReleasedPayload{OrderID: order.OrderID, Items: items, Reason: order.Reason})
		out := domain.EventEnvelope{EventID: src.EventID + "-inventory-release", EventType: domain.InventoryReleased, EventVersion: 1, OccurredAt: now, AggregateID: order.OrderID, CorrelationID: src.CorrelationID, CausationID: src.EventID, TraceParent: src.TraceParent, Payload: payload}
		encoded, _ := json.Marshal(out)
		if _, err := tx.Exec(ctx, `INSERT INTO outbox_events(id,topic,event_key,event_type,payload,created_at) VALUES($1,$2,$3,$4,$5::jsonb,$6)`, out.EventID, topic, order.OrderID, out.EventType, encoded, now); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO processed_events(event_id,event_type,processed_at) VALUES($1,$2,$3)`, src.EventID, src.EventType, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
