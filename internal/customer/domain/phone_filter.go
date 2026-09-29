package domain

import (
	"strings"
	"unicode"
)

func PhoneDigits(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func PhoneLikePatterns(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	digits := PhoneDigits(trimmed)
	patterns := make([]string, 0, 4)
	seen := map[string]struct{}{}

	add := func(pattern string) {
		if pattern == "" {
			return
		}
		if _, ok := seen[pattern]; ok {
			return
		}
		seen[pattern] = struct{}{}
		patterns = append(patterns, pattern)
	}

	if digits == "" {
		add("%" + trimmed + "%")
		return patterns
	}

	add("%" + digits + "%")
	if strings.HasPrefix(digits, "0") && len(digits) > 1 {
		add("%" + strings.TrimPrefix(digits, "0") + "%")
	}
	if !strings.HasPrefix(digits, "0") && len(digits) >= 9 {
		add("%0" + digits + "%")
	}
	if len(digits) >= 10 {
		add("%" + digits[len(digits)-10:] + "%")
	}

	return patterns
}

func CustomerPhoneMatchesFilter(customer Customer, raw string) bool {
	patterns := PhoneLikePatterns(raw)
	if len(patterns) == 0 {
		return true
	}

	phone := strings.TrimSpace(customer.Cep)
	for _, pattern := range patterns {
		substr := strings.Trim(pattern, "%")
		if substr != "" && strings.Contains(phone, substr) {
			return true
		}
	}

	return false
}
