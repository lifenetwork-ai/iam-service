package constants

// IdentifierType represents the types of identifiers used for login
type IdentifierType string

func (t IdentifierType) String() string {
	return string(t)
}

const (
	IdentifierEmail IdentifierType = "email"
	IdentifierPhone IdentifierType = "phone"
)
