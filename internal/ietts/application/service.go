package application

import (
	"context"
	"errors"
	"strings"

	"github.com/umran/new.crm/backend/internal/ietts/domain"
	"gorm.io/gorm"
)

var (
	ErrIettsListUnavailable      = errors.New("ietts list unavailable")
	ErrIettsConvertUnavailable   = errors.New("ietts convert unavailable")
	ErrIettsRecordNotFound       = errors.New("ietts record not found")
	ErrIettsInvalidConvertInput  = errors.New("ietts invalid convert input")
)

const customerAddressDetailMaxLength = 255

type Repository interface {
	ListRecords(ctx context.Context, query domain.ListQuery) (domain.ListResult, error)
	FindRecordByUUID(ctx context.Context, uuid string) (domain.Record, error)
}

type CustomerFromIettsWriter interface {
	CreateCustomerFromIetts(ctx context.Context, input domain.CustomerFromIettsInput) (uint64, error)
}

type Service struct {
	repository     Repository
	customerWriter CustomerFromIettsWriter
}

func NewService(repository Repository, customerWriter CustomerFromIettsWriter) *Service {
	return &Service{
		repository:     repository,
		customerWriter: customerWriter,
	}
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

func (s *Service) ConvertToCustomer(ctx context.Context, uuid string) (domain.ConvertToCustomerResult, error) {
	normalizedUUID := strings.TrimSpace(uuid)
	if s == nil || s.repository == nil || s.customerWriter == nil || normalizedUUID == "" {
		return domain.ConvertToCustomerResult{}, ErrIettsInvalidConvertInput
	}

	record, err := s.repository.FindRecordByUUID(ctx, normalizedUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ConvertToCustomerResult{}, ErrIettsRecordNotFound
		}

		return domain.ConvertToCustomerResult{}, ErrIettsConvertUnavailable
	}

	ad, soyad := domain.SplitBusinessName(record.BusinessName)
	customerID, err := s.customerWriter.CreateCustomerFromIetts(ctx, domain.CustomerFromIettsInput{
		Unvan:         record.CompanyName,
		Ad:            ad,
		Soyad:         soyad,
		AddressDetail: domain.TruncateRunes(record.BusinessAddress, customerAddressDetailMaxLength),
	})
	if err != nil {
		return domain.ConvertToCustomerResult{}, ErrIettsConvertUnavailable
	}

	return domain.ConvertToCustomerResult{CustomerID: customerID}, nil
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
