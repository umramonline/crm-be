package application

import (
	"context"
	"testing"

	"github.com/umran/new.crm/backend/internal/customer/domain"
)

type fakeCustomerProvider struct {
	listZonesBranchIDs    []uint64
	listBranchesBranchIDs []uint64
}

func (f *fakeCustomerProvider) ListCustomers(_ context.Context, _ domain.ListQuery) (domain.ListResult, error) {
	return domain.ListResult{}, nil
}

func (f *fakeCustomerProvider) ListZones(_ context.Context, branchIDs []uint64) ([]domain.Zone, error) {
	f.listZonesBranchIDs = branchIDs

	return []domain.Zone{{ID: 1, Name: "Zone 1"}}, nil
}

func (f *fakeCustomerProvider) SearchCustomer(_ context.Context, _ string) (domain.CustomerDetail, bool, error) {
	return domain.CustomerDetail{}, false, nil
}

func (f *fakeCustomerProvider) GetCustomer(_ context.Context, _ uint64) (domain.CustomerDetail, error) {
	return domain.CustomerDetail{}, nil
}

func (f *fakeCustomerProvider) PhoneExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (f *fakeCustomerProvider) ListCities(_ context.Context) ([]domain.City, error) {
	return nil, nil
}

func (f *fakeCustomerProvider) ListTowns(_ context.Context, _ uint64) ([]domain.Town, error) {
	return nil, nil
}

func (f *fakeCustomerProvider) ListBranches(_ context.Context, branchIDs []uint64) ([]domain.Branch, error) {
	f.listBranchesBranchIDs = branchIDs

	return []domain.Branch{{ID: 1, Name: "Branch 1"}}, nil
}

func (f *fakeCustomerProvider) ListBranchUsers(_ context.Context, _ uint64) ([]domain.BranchUser, error) {
	return nil, nil
}

func TestListBranchesIncludeAllFetchesAllBranches(t *testing.T) {
	provider := &fakeCustomerProvider{}
	service := NewService(provider, nil)

	_, err := service.ListBranches(context.Background(), []uint64{3, 5}, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if provider.listBranchesBranchIDs != nil {
		t.Fatalf("expected nil branch IDs for includeAll, got %#v", provider.listBranchesBranchIDs)
	}
}

func TestListBranchesScopedToUserBranchIDs(t *testing.T) {
	provider := &fakeCustomerProvider{}
	service := NewService(provider, nil)

	_, err := service.ListBranches(context.Background(), []uint64{3, 5}, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(provider.listBranchesBranchIDs) != 2 || provider.listBranchesBranchIDs[0] != 3 || provider.listBranchesBranchIDs[1] != 5 {
		t.Fatalf("expected scoped branch IDs, got %#v", provider.listBranchesBranchIDs)
	}
}

func TestListBranchesReturnsEmptyWithoutAccess(t *testing.T) {
	provider := &fakeCustomerProvider{}
	service := NewService(provider, nil)

	branches, err := service.ListBranches(context.Background(), nil, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(branches) != 0 {
		t.Fatalf("expected empty branches, got %#v", branches)
	}

	if provider.listBranchesBranchIDs != nil {
		t.Fatalf("expected provider not to be called, got %#v", provider.listBranchesBranchIDs)
	}
}

func TestListZonesIncludeAllFetchesAllZones(t *testing.T) {
	provider := &fakeCustomerProvider{}
	service := NewService(provider, nil)

	_, err := service.ListZones(context.Background(), []uint64{3, 5}, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if provider.listZonesBranchIDs != nil {
		t.Fatalf("expected nil branch IDs for includeAll, got %#v", provider.listZonesBranchIDs)
	}
}

func TestListZonesScopedToUserBranchIDs(t *testing.T) {
	provider := &fakeCustomerProvider{}
	service := NewService(provider, nil)

	_, err := service.ListZones(context.Background(), []uint64{7}, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(provider.listZonesBranchIDs) != 1 || provider.listZonesBranchIDs[0] != 7 {
		t.Fatalf("expected scoped branch IDs, got %#v", provider.listZonesBranchIDs)
	}
}
