package utils

import (
	"errors"
	"regexp"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

var (
	// errInvalidPhone is returned when the input cannot be parsed or is not a valid phone number.
	errInvalidPhone = errors.New("invalid phone number")

	// errPhoneNotAllowed is returned when the parsed number belongs to a region that is not allowed.
	errPhoneNotAllowed = errors.New("phone number not allowed for this service")

	// Optional: quick sanity regex for E.164 (used in IsE164)
	e164Re = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

	// Allowed regions for phone numbers
	allowedRegions = []string{"VN", "TH", "ID", "KR", "CN"}
)

// NormalizePhoneE164 parses and normalizes a phone number to E.164.
// It also enforces an allowlist of regions (ISO 3166-1 alpha-2, e.g., "VN", "TH", "ID", "KR", "CN").
//
// Behavior:
//   - Strips spaces/dashes/parentheses.
//   - Converts "00..." international prefix to "+...".
//   - If the input starts with "+", it is parsed without a default region.
//   - Otherwise, it first tries defaultRegion, then (optionally) tries the remaining allowedRegions.
//   - Rejects numbers in the "001" pseudo-region (global/service codes like +800).
//   - Returns the normalized E.164 string and the detected 2-letter region.
//
// This should be called at all boundaries: register, login, add-identifier, OTP verify, check-identifier.
func NormalizePhoneE164(raw, defaultRegion string) (e164, region string, err error) {
	s := sanitizePhoneInput(raw)

	num, err := parsePhoneNumber(s, defaultRegion, allowedRegions)
	if err != nil {
		return "", "", err
	}

	region = phonenumbers.GetRegionCodeForNumber(num)
	if region == "001" { // UIFN / global service codes
		return "", "", errInvalidPhone
	}

	if len(allowedRegions) > 0 && !containsRegion(allowedRegions, region) {
		return "", "", errPhoneNotAllowed
	}

	return phonenumbers.Format(num, phonenumbers.E164), region, nil
}

// parsePhoneNumber encapsulates parsing/validation with early returns to keep complexity low.
func parsePhoneNumber(s, defaultRegion string, allowed []string) (*phonenumbers.PhoneNumber, error) {
	if strings.HasPrefix(s, "+") {
		n, err := phonenumbers.Parse(s, "")
		if err != nil || !phonenumbers.IsValidNumber(n) {
			return nil, errInvalidPhone
		}
		return n, nil
	}

	// Try default region first.
	if n, err := phonenumbers.Parse(s, strings.ToUpper(defaultRegion)); err == nil && phonenumbers.IsValidNumber(n) {
		return n, nil
	}

	// Optionally try the remaining allowed regions (only when no '+').
	for _, r := range allowed {
		if strings.EqualFold(r, defaultRegion) {
			continue
		}
		if n, err := phonenumbers.Parse(s, strings.ToUpper(r)); err == nil && phonenumbers.IsValidNumber(n) {
			return n, nil
		}
	}
	return nil, errInvalidPhone
}

// sanitizePhoneInput removes common separators and normalizes "00" international prefix to "+".
func sanitizePhoneInput(s string) string {
	s = strings.TrimSpace(s)
	// Remove common formatting chars.
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", "\u00A0", "")
	s = replacer.Replace(s)
	// Convert international prefix "00..." to "+..."
	if after, ok := strings.CutPrefix(s, "00"); ok {
		s = "+" + after
	}
	return s
}

// containsRegion checks if region (case-insensitive) is in the allowlist.
func containsRegion(allow []string, r string) bool {
	for _, a := range allow {
		if strings.EqualFold(a, r) {
			return true
		}
	}
	return false
}

// IsPhoneE164 returns true if the input already looks like a normalized E.164 string.
func IsPhoneE164(s string) bool {
	return e164Re.MatchString(strings.TrimSpace(s))
}
