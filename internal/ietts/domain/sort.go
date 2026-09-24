package domain

import "strings"

// Allowed list sort columns (ietts_records table).
var allowedListSortColumns = map[string]struct{}{
	"document_number":     {},
	"company_name":        {},
	"business_name":       {},
	"business_address":    {},
	"document_issue_date": {},
	"document_status":     {},
	"city":                {},
	"district":            {},
	"created_at":          {},
}

func NormalizeListSortBy(sortBy string) string {
	normalized := strings.ToLower(strings.TrimSpace(sortBy))
	if _, ok := allowedListSortColumns[normalized]; ok {
		return normalized
	}

	return ""
}
