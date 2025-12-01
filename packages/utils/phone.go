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

	// Optional: quick sanity regex for E.164 (used in IsE164)
	e164Re = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)
)

// NormalizePhoneE164 parses and normalizes a phone number to E.164 format.
//
// Behavior:
//   - Strips spaces/dashes/parentheses.
//   - Converts "00..." international prefix to "+...".
//   - If the input starts with "+", it is parsed globally (no region).
//   - Otherwise, it tries defaultRegion first.
//   - Rejects numbers in the "001" pseudo-region (global/service codes like +800).
//   - Returns the normalized E.164 string and the detected 2-letter region.
//
// This should be called at all boundaries: register, login, add-identifier, OTP verify, check-identifier.
func NormalizePhoneE164(raw, defaultRegion string) (e164, region string, err error) {
	s := sanitizePhoneInput(raw)

	num, err := parsePhoneNumber(s, defaultRegion)
	if err != nil {
		return "", "", err
	}

	region = phonenumbers.GetRegionCodeForNumber(num)
	if region == "001" { // UIFN / global service codes
		return "", "", errInvalidPhone
	}

	return phonenumbers.Format(num, phonenumbers.E164), region, nil
}

// parsePhoneNumber encapsulates parsing/validation with early returns to keep complexity low.
func parsePhoneNumber(s, defaultRegion string) (*phonenumbers.PhoneNumber, error) {
	// Case 1: số có dấu '+' => parse toàn cầu
	if strings.HasPrefix(s, "+") {
		n, err := phonenumbers.Parse(s, "")
		if err != nil || !phonenumbers.IsValidNumber(n) {
			return nil, errInvalidPhone
		}
		return n, nil
	}

	// Case 2: không có '+' => parse theo defaultRegion
	if n, err := phonenumbers.Parse(s, strings.ToUpper(defaultRegion)); err == nil && phonenumbers.IsValidNumber(n) {
		return n, nil
	}

	return nil, errInvalidPhone
}

// sanitizePhoneInput removes common separators and normalizes "00" international prefix to "+".
func sanitizePhoneInput(s string) string {
	s = strings.TrimSpace(s)
	replacer := strings.NewReplacer(
		" ", "", "-", "", "(", "", ")", "", "\u00A0", "",
	)
	s = replacer.Replace(s)

	// Convert "00" international prefix to "+"
	if after, ok := strings.CutPrefix(s, "00"); ok {
		s = "+" + after
	}

	return s
}

// IsPhoneE164 returns true if the input already looks like a normalized E.164 string.
func IsPhoneE164(s string) bool {
	return e164Re.MatchString(strings.TrimSpace(s))
}
