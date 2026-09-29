package domain

import "testing"

func TestNormalizeListQuery(t *testing.T) {
	query := NormalizeListQuery(ListQuery{
		Page:      0,
		PerPage:   500,
		SortBy:    "unknown",
		SortOrder: "ASC",
	})

	if query.Page != 1 || query.PerPage != MaxListPerPage || query.SortBy != "" || query.SortOrder != "asc" {
		t.Fatalf("unexpected normalized query: %#v", query)
	}
}

func TestNormalizeListQueryStripsAllFilterLabel(t *testing.T) {
	query := NormalizeListQuery(ListQuery{
		Situation:  "Tümü",
		BranchName: "Tümü",
		ZoneName:   "Tümü",
		Type:       "Tümü",
		Cep:        "05537",
	})

	if query.Situation != "" || query.BranchName != "" || query.ZoneName != "" || query.Type != "" {
		t.Fatalf("expected Tümü filters cleared, got %#v", query)
	}
	if query.Cep != "05537" {
		t.Fatalf("expected cep preserved, got %q", query.Cep)
	}
}

func TestPaginateCustomers(t *testing.T) {
	items := make([]Customer, 25)
	for i := range items {
		items[i].ID = uint64(i + 1)
	}

	result := PaginateCustomers(items, 2, 10)
	if len(result.Items) != 10 || result.Pagination.Total != 25 || result.Pagination.CurrentPage != 2 {
		t.Fatalf("unexpected pagination: %#v", result.Pagination)
	}
}
