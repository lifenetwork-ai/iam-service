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

// MethodType represents the types of login/registration/setting methods
type MethodType string

func (t MethodType) String() string {
	return string(t)
}

const (
	MethodTypePassword MethodType = "password"
	MethodTypeCode     MethodType = "code"
	MethodTypeProfile  MethodType = "profile"
)

// FlowType represents the types of flows in the identity service
type FlowType string

func (t FlowType) String() string {
	return string(t)
}

const (
	FlowTypeLogin    FlowType = "login"
	FlowTypeRegister FlowType = "register"
	FlowTypeSetting  FlowType = "setting"
)
