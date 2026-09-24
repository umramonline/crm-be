package domain

import "testing"

func TestTaskNormalizeListSortBy(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"title", "title"},
		{"customer_count", "customer_count"},
		{"due_date", "due_date"},
		{"priority", "priority"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		if got := NormalizeListSortBy(tt.in); got != tt.want {
			t.Fatalf("NormalizeListSortBy(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
