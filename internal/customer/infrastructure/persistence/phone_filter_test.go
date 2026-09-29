package persistence

import (
	"testing"

	"github.com/umran/new.crm/backend/internal/customer/domain"
)

func TestPhoneLikePatternsNormalizesTurkishMobile(t *testing.T) {
	patterns := domain.PhoneLikePatterns("532 123 45 67")
	if len(patterns) < 2 {
		t.Fatalf("expected multiple patterns, got %#v", patterns)
	}

	found := false
	for _, pattern := range patterns {
		if pattern == "%05321234567%" || pattern == "%5321234567%" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 05 or digit-only pattern, got %#v", patterns)
	}
}
