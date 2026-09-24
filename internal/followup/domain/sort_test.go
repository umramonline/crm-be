package domain

import "testing"

func TestFollowUpNormalizeListSortBy(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"title", "title"},
		{"customer", "customer"},
		{"visit_date", "visit_date"},
		{"agreement_reached", "agreement_reached"},
		{"unknown", ""},
		{"'; DROP TABLE--", ""},
	}

	for _, tt := range tests {
		if got := NormalizeListSortBy(tt.in); got != tt.want {
			t.Fatalf("NormalizeListSortBy(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
