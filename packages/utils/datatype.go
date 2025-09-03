package utils

import (
	"fmt"
	"html"
	"net/mail"
	"regexp"
	"strings"

	"github.com/lifenetwork-ai/iam-service/constants"
)

func IsPhoneNumber(phone string) bool {
	// Check if the phone number is valid
	phoneValidator := regexp.MustCompile(`^(\+?(\d{1,3}))?(\d{10,15})$`)
	return phoneValidator.MatchString(phone)
}

func IsEmail(s string) bool {
	localPartRe := regexp.MustCompile(`^[A-Za-z0-9.!#$%&'*+/=?^_{}|~-]+$`)
	domainRe := regexp.MustCompile(`^(?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?\.)+[A-Za-z]{2,}$`)

	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return false
	}
	e := strings.ToLower(addr.Address)
	at := strings.LastIndexByte(e, '@')
	if at <= 0 || at == len(e)-1 {
		return false
	}
	local := e[:at]
	domain := e[at+1:]

	if !localPartRe.MatchString(local) ||
		strings.HasPrefix(local, ".") ||
		strings.HasSuffix(local, ".") ||
		strings.Contains(local, "..") {
		return false
	}
	// This line catches your case (no dot in domain)
	if !domainRe.MatchString(domain) {
		return false
	}
	if len(local) > 64 || len(e) > 254 {
		return false
	}
	return true
}

func GetIdentifierType(identifier string) (string, error) {
	if IsEmail(identifier) {
		return constants.IdentifierEmail.String(), nil
	}
	if IsPhoneNumber(identifier) {
		return constants.IdentifierPhone.String(), nil
	}
	return "", fmt.Errorf("invalid identifier format")
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
