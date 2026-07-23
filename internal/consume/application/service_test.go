package application

import (
	"context"
	"errors"
	"testing"

	"github.com/umran/new.crm/backend/internal/consume/domain"
)

type fakeConsumeRepository struct {
	createdResult domain.ConsumeResult
	updatedResult domain.ConsumeResult
	deletedResult domain.ConsumeResult
	createdErr    error
	updatedErr    error
	deletedErr    error
	createdEvent  domain.CustomerCreatedEvent
	updatedEvent  domain.CustomerUpdatedEvent
	deletedEvent  domain.CustomerDeletedEvent
}

func (f *fakeConsumeRepository) ConsumeCustomerCreated(_ context.Context, event domain.CustomerCreatedEvent) (domain.ConsumeResult, error) {
	f.createdEvent = event

	return f.createdResult, f.createdErr
}

func (f *fakeConsumeRepository) ConsumeCustomerUpdated(_ context.Context, event domain.CustomerUpdatedEvent) (domain.ConsumeResult, error) {
	f.updatedEvent = event

	return f.updatedResult, f.updatedErr
}

func (f *fakeConsumeRepository) ConsumeCustomerDeleted(_ context.Context, event domain.CustomerDeletedEvent) (domain.ConsumeResult, error) {
	f.deletedEvent = event

	return f.deletedResult, f.deletedErr
}

func TestConsumeRejectsUnsupportedEventType(t *testing.T) {
	service := NewService(&fakeConsumeRepository{})

	_, err := service.Consume(context.Background(), domain.ConsumeCommand{
		EventID:   "event-1",
		EventType: "customer.merged",
		Payload:   []byte(`{"event_id":"event-1","event_type":"customer.merged","uo_id":10,"occurred_at":"2026-07-22T15:22:50+03:00"}`),
	})
	if !errors.Is(err, ErrUnsupportedEventType) {
		t.Fatalf("expected ErrUnsupportedEventType, got %v", err)
	}
}

func TestConsumeRejectsMissingEventID(t *testing.T) {
	service := NewService(&fakeConsumeRepository{})

	_, err := service.Consume(context.Background(), domain.ConsumeCommand{
		EventType: domain.EventTypeCustomerCreated,
		Payload:   []byte(`{"event_type":"customer.created","uo_id":10,"occurred_at":"2026-07-22T15:22:50+03:00"}`),
	})
	if !errors.Is(err, ErrInvalidEventPayload) {
		t.Fatalf("expected ErrInvalidEventPayload, got %v", err)
	}
}

func TestConsumeRejectsMissingOccurredAt(t *testing.T) {
	service := NewService(&fakeConsumeRepository{})

	_, err := service.Consume(context.Background(), domain.ConsumeCommand{
		EventID:   "event-1",
		EventType: domain.EventTypeCustomerCreated,
		Payload:   []byte(`{"event_id":"event-1","event_type":"customer.created","uo_id":10}`),
	})
	if !errors.Is(err, ErrInvalidEventPayload) {
		t.Fatalf("expected ErrInvalidEventPayload, got %v", err)
	}
}

func TestConsumeDelegatesCustomerCreated(t *testing.T) {
	repository := &fakeConsumeRepository{
		createdResult: domain.ConsumeResult{
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
			"telefon":"05550000000",
			"occurred_at":"2026-07-22T15:22:50+03:00"
		}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CustomerID != 99 {
		t.Fatalf("expected customer id 99, got %d", result.CustomerID)
	}

	if repository.createdEvent.UOId != 10 {
		t.Fatalf("expected repository to receive uo_id 10, got %d", repository.createdEvent.UOId)
	}

	if repository.createdEvent.EventID != "event-1" {
		t.Fatalf("expected repository to receive event_id event-1, got %q", repository.createdEvent.EventID)
	}
}

func TestConsumeDelegatesCustomerUpdated(t *testing.T) {
	repository := &fakeConsumeRepository{
		updatedResult: domain.ConsumeResult{
			EventID:    "event-2",
			CustomerID: 77,
			Action:     "updated",
		},
	}
	service := NewService(repository)

	result, err := service.Consume(context.Background(), domain.ConsumeCommand{
		EventID:   "event-2",
		EventType: domain.EventTypeCustomerUpdated,
		Payload: []byte(`{
			"event_id":"event-2",
			"event_type":"customer.updated",
			"uo_id":10,
			"occurred_at":"2026-07-22T16:22:50+03:00"
		}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Action != "updated" {
		t.Fatalf("expected updated action, got %q", result.Action)
	}

	if repository.updatedEvent.UOId != 10 {
		t.Fatalf("expected repository to receive uo_id 10, got %d", repository.updatedEvent.UOId)
	}
}

func TestConsumeDelegatesCustomerDeleted(t *testing.T) {
	repository := &fakeConsumeRepository{
		deletedResult: domain.ConsumeResult{
			EventID:    "event-3",
			CustomerID: 55,
			Action:     "deleted",
		},
	}
	service := NewService(repository)

	result, err := service.Consume(context.Background(), domain.ConsumeCommand{
		EventID:   "event-3",
		EventType: domain.EventTypeCustomerDeleted,
		Payload: []byte(`{
			"event_id":"event-3",
			"event_type":"customer.deleted",
			"uo_id":10,
			"occurred_at":"2026-07-22T17:22:50+03:00"
		}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Action != "deleted" {
		t.Fatalf("expected deleted action, got %q", result.Action)
	}

	if repository.deletedEvent.UOId != 10 {
		t.Fatalf("expected repository to receive uo_id 10, got %d", repository.deletedEvent.UOId)
	}

	if repository.deletedEvent.EventID != "event-3" {
		t.Fatalf("expected repository to receive event_id event-3, got %q", repository.deletedEvent.EventID)
	}
}
