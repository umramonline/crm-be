package application

import (
	"context"
	"errors"
	"testing"

	"github.com/umran/new.crm/backend/internal/consume/domain"
)

type fakeConsumeRepository struct {
	result domain.ConsumeResult
	err    error
	event  domain.CustomerCreatedEvent
}

func (f *fakeConsumeRepository) ConsumeCustomerCreated(_ context.Context, event domain.CustomerCreatedEvent) (domain.ConsumeResult, error) {
	f.event = event

	return f.result, f.err
}

func TestConsumeRejectsUnsupportedEventType(t *testing.T) {
	service := NewService(&fakeConsumeRepository{})

	_, err := service.Consume(context.Background(), domain.ConsumeCommand{
		EventID:   "event-1",
		EventType: "customer.updated",
		Payload:   []byte(`{"event_id":"event-1","event_type":"customer.updated","uo_id":10}`),
	})
	if !errors.Is(err, ErrUnsupportedEventType) {
		t.Fatalf("expected ErrUnsupportedEventType, got %v", err)
	}
}

func TestConsumeRejectsMissingEventID(t *testing.T) {
	service := NewService(&fakeConsumeRepository{})

	_, err := service.Consume(context.Background(), domain.ConsumeCommand{
		EventType: domain.EventTypeCustomerCreated,
		Payload:   []byte(`{"event_type":"customer.created","uo_id":10}`),
	})
	if !errors.Is(err, ErrInvalidEventPayload) {
		t.Fatalf("expected ErrInvalidEventPayload, got %v", err)
	}
}

func TestConsumeDelegatesCustomerCreated(t *testing.T) {
	repository := &fakeConsumeRepository{
		result: domain.ConsumeResult{
			EventID:    "event-1",
			CustomerID: 99,
			Action:     "created",
		},
	}
	service := NewService(repository)

	result, err := service.Consume(context.Background(), domain.ConsumeCommand{
		EventID:   "event-1",
		EventType: domain.EventTypeCustomerCreated,
		Payload: []byte(`{
			"event_id":"event-1",
			"event_type":"customer.created",
			"uo_id":10,
			"telefon":"05550000000"
		}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CustomerID != 99 {
		t.Fatalf("expected customer id 99, got %d", result.CustomerID)
	}

	if repository.event.UOId != 10 {
		t.Fatalf("expected repository to receive uo_id 10, got %d", repository.event.UOId)
	}

	if repository.event.EventID != "event-1" {
		t.Fatalf("expected repository to receive event_id event-1, got %q", repository.event.EventID)
	}
}
