package kafka

import (
	"context"
	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/domain"
	"testing"
)

type handlerStub struct{ calls int }

func (h *handlerStub) HandleEvent(context.Context, domain.EventEnvelope) error { h.calls++; return nil }

func TestHandlerContract(t *testing.T) {
	var handler EventHandler = &handlerStub{}
	if err := handler.HandleEvent(context.Background(), domain.EventEnvelope{EventID: "evt-1"}); err != nil {
		t.Fatal(err)
	}
}
