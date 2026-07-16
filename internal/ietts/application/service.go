package application

import (
	"context"
	"errors"
	"strings"

	"github.com/umran/new.crm/backend/internal/ietts/domain"
)

var ErrIettsListUnavailable = errors.New("ietts list unavailable")

type Repository interface {
	ListRecords(ctx context.Context, query domain.ListQuery) (domain.ListResult, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListRecords(ctx context.Context, query domain.ListQuery) (domain.ListResult, error) {
	if s == nil || s.repository == nil {
		return domain.ListResult{}, ErrIettsListUnavailable
	}

	result, err := s.repository.ListRecords(ctx, normalizeListQuery(query))
	if err != nil {
		return domain.ListResult{}, ErrIettsListUnavailable
	}

	return result, nil
}

func normalizeListQuery(query domain.ListQuery) domain.ListQuery {
	sortBy := strings.ToLower(strings.TrimSpace(query.SortBy))
	switch sortBy {
	case "document_issue_date", "created_at":
	default:
		sortBy = ""
	}

	sortOrder := strings.ToLower(strings.TrimSpace(query.SortOrder))
	if sortOrder != "asc" {
		sortOrder = "desc"
	}

	return domain.ListQuery{
		Page:              query.Page,
		PerPage:           query.PerPage,
		DocumentNumber:    strings.TrimSpace(query.DocumentNumber),
		CompanyName:       strings.TrimSpace(query.CompanyName),
		BusinessName:      strings.TrimSpace(query.BusinessName),
		BusinessAddress:   strings.TrimSpace(query.BusinessAddress),
		DocumentIssueDate: strings.TrimSpace(query.DocumentIssueDate),
		DocumentStatus:    strings.TrimSpace(query.DocumentStatus),
		City:              strings.TrimSpace(query.City),
		District:          strings.TrimSpace(query.District),
		CreatedAt:         strings.TrimSpace(query.CreatedAt),
		SortBy:            sortBy,
		SortOrder:         sortOrder,
	}
}
