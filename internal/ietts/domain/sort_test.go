package domain

import "testing"

func TestNormalizeListSortBy(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"company_name", "company_name"},
		{"  CITY ", "city"},
		{"document_number", "document_number"},
		{"customer_id", ""},
		{"; DROP TABLE", ""},
	}

	for _, tt := range tests {
		if got := NormalizeListSortBy(tt.in); got != tt.want {
			t.Fatalf("NormalizeListSortBy(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
