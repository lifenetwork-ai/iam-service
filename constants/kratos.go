package constants

// IdentifierType represents the types of identifiers used for login
type IdentifierType string

func (t IdentifierType) String() string {
	return string(t)
}

const (
	IdentifierEmail    IdentifierType = "email"
	IdentifierPhone    IdentifierType = "phone_number"
	IdentifierUsername IdentifierType = "username"
	IdentifierTenant   IdentifierType = "tenant"
)

// MethodType represents the types of login/registration methods
type MethodType string

func (t MethodType) String() string {
	return string(t)
}

const (
	MethodTypePassword MethodType = "password"
	MethodTypeCode     MethodType = "code"
)
