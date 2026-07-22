package application

import (
	"context"
	"errors"
	"strings"

	"github.com/umran/new.crm/backend/internal/consume/domain"
)

var (
	ErrInvalidEventPayload   = errors.New("invalid event payload")
	ErrUnsupportedEventType  = errors.New("unsupported event type")
)

type Repository interface {
	ConsumeCustomerCreated(ctx context.Context, event domain.CustomerCreatedEvent) (domain.ConsumeResult, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Consume(ctx context.Context, event domain.CustomerCreatedEvent) (domain.ConsumeResult, error) {
	if s == nil || s.repository == nil {
		return domain.ConsumeResult{}, ErrInvalidEventPayload
	}

	event.EventID = strings.TrimSpace(event.EventID)
	event.EventType = strings.TrimSpace(event.EventType)

	if event.EventID == "" {
		return domain.ConsumeResult{}, ErrInvalidEventPayload
	}

	if event.EventType != domain.EventTypeCustomerCreated {
		return domain.ConsumeResult{}, ErrUnsupportedEventType
	}

	if event.UOId == 0 {
		return domain.ConsumeResult{}, ErrInvalidEventPayload
	}

	return s.repository.ConsumeCustomerCreated(ctx, event)
}
