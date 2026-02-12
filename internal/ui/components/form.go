package components

import "strings"

func ValidateCompanyName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "Company name is required."
	}
	if len([]rune(trimmed)) < 2 || len([]rune(trimmed)) > 40 {
		return "Use 2-40 characters."
	}
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' || r == '\'' || r == '&' {
			continue
		}
		return "Only letters, numbers, spaces, -, ', & are allowed."
	}
	return ""
}
