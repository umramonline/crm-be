package domain

import "strings"

var allowedFollowUpListSortColumns = map[string]struct{}{
	"title":                 {},
	"customer":              {},
	"assigned_user_full_name": {},
	"branch_name":           {},
	"visit_date":            {},
	"next_visit_date":       {},
	"agreement_reached":     {},
}

func NormalizeListSortBy(sortBy string) string {
	normalized := strings.ToLower(strings.TrimSpace(sortBy))
	if _, ok := allowedFollowUpListSortColumns[normalized]; ok {
		return normalized
	}

	return ""
}
