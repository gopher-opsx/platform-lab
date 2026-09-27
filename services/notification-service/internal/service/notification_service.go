package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/domain"
	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/metrics"
	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/repository"
)

type NotificationService struct {
	repository repository.NotificationRepository
	now        func() time.Time
}

func NewNotificationService(repository repository.NotificationRepository) *NotificationService {
	return &NotificationService{repository: repository, now: time.Now}
}

func (s *NotificationService) HandleEvent(ctx context.Context, event domain.EventEnvelope) error {
	if event.EventID == "" {
		return errors.New("eventId is required")
	}
	if event.EventType != domain.OrderConfirmedEventType && event.EventType != domain.OrderCancelledEventType {
		return nil
	}

	var payload domain.OrderEventPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("decode order event: %w", err)
	}
	if payload.OrderID == "" {
		payload.OrderID = event.AggregateID
	}
	if payload.OrderID == "" {
		return errors.New("orderId is required")
	}

	n := domain.Notification{
		ID: event.EventID + "-notification", OrderID: payload.OrderID,
		SourceEventID: event.EventID, SourceEventType: event.EventType,
		Channel: "email", Recipient: "customer:" + payload.OrderID,
		Status: "delivered", DeliveredAt: s.now().UTC(),
	}
	if event.EventType == domain.OrderConfirmedEventType {
		n.Subject = "Your order is confirmed"
		n.Body = fmt.Sprintf("Order %s has been confirmed. Payment reference: %s.", payload.OrderID, payload.PaymentID)
	} else {
		n.Subject = "Your order was cancelled"
		n.Body = fmt.Sprintf("Order %s was cancelled. Reason: %s.", payload.OrderID, fallback(payload.Reason, "payment could not be completed"))
	}

	created, err := s.repository.StoreDelivered(ctx, event, n)
	if err == nil && created {
		metrics.IncBusiness("platform_lab_notifications_delivered_total")
	}
	return err
}

func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
