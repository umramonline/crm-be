package application

import (
	"context"
	"errors"
	"strings"

	"github.com/umran/new.crm/backend/internal/consume/domain"
)

var (
	ErrInvalidEventPayload  = errors.New("invalid event payload")
	ErrUnsupportedEventType = errors.New("unsupported event type")
	ErrCustomerNotFound     = errors.New("customer not found")
)

type eventHandler func(ctx context.Context, command domain.ConsumeCommand) (domain.ConsumeResult, error)

type Repository interface {
	ConsumeCustomerCreated(ctx context.Context, event domain.CustomerCreatedEvent) (domain.ConsumeResult, error)
	ConsumeCustomerUpdated(ctx context.Context, event domain.CustomerUpdatedEvent) (domain.ConsumeResult, error)
	ConsumeCustomerDeleted(ctx context.Context, event domain.CustomerDeletedEvent) (domain.ConsumeResult, error)
}

type Service struct {
	repository Repository
	handlers   map[string]eventHandler
}

func NewService(repository Repository) *Service {
	service := &Service{
		repository: repository,
		handlers:   make(map[string]eventHandler),
	}

	service.registerHandlers()

	return service
}

func (s *Service) registerHandlers() {
	s.handlers[domain.EventTypeCustomerCreated] = s.handleCustomerCreated
	s.handlers[domain.EventTypeCustomerUpdated] = s.handleCustomerUpdated
	s.handlers[domain.EventTypeCustomerDeleted] = s.handleCustomerDeleted
}

func (s *Service) Consume(ctx context.Context, command domain.ConsumeCommand) (domain.ConsumeResult, error) {
	if s == nil || s.repository == nil {
		return domain.ConsumeResult{}, ErrInvalidEventPayload
	}

	command.EventID = strings.TrimSpace(command.EventID)
	command.EventType = strings.TrimSpace(command.EventType)

	if command.EventID == "" || command.EventType == "" {
		return domain.ConsumeResult{}, ErrInvalidEventPayload
	}

	if len(command.Payload) == 0 {
		return domain.ConsumeResult{}, ErrInvalidEventPayload
	}

	handler, ok := s.handlers[command.EventType]
	if !ok {
		return domain.ConsumeResult{}, ErrUnsupportedEventType
	}

	return handler(ctx, command)
}
