package domain

import "testing"

func TestPhoneLikePatternsIncludesLeadingZeroVariant(t *testing.T) {
	patterns := PhoneLikePatterns("5555544444")
	found := false
	for _, pattern := range patterns {
		if pattern == "%05555544444%" || pattern == "%5555544444%" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected phone patterns, got %#v", patterns)
	}
}
