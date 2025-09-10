package utils

import (
	"strings"

	"github.com/lifenetwork-ai/iam-service/constants"
)

// NormalizeLang trims and lowercases the lang string.
func NormalizeLang(lang string) string {
	return strings.ToLower(strings.TrimSpace(lang))
}

// IsLangSupported checks if a lang code is in the whitelist.
func IsLangSupported(lang string) bool {
	_, ok := constants.LangSupported[lang]
	return ok
}
