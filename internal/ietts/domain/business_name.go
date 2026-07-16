package domain

import "strings"

func SplitBusinessName(value string) (ad string, soyad string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", ""
	}

	parts := strings.Fields(trimmed)
	if len(parts) == 1 {
		return parts[0], ""
	}

	return parts[0], strings.Join(parts[1:], " ")
}

func StringValue(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}

func TruncateRunes(value string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= maxLen {
		return value
	}

	return string(runes[:maxLen])
}
