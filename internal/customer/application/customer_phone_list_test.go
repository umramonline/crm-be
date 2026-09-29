package application

import (
	"context"
	"testing"

	"github.com/umran/new.crm/backend/internal/customer/domain"
)

func TestListMergedCustomersUOPhoneScanMatchesPartialPhone(t *testing.T) {
	provider := &recordingCustomerProvider{
		items: []domain.Customer{
			{UOId: 10, Cep: "05555512345", Unvan: "Test Galeri"},
			{UOId: 11, Cep: "05321234567", Unvan: "Other"},
		},
	}

	service := &Service{
		repository: &fakeCustomerRepository{},
		provider:   provider,
	}

	result, err := service.listMergedCustomersUOPhoneScan(context.Background(), domain.ListQuery{
		Page:    1,
		PerPage: 20,
		Cep:     "055555",
	})
	if err != nil {
		t.Fatalf("listMergedCustomersUOPhoneScan: %v", err)
	}
	if result.Pagination.Total != 1 {
		t.Fatalf("expected 1 match, got total=%d items=%#v", result.Pagination.Total, result.Items)
	}
	if result.Items[0].UOId != 10 {
		t.Fatalf("expected UO id 10, got %#v", result.Items[0])
	}
	if provider.lastQuery.Cep != "" {
		t.Fatalf("expected UO scan without cep param, got cep=%q", provider.lastQuery.Cep)
	}
}
