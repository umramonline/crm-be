package application

import (
	"context"
	"errors"
	"testing"

	"github.com/umran/new.crm/backend/internal/ietts/domain"
)

type fakeRepository struct {
	query domain.ListQuery
}

func (f *fakeRepository) ListRecords(_ context.Context, query domain.ListQuery) (domain.ListResult, error) {
	f.query = query

	return domain.ListResult{}, nil
}

func TestListRecordsRejectsInvalidSortBy(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ListRecords(context.Background(), domain.ListQuery{
		SortBy:    "company_name",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.query.SortBy != "" {
		t.Fatalf("expected empty sort_by, got %q", repository.query.SortBy)
	}
}

func TestListRecordsAcceptsDocumentIssueDateSort(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ListRecords(context.Background(), domain.ListQuery{
		SortBy:    "document_issue_date",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.query.SortBy != "document_issue_date" {
		t.Fatalf("expected document_issue_date sort_by, got %q", repository.query.SortBy)
	}
	if repository.query.SortOrder != "asc" {
		t.Fatalf("expected asc sort_order, got %q", repository.query.SortOrder)
	}
}

func TestListRecordsDefaultsSortOrderToDesc(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ListRecords(context.Background(), domain.ListQuery{
		SortBy: "document_issue_date",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.query.SortOrder != "desc" {
		t.Fatalf("expected desc sort_order, got %q", repository.query.SortOrder)
	}
}

func TestListRecordsTrimsFilters(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ListRecords(context.Background(), domain.ListQuery{
		DocumentNumber: "  ABC  ",
		City:           "  Istanbul ",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.query.DocumentNumber != "ABC" {
		t.Fatalf("expected trimmed document_number, got %q", repository.query.DocumentNumber)
	}
	if repository.query.City != "Istanbul" {
		t.Fatalf("expected trimmed city, got %q", repository.query.City)
	}
}

func TestListRecordsAcceptsCreatedAtSort(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ListRecords(context.Background(), domain.ListQuery{
		SortBy:    "created_at",
		SortOrder: "asc",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.query.SortBy != "created_at" {
		t.Fatalf("expected created_at sort_by, got %q", repository.query.SortBy)
	}
}

func TestListRecordsUnavailableWithoutRepository(t *testing.T) {
	service := NewService(nil)

	_, err := service.ListRecords(context.Background(), domain.ListQuery{})
	if !errors.Is(err, ErrIettsListUnavailable) {
		t.Fatalf("expected ErrIettsListUnavailable, got %v", err)
	}
}
