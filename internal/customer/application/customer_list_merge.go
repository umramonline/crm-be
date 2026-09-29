package application

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/umran/new.crm/backend/internal/customer/domain"
)

const maxUOIntersectScanPages = 200

func firstNonEmptyCustomerPhone(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func hasUmramonlineFieldFilters(query domain.ListQuery) bool {
	return strings.TrimSpace(query.Situation) != "" ||
		strings.TrimSpace(query.BranchName) != "" ||
		strings.TrimSpace(query.ZoneName) != "" ||
		strings.TrimSpace(query.PlusCardNo) != "" ||
		strings.TrimSpace(query.City) != "" ||
		strings.TrimSpace(query.Town) != ""
}

func isCreditPointSort(sortBy string) bool {
	return sortBy == "credit" || sortBy == "point"
}

func filterCustomerByUmramonlineFields(customer domain.Customer, query domain.ListQuery) bool {
	if value := strings.TrimSpace(query.Situation); value != "" && customer.Situation != value {
		return false
	}
	if value := strings.TrimSpace(query.BranchName); value != "" && customer.BranchName != value {
		return false
	}
	if value := strings.TrimSpace(query.ZoneName); value != "" && customer.ZoneName != value {
		return false
	}
	if value := strings.TrimSpace(query.PlusCardNo); value != "" &&
		!strings.Contains(strings.ToLower(customer.PlusCardNo), strings.ToLower(value)) {
		return false
	}
	if value := strings.TrimSpace(query.City); value != "" && customer.City != value {
		return false
	}
	if value := strings.TrimSpace(query.Town); value != "" && customer.Town != value {
		return false
	}

	return true
}

func sortMergedCustomers(items []domain.Customer, sortBy string, sortOrder string) {
	if len(items) < 2 {
		return
	}

	ascending := strings.ToLower(strings.TrimSpace(sortOrder)) == "asc"

	sort.Slice(items, func(i, j int) bool {
		left := items[i]
		right := items[j]

		switch sortBy {
		case "credit":
			if left.Credit == right.Credit {
				return left.ID < right.ID
			}
			if ascending {
				return left.Credit < right.Credit
			}
			return left.Credit > right.Credit
		case "point":
			if left.Point == right.Point {
				return left.ID < right.ID
			}
			if ascending {
				return left.Point < right.Point
			}
			return left.Point > right.Point
		case "vehicle_stock_count":
			leftStock := int64(-1)
			rightStock := int64(-1)
			if left.VehicleStockCount != nil {
				leftStock = int64(*left.VehicleStockCount)
			}
			if right.VehicleStockCount != nil {
				rightStock = int64(*right.VehicleStockCount)
			}
			if leftStock == rightStock {
				return left.ID < right.ID
			}
			if ascending {
				return leftStock < rightStock
			}
			return leftStock > rightStock
		case "created_at":
			leftTime := customerCreatedAtUnix(left)
			rightTime := customerCreatedAtUnix(right)
			if leftTime == rightTime {
				return left.ID < right.ID
			}
			if ascending {
				return leftTime < rightTime
			}
			return leftTime > rightTime
		default:
			return left.ID > right.ID
		}
	})
}

func customerCreatedAtUnix(customer domain.Customer) int64 {
	if customer.CreatedAt == nil {
		return 0
	}

	value := strings.TrimSpace(*customer.CreatedAt)
	if value == "" {
		return 0
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Unix()
		}
	}

	return 0
}

func (s *Service) mergeBackendListPageWithUmramonline(
	ctx context.Context,
	backendResult domain.ListResult,
) (domain.ListResult, error) {
	uoIDs := make([]uint64, 0, len(backendResult.Items))
	for _, item := range backendResult.Items {
		if item.UOId > 0 {
			uoIDs = append(uoIDs, item.UOId)
		}
	}

	uoByID := map[uint64]domain.Customer{}
	if len(uoIDs) > 0 {
		uoResult, err := s.provider.ListCustomers(ctx, domain.ListQuery{
			Page:    1,
			PerPage: len(uoIDs),
			IDs:     uoIDs,
		})
		if err != nil {
			return domain.ListResult{}, ErrCustomerListUnavailable
		}
		for _, item := range uoResult.Items {
			uoByID[item.UOId] = item
		}
	}

	items := make([]domain.Customer, 0, len(backendResult.Items))
	for _, backendItem := range backendResult.Items {
		items = append(items, mergeCustomer(backendItem, uoByID[backendItem.UOId]))
	}

	backendResult.Items = items
	return backendResult, nil
}

func (s *Service) listMergedCustomersBackendPaginated(
	ctx context.Context,
	query domain.ListQuery,
) (domain.ListResult, error) {
	backendResult, err := s.repository.ListCustomers(ctx, query)
	if err != nil {
		return domain.ListResult{}, ErrCustomerListUnavailable
	}

	return s.mergeBackendListPageWithUmramonline(ctx, backendResult)
}

func (s *Service) listMergedCustomersUOIntersectScan(
	ctx context.Context,
	query domain.ListQuery,
	allowed map[uint64]struct{},
) (domain.ListResult, error) {
	matched := make([]domain.Customer, 0)
	uoPage := 1
	lastPage := 1
	scanPerPage := 100

	for uoPage <= lastPage && uoPage <= maxUOIntersectScanPages {
		scanQuery := query
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
				if _, ok := allowed[uoItem.UOId]; !ok {
					continue
				}

				merged := mergeCustomer(backendByUOID[uoItem.UOId], uoItem)
				if !filterCustomerByUmramonlineFields(merged, query) {
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

func (s *Service) listMergedCustomersFromAllowedUOIDs(
	ctx context.Context,
	query domain.ListQuery,
	allowed []uint64,
) (domain.ListResult, error) {
	if len(allowed) == 0 {
		return emptyListResult(query), nil
	}

	uoResult, err := s.provider.ListCustomers(ctx, domain.ListQuery{
		Page:       1,
		PerPage:    len(allowed),
		IDs:        allowed,
		SortBy:     query.SortBy,
		SortOrder:  query.SortOrder,
		Situation:  query.Situation,
		BranchName: query.BranchName,
		ZoneName:   query.ZoneName,
		PlusCardNo: query.PlusCardNo,
		City:       query.City,
		Town:       query.Town,
		BranchIDs:  query.BranchIDs,
	})
	if err != nil {
		return domain.ListResult{}, ErrCustomerListUnavailable
	}

	backendByUOID, err := s.backendCustomersByUOID(ctx, extractUOIds(uoResult.Items))
	if err != nil {
		return domain.ListResult{}, ErrCustomerListUnavailable
	}

	matched := make([]domain.Customer, 0, len(uoResult.Items))
	for _, uoItem := range uoResult.Items {
		merged := mergeCustomer(backendByUOID[uoItem.UOId], uoItem)
		if !filterCustomerByUmramonlineFields(merged, query) {
			continue
		}
		matched = append(matched, merged)
	}

	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortMergedCustomers(matched, sortBy, query.SortOrder)

	return domain.PaginateCustomers(matched, query.Page, query.PerPage), nil
}

func allowedUOSet(ids []uint64) map[uint64]struct{} {
	set := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		if id > 0 {
			set[id] = struct{}{}
		}
	}

	return set
}
