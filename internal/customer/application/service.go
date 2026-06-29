package application

import (
	"context"
	"errors"
	"strings"

	"github.com/umran/new.crm/backend/internal/customer/domain"
)

var ErrCustomerListUnavailable = errors.New("customer list unavailable")

var ErrZoneListUnavailable = errors.New("zone list unavailable")

var ErrCustomerSearchUnavailable = errors.New("customer search unavailable")

var ErrReferenceDataUnavailable = errors.New("reference data unavailable")

var ErrInvalidCustomerSearchQuery = errors.New("customer search query is required")

type CustomerProvider interface {
	ListCustomers(ctx context.Context, query domain.ListQuery) (domain.ListResult, error)
	ListZones(ctx context.Context) ([]domain.Zone, error)
	SearchCustomer(ctx context.Context, query string) (domain.CustomerDetail, bool, error)
	ListCities(ctx context.Context) ([]domain.City, error)
	ListTowns(ctx context.Context, cityID uint64) ([]domain.Town, error)
	ListBranches(ctx context.Context) ([]domain.Branch, error)
}

type CustomerRepository interface {
	SearchCustomer(ctx context.Context, query string) (domain.CustomerDetail, bool, error)
}

type Service struct {
	provider   CustomerProvider
	repository CustomerRepository
}

func NewService(provider CustomerProvider, repositories ...CustomerRepository) *Service {
	var repository CustomerRepository
	if len(repositories) > 0 {
		repository = repositories[0]
	}

	return &Service{provider: provider, repository: repository}
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

func (s *Service) SearchCustomer(ctx context.Context, query string) (domain.CustomerSearchResult, error) {
	normalizedQuery := strings.TrimSpace(query)
	if normalizedQuery == "" {
		return domain.CustomerSearchResult{}, ErrInvalidCustomerSearchQuery
	}

	// ilk başta veritabanında arama yapıyoruz
	customer, found, err := s.repository.SearchCustomer(ctx, normalizedQuery)
	if err != nil {
		return domain.CustomerSearchResult{}, ErrCustomerSearchUnavailable
	}

	if found {
		return domain.CustomerSearchResult{
			Found:    true,
			Source:   "backend",
			Customer: &customer,
		}, nil
	}

	// eğer veritabanında bulunamadıysa, umramonline'dan arama yapıyoruz
	customer, found, err = s.provider.SearchCustomer(ctx, normalizedQuery)
	if err != nil {
		return domain.CustomerSearchResult{}, ErrCustomerSearchUnavailable
	}

	if !found {
		return domain.CustomerSearchResult{Found: false}, nil
	}

	return domain.CustomerSearchResult{
		Found:    true,
		Source:   "umramonline",
		Customer: &customer,
	}, nil
}

func (s *Service) ListCities(ctx context.Context) ([]domain.City, error) {
	if s == nil || s.provider == nil {
		return nil, ErrReferenceDataUnavailable
	}

	return s.provider.ListCities(ctx)
}

func (s *Service) ListTowns(ctx context.Context, cityID uint64) ([]domain.Town, error) {
	if cityID == 0 {
		return []domain.Town{}, nil
	}

	if s == nil || s.provider == nil {
		return nil, ErrReferenceDataUnavailable
	}

	return s.provider.ListTowns(ctx, cityID)
}

func (s *Service) ListBranches(ctx context.Context) ([]domain.Branch, error) {
	if s == nil || s.provider == nil {
		return nil, ErrReferenceDataUnavailable
	}

	return s.provider.ListBranches(ctx)
}
