package application

import (
	"context"
	"testing"

	"github.com/umran/new.crm/backend/internal/customer/domain"
)

type fakeCustomerRepository struct {
	uoIDs []uint64
	items []domain.Customer
}

func (f *fakeCustomerRepository) ListCustomers(_ context.Context, query domain.ListQuery) (domain.ListResult, error) {
	return domain.PaginateCustomers(f.items, query.Page, query.PerPage), nil
}

func (f *fakeCustomerRepository) ListCustomerUOIds(_ context.Context, _ domain.ListQuery) ([]uint64, error) {
	return f.uoIDs, nil
}

func (f *fakeCustomerRepository) ListCustomersByUOIds(_ context.Context, uoIDs []uint64) ([]domain.Customer, error) {
	byID := map[uint64]domain.Customer{}
	for _, item := range f.items {
		if item.UOId > 0 {
			byID[item.UOId] = item
		}
	}

	result := make([]domain.Customer, 0, len(uoIDs))
	for _, id := range uoIDs {
		if item, ok := byID[id]; ok {
			result = append(result, item)
		}
	}

	return result, nil
}

func (f *fakeCustomerRepository) SearchCustomer(_ context.Context, _ string) (domain.CustomerDetail, bool, error) {
	return domain.CustomerDetail{}, false, nil
}

func (f *fakeCustomerRepository) GetCustomer(_ context.Context, _ uint64) (domain.CustomerDetail, error) {
	return domain.CustomerDetail{}, nil
}

func (f *fakeCustomerRepository) PhoneExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (f *fakeCustomerRepository) PhoneExistsExcept(_ context.Context, _ string, _ uint64) (bool, error) {
	return false, nil
}

func (f *fakeCustomerRepository) CreateCustomer(_ context.Context, _ domain.CreateCustomerInput) (domain.CustomerDetail, error) {
	return domain.CustomerDetail{}, nil
}

func (f *fakeCustomerRepository) GetFullRegistrationCustomer(_ context.Context, _ uint64) (domain.CustomerDetail, error) {
	return domain.CustomerDetail{}, nil
}

func (f *fakeCustomerRepository) CompleteFullRegistration(_ context.Context, _ uint64, _ domain.FullRegistrationInput) (domain.CustomerDetail, error) {
	return domain.CustomerDetail{}, nil
}

func (f *fakeCustomerRepository) UpdateSourceEditableFullRegistration(_ context.Context, _ uint64, _ domain.FullRegistrationInput) (domain.CustomerDetail, error) {
	return domain.CustomerDetail{}, nil
}

type recordingCustomerProvider struct {
	lastQuery domain.ListQuery
	items     []domain.Customer
}

func (p *recordingCustomerProvider) ListCustomers(_ context.Context, query domain.ListQuery) (domain.ListResult, error) {
	p.lastQuery = query

	pageItems := p.items
	if len(query.IDs) > 0 {
		allowed := map[uint64]struct{}{}
		for _, id := range query.IDs {
			allowed[id] = struct{}{}
		}

		filtered := make([]domain.Customer, 0, len(p.items))
		for _, item := range p.items {
			if _, ok := allowed[item.UOId]; ok {
				filtered = append(filtered, item)
			}
		}
		pageItems = filtered
	}

	return domain.PaginateCustomers(pageItems, query.Page, query.PerPage), nil
}

func (p *recordingCustomerProvider) ListZones(_ context.Context, _ []uint64) ([]domain.Zone, error) {
	return nil, nil
}

func (p *recordingCustomerProvider) SearchCustomer(_ context.Context, _ string) (domain.CustomerDetail, bool, error) {
	return domain.CustomerDetail{}, false, nil
}

func (p *recordingCustomerProvider) GetCustomer(_ context.Context, _ uint64) (domain.CustomerDetail, error) {
	return domain.CustomerDetail{}, nil
}

func (p *recordingCustomerProvider) PhoneExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (p *recordingCustomerProvider) ListCities(_ context.Context) ([]domain.City, error) {
	return nil, nil
}

func (p *recordingCustomerProvider) ListTowns(_ context.Context, _ uint64) ([]domain.Town, error) {
	return nil, nil
}

func (p *recordingCustomerProvider) ListBranches(_ context.Context, _ []uint64) ([]domain.Branch, error) {
	return nil, nil
}

func (p *recordingCustomerProvider) ListBranchUsers(_ context.Context, _ uint64) ([]domain.BranchUser, error) {
	return nil, nil
}

func TestListMergedCustomersBackendPaginatedUsesRepositoryPagination(t *testing.T) {
	repository := &fakeCustomerRepository{
		items: []domain.Customer{
			{ID: 1, UOId: 10, Unvan: "Alpha"},
			{ID: 2, UOId: 20, Unvan: "Beta"},
		},
	}
	provider := &recordingCustomerProvider{
		items: []domain.Customer{
			{UOId: 10, BranchName: "Bayi A"},
			{UOId: 20, BranchName: "Bayi B"},
		},
	}

	service := NewService(provider, repository)
	result, err := service.ListCustomers(context.Background(), domain.ListQuery{
		Page:    1,
		PerPage: 1,
		Unvan:   "a",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Items) != 1 || result.Pagination.Total != 2 {
		t.Fatalf("unexpected pagination result: %#v", result)
	}

	if result.Items[0].BranchName != "Bayi A" {
		t.Fatalf("expected merged branch name, got %q", result.Items[0].BranchName)
	}
}

func TestListMergedCustomersUmramonlineLeadPassesFilters(t *testing.T) {
	repository := &fakeCustomerRepository{}
	provider := &recordingCustomerProvider{
		items: []domain.Customer{
			{UOId: 10, Situation: "Aktif Müşteri", BranchName: "Bayi A"},
		},
	}

	service := NewService(provider, repository)
	_, err := service.ListCustomers(context.Background(), domain.ListQuery{
		Page:       1,
		PerPage:    20,
		Situation:  "Aktif Müşteri",
		BranchName: "Bayi A",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if provider.lastQuery.Situation != "Aktif Müşteri" || provider.lastQuery.BranchName != "Bayi A" {
		t.Fatalf("expected UO filters to be forwarded, got %#v", provider.lastQuery)
	}
}
