package repository

import (
	"context"

	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/domain"
)

type OutboxRepository interface {
	LoadBatch(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkPublished(ctx context.Context, id string) error
	RecordFailure(ctx context.Context, id string) error
}
