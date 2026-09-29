package domain

import "strings"

var allowedCustomerListSortColumns = map[string]struct{}{
	"credit":              {},
	"point":               {},
	"created_at":          {},
	"vehicle_stock_count": {},
}

func NormalizeListSortBy(sortBy string) string {
	normalized := strings.ToLower(strings.TrimSpace(sortBy))
	if _, ok := allowedCustomerListSortColumns[normalized]; ok {
		return normalized
	}

	return ""
}
