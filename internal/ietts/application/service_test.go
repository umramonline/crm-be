package application

import (
	"context"
	"errors"
	"testing"

	"github.com/umran/new.crm/backend/internal/ietts/domain"
	"gorm.io/gorm"
)

type fakeRepository struct {
	listQuery   domain.ListQuery
	record      domain.Record
	findErr     error
}

func (f *fakeRepository) ListRecords(_ context.Context, query domain.ListQuery) (domain.ListResult, error) {
	f.listQuery = query

	return domain.ListResult{}, nil
}

func (f *fakeRepository) FindRecordByUUID(_ context.Context, _ string) (domain.Record, error) {
	if f.findErr != nil {
		return domain.Record{}, f.findErr
	}

	return f.record, nil
}

type fakeCustomerWriter struct {
	input domain.CustomerFromIettsInput
	id    uint64
	err   error
}

func (f *fakeCustomerWriter) CreateCustomerFromIetts(_ context.Context, input domain.CustomerFromIettsInput) (uint64, error) {
	f.input = input
	if f.err != nil {
		return 0, f.err
	}

	return f.id, nil
}

func TestListRecordsRejectsInvalidSortBy(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, &fakeCustomerWriter{})

	_, err := service.ListRecords(context.Background(), domain.ListQuery{
		SortBy:    "company_name",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.listQuery.SortBy != "" {
		t.Fatalf("expected empty sort_by, got %q", repository.listQuery.SortBy)
	}
}

func TestListRecordsAcceptsDocumentIssueDateSort(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, &fakeCustomerWriter{})

	_, err := service.ListRecords(context.Background(), domain.ListQuery{
		SortBy:    "document_issue_date",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.listQuery.SortBy != "document_issue_date" {
		t.Fatalf("expected document_issue_date sort_by, got %q", repository.listQuery.SortBy)
	}
	if repository.listQuery.SortOrder != "asc" {
		t.Fatalf("expected asc sort_order, got %q", repository.listQuery.SortOrder)
	}
}

func TestListRecordsAcceptsCreatedAtSort(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, &fakeCustomerWriter{})

	_, err := service.ListRecords(context.Background(), domain.ListQuery{
		SortBy:    "created_at",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.listQuery.SortBy != "created_at" {
		t.Fatalf("expected created_at sort_by, got %q", repository.listQuery.SortBy)
	}
}

func TestListRecordsDefaultsSortOrderToDesc(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, &fakeCustomerWriter{})

	_, err := service.ListRecords(context.Background(), domain.ListQuery{
		SortBy: "document_issue_date",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.listQuery.SortOrder != "desc" {
		t.Fatalf("expected desc sort_order, got %q", repository.listQuery.SortOrder)
	}
}

func TestListRecordsTrimsFilters(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, &fakeCustomerWriter{})

	_, err := service.ListRecords(context.Background(), domain.ListQuery{
		DocumentNumber: "  ABC  ",
		City:           "  Istanbul ",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.listQuery.DocumentNumber != "ABC" {
		t.Fatalf("expected trimmed document_number, got %q", repository.listQuery.DocumentNumber)
	}
	if repository.listQuery.City != "Istanbul" {
		t.Fatalf("expected trimmed city, got %q", repository.listQuery.City)
	}
}

func TestListRecordsUnavailableWithoutRepository(t *testing.T) {
	service := NewService(nil, &fakeCustomerWriter{})

	_, err := service.ListRecords(context.Background(), domain.ListQuery{})
	if !errors.Is(err, ErrIettsListUnavailable) {
		t.Fatalf("expected ErrIettsListUnavailable, got %v", err)
	}
}

func TestConvertToCustomerMapsFields(t *testing.T) {
	repository := &fakeRepository{
		record: domain.Record{
			UUID:            "uuid-1",
			CompanyName:     "ACME LTD",
			BusinessName:    "AHMET MEHMET YILMAZ",
			BusinessAddress: "Example Street 1",
		},
	}
	customerWriter := &fakeCustomerWriter{id: 86}
	service := NewService(repository, customerWriter)

	result, err := service.ConvertToCustomer(context.Background(), "uuid-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.CustomerID != 86 {
		t.Fatalf("expected customer id 86, got %d", result.CustomerID)
	}
	if customerWriter.input.Unvan != "ACME LTD" {
		t.Fatalf("expected unvan ACME LTD, got %q", customerWriter.input.Unvan)
	}
	if customerWriter.input.Ad != "AHMET" {
		t.Fatalf("expected ad AHMET, got %q", customerWriter.input.Ad)
	}
	if customerWriter.input.Soyad != "MEHMET YILMAZ" {
		t.Fatalf("expected soyad MEHMET YILMAZ, got %q", customerWriter.input.Soyad)
	}
	if customerWriter.input.AddressDetail != "Example Street 1" {
		t.Fatalf("expected address detail, got %q", customerWriter.input.AddressDetail)
	}
}

func TestConvertToCustomerReturnsNotFound(t *testing.T) {
	repository := &fakeRepository{findErr: gorm.ErrRecordNotFound}
	service := NewService(repository, &fakeCustomerWriter{})

	_, err := service.ConvertToCustomer(context.Background(), "missing-uuid")
	if !errors.Is(err, ErrIettsRecordNotFound) {
		t.Fatalf("expected ErrIettsRecordNotFound, got %v", err)
	}
}

func TestConvertToCustomerRejectsEmptyUUID(t *testing.T) {
	service := NewService(&fakeRepository{}, &fakeCustomerWriter{})

	_, err := service.ConvertToCustomer(context.Background(), " ")
	if !errors.Is(err, ErrIettsInvalidConvertInput) {
		t.Fatalf("expected ErrIettsInvalidConvertInput, got %v", err)
	}
}
