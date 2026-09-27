package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/metrics"
	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/repository"
)

type InventoryService struct {
	repo  repository.InventoryRepository
	topic string
}

func New(repo repository.InventoryRepository, topic string) *InventoryService {
	return &InventoryService{repo: repo, topic: topic}
}

func (s *InventoryService) HandleEvent(ctx context.Context, e domain.EventEnvelope) error {
	switch e.EventType {
	case domain.OrderCreated:
		var order domain.OrderCreatedPayload
		if err := json.Unmarshal(e.Payload, &order); err != nil {
			return fmt.Errorf("decode order.created: %w", err)
		}
		err := s.repo.ReserveForOrder(ctx, e, order, s.topic)
		if err == nil {
			metrics.IncBusiness("platform_lab_inventory_reserved_total")
		}
		return err
	case domain.OrderCancelled:
		var order domain.OrderCancelledPayload
		if err := json.Unmarshal(e.Payload, &order); err != nil {
			return fmt.Errorf("decode order.cancelled: %w", err)
		}
		if order.OrderID == "" {
			order.OrderID = e.AggregateID
		}
		err := s.repo.ReleaseForOrder(ctx, e, order, s.topic)
		if err == nil {
			metrics.IncBusiness("platform_lab_inventory_released_total")
		}
		return err
	default:
		return nil
	}
}
