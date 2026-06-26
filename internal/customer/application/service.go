package application

import (
	"context"
	"errors"

	"github.com/umran/new.crm/backend/internal/customer/domain"
)

var ErrCustomerListUnavailable = errors.New("customer list unavailable")

var ErrZoneListUnavailable = errors.New("zone list unavailable")

type CustomerProvider interface {
	ListCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error)
	ListZones(ctx context.Context) ([]domain.Zone, error)
}

type Service struct {
	provider CustomerProvider
}

func NewService(provider CustomerProvider) *Service {
	return &Service{provider: provider}
}

func (s *Service) ListCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	if s == nil || s.provider == nil {
		return domain.ListResult{}, ErrCustomerListUnavailable
	}

	return s.provider.ListCustomers(ctx, query)
}

func (s *Service) ListZones(ctx context.Context) ([]domain.Zone, error) {
	if s == nil || s.provider == nil {
		return nil, ErrZoneListUnavailable
	}

	return s.provider.ListZones(ctx)
}
