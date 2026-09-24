package domain

import "strings"

var allowedTaskListSortColumns = map[string]struct{}{
	"title":                     {},
	"customer_count":            {},
	"assigned_user_full_name":   {},
	"branch_name":               {},
	"visit_date":                {},
	"due_date":                  {},
	"priority":                  {},
	"created_by_user_full_name": {},
}

func NormalizeListSortBy(sortBy string) string {
	normalized := strings.ToLower(strings.TrimSpace(sortBy))
	if _, ok := allowedTaskListSortColumns[normalized]; ok {
		return normalized
	}

	return ""
}
