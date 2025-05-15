package utils

import (
	"html"
	"regexp"
	"strings"
)

func IsPhoneNumber(phone string) bool {
	// Check if the phone number is valid
	phoneValidator := regexp.MustCompile(`^(\+?(\d{1,3}))?(\d{10,15})$`)
	return phoneValidator.MatchString(phone)
}

func IsEmail(email string) bool {
	// Check if the email is valid
	emailValidator := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailValidator.MatchString(email)
}

// SafeString sanitizes input to prevent SQL injection and XSS attacks.
func SafeString(input string) string {
	// Trim spaces
	safe := strings.TrimSpace(input)

	// Escape HTML to prevent XSS
	safe = html.EscapeString(safe)

	// Remove potentially dangerous SQL injection patterns
	sqlInjectionPattern := regexp.MustCompile(`(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE|EXEC|UNION|OR|AND)\b|(--|;))`)
	safe = sqlInjectionPattern.ReplaceAllString(safe, "")

	return safe
}

// IsSQLInjection checks if the input contains common SQL injection patterns
func IsSQLInjection(input string) bool {
	// Convert to lowercase for case-insensitive comparison
	lowerInput := strings.ToLower(input)

	// List of suspicious SQL keywords and characters
	sqlPatterns := []string{
		"select ", "insert ", "update ", "delete ", "drop ", "alter ", "truncate ",
		"union ", "exec ", "or ", "and ", "like ", "benchmark(", "sleep(", "load_file(",
		"outfile ", "--", ";", "#", "/*", "xp_", "declare ", "cast(", "convert(",
	}

	// Check for common SQL injection keywords
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	// Regular expression to detect suspicious SQL syntax patterns
	sqlRegex := regexp.MustCompile(`(?i)(\b(select|insert|update|delete|drop|alter|truncate|union|exec|or|and)\b|(--|;|#|/\*|\*/|xp_|declare|benchmark\(|sleep\(|load_file\(|outfile))`)

	return sqlRegex.MatchString(input)
}
