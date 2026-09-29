package domain

import "testing"

func TestCustomerNormalizeListSortBy(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"credit", "credit"},
		{"point", "point"},
		{"created_at", "created_at"},
		{"vehicle_stock_count", "vehicle_stock_count"},
		{"CREATED_AT", "created_at"},
		{"unknown", ""},
		{"'; DROP TABLE--", ""},
	}

	for _, tt := range tests {
		if got := NormalizeListSortBy(tt.in); got != tt.want {
			t.Fatalf("NormalizeListSortBy(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
