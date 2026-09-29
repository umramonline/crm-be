package application

import (
	"context"
	"strings"

	"github.com/umran/new.crm/backend/internal/customer/domain"
)

func filterCustomerByBackendTextFilters(customer domain.Customer, query domain.ListQuery) bool {
	if value := strings.TrimSpace(query.Unvan); value != "" {
		pattern := strings.ToLower(value)
		unvan := strings.ToLower(strings.TrimSpace(customer.Unvan))
		ad := strings.ToLower(strings.TrimSpace(customer.Ad))
		soyad := strings.ToLower(strings.TrimSpace(customer.Soyad))
		fullName := strings.TrimSpace(ad + " " + soyad)
		if !strings.Contains(unvan, pattern) &&
			!strings.Contains(ad, pattern) &&
			!strings.Contains(soyad, pattern) &&
			!strings.Contains(fullName, pattern) {
			return false
		}
	}

	if value := strings.TrimSpace(query.Ad); value != "" &&
		!strings.Contains(strings.ToLower(customer.Ad), strings.ToLower(value)) {
		return false
	}

	if value := strings.TrimSpace(query.Soyad); value != "" &&
		!strings.Contains(strings.ToLower(customer.Soyad), strings.ToLower(value)) {
		return false
	}

	if value := strings.TrimSpace(query.Type); value != "" &&
		strings.ToLower(strings.TrimSpace(customer.Type)) != strings.ToLower(value) {
		return false
	}

	if value := strings.TrimSpace(query.CreatedAt); value != "" {
		createdAt := ""
		if customer.CreatedAt != nil {
			createdAt = strings.TrimSpace(*customer.CreatedAt)
		}
		if !strings.Contains(createdAt, value) {
			return false
		}
	}

	return true
}

func customerMatchesPhoneListFilters(customer domain.Customer, query domain.ListQuery) bool {
	if phone := strings.TrimSpace(query.Cep); phone != "" &&
		!domain.CustomerPhoneMatchesFilter(customer, phone) {
		return false
	}

	if !filterCustomerByBackendTextFilters(customer, query) {
		return false
	}

	return filterCustomerByUmramonlineFields(customer, query)
}

func uoListQueryWithoutPhoneFilter(query domain.ListQuery) domain.ListQuery {
	scan := query
	scan.Cep = ""
	return scan
}

func (s *Service) listMergedCustomersUOPhoneScan(
	ctx context.Context,
	query domain.ListQuery,
) (domain.ListResult, error) {
	phone := strings.TrimSpace(query.Cep)
	if phone == "" {
		return emptyListResult(query), nil
	}

	matched := make([]domain.Customer, 0)
	uoPage := 1
	lastPage := 1
	scanPerPage := 100
	scanQuery := uoListQueryWithoutPhoneFilter(query)

	for uoPage <= lastPage && uoPage <= maxUOIntersectScanPages {
		scanQuery.Page = uoPage
		scanQuery.PerPage = scanPerPage
		scanQuery.IDs = nil

		uoBatch, err := s.provider.ListCustomers(ctx, scanQuery)
		if err != nil {
			return domain.ListResult{}, ErrCustomerListUnavailable
		}

		if len(uoBatch.Items) > 0 {
			backendByUOID, err := s.backendCustomersByUOID(ctx, extractUOIds(uoBatch.Items))
			if err != nil {
				return domain.ListResult{}, ErrCustomerListUnavailable
			}

			for _, uoItem := range uoBatch.Items {
				merged := mergeCustomer(backendByUOID[uoItem.UOId], uoItem)
				if !customerMatchesPhoneListFilters(merged, query) {
					continue
				}
				matched = append(matched, merged)
			}
		}

		if uoBatch.Pagination.LastPage <= 0 {
			break
		}

		lastPage = uoBatch.Pagination.LastPage
		uoPage++
	}

	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortMergedCustomers(matched, sortBy, query.SortOrder)

	return domain.PaginateCustomers(matched, query.Page, query.PerPage), nil
}

func (s *Service) listMergedCustomersWithPhoneFilter(
	ctx context.Context,
	query domain.ListQuery,
) (domain.ListResult, error) {
	backendResult, err := s.repository.ListCustomers(ctx, query)
	if err != nil {
		return domain.ListResult{}, ErrCustomerListUnavailable
	}
	if backendResult.Pagination.Total > 0 {
		return s.mergeBackendListPageWithUmramonline(ctx, backendResult)
	}

	return s.listMergedCustomersUOPhoneScan(ctx, query)
}
