package domain

import "strings"

const (
	DefaultListPerPage = 10
	MaxListPerPage     = 100
)

func NormalizeListSortOrder(sortOrder string) string {
	if strings.ToLower(strings.TrimSpace(sortOrder)) == "asc" {
		return "asc"
	}

	return "desc"
}

func NormalizeListFilterValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	switch strings.ToLower(trimmed) {
	case "tümü", "tumu", "all":
		return ""
	default:
		return trimmed
	}
}

func NormalizeListQuery(query ListQuery) ListQuery {
	query.SortBy = NormalizeListSortBy(query.SortBy)
	query.SortOrder = NormalizeListSortOrder(query.SortOrder)

	query.Situation = NormalizeListFilterValue(query.Situation)
	query.Unvan = NormalizeListFilterValue(query.Unvan)
	query.Cep = strings.TrimSpace(query.Cep)
	query.Ad = NormalizeListFilterValue(query.Ad)
	query.Soyad = NormalizeListFilterValue(query.Soyad)
	query.BranchName = NormalizeListFilterValue(query.BranchName)
	query.ZoneName = NormalizeListFilterValue(query.ZoneName)
	query.PlusCardNo = NormalizeListFilterValue(query.PlusCardNo)
	query.Source = NormalizeListFilterValue(query.Source)
	query.City = NormalizeListFilterValue(query.City)
	query.Town = NormalizeListFilterValue(query.Town)
	query.CreatedAt = NormalizeListFilterValue(query.CreatedAt)
	query.Type = NormalizeListFilterValue(query.Type)

	if query.Page <= 0 {
		query.Page = 1
	}

	if query.PerPage <= 0 {
		query.PerPage = DefaultListPerPage
	}
	if query.PerPage > MaxListPerPage {
		query.PerPage = MaxListPerPage
	}

	return query
}

func PaginateCustomers(items []Customer, page int, perPage int) ListResult {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = DefaultListPerPage
	}

	total := len(items)
	lastPage := (total + perPage - 1) / perPage
	if lastPage <= 0 {
		lastPage = 1
	}

	if page > lastPage {
		page = lastPage
	}

	start := (page - 1) * perPage
	if start > total {
		start = total
	}

	end := start + perPage
	if end > total {
		end = total
	}

	pageItems := items[start:end]

	var from *int
	var to *int
	if total > 0 && len(pageItems) > 0 {
		fromValue := start + 1
		toValue := start + len(pageItems)
		from = &fromValue
		to = &toValue
	}

	return ListResult{
		Items: pageItems,
		Pagination: Pagination{
			CurrentPage: page,
			LastPage:    lastPage,
			PerPage:     perPage,
			Total:       total,
			From:        from,
			To:          to,
		},
	}
}
