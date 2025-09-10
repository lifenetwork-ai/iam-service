package constants

// Supported languages
const (
	LangVI = "vi"
	LangEN = "en"
)

// LangSupported is the whitelist of allowed languages.
var LangSupported = map[string]struct{}{
	LangVI: {},
	LangEN: {},
}
