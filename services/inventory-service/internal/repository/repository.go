package repository

import (
	"context"

	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/domain"
)

type InventoryRepository interface {
	ReserveForOrder(context.Context, domain.EventEnvelope, domain.OrderCreatedPayload, string) error
	ReleaseForOrder(context.Context, domain.EventEnvelope, domain.OrderCancelledPayload, string) error
}

type OutboxRepository interface {
	LoadBatch(context.Context, int) ([]domain.OutboxEvent, error)
	MarkPublished(context.Context, string) error
	RecordFailure(context.Context, string) error
}
