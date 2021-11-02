// Code generated by go generate; DO NOT EDIT.

package errors

var (
	BasicAuth                   = Definitions[0]
	BasicAuthCredentialsMissing = Definitions[1]
	Jwt                         = Definitions[2]
	JwtTokenExpired             = Definitions[3]
	JwtTokenInvalid             = Definitions[4]
	JwtTokenMissing             = Definitions[5]
	Oauth2                      = Definitions[6]
	Saml2                       = Definitions[7]
	BetaOperationDenied         = Definitions[8]
	BetaScope                   = Definitions[9]
	BetaInsufficientScope       = Definitions[10]
)

// typeDefinitions holds all related error definitions which are
// catchable with an error_handler definition.
type typeDefinitions map[string]*Error

// types holds all implemented ones. The name must match the structs
// snake-name for fallback purposes. See TypeToSnake usage and reference.
var types = typeDefinitions{
	"basic_auth":                     BasicAuth,
	"basic_auth_credentials_missing": BasicAuthCredentialsMissing,
	"jwt":                            Jwt,
	"jwt_token_expired":              JwtTokenExpired,
	"jwt_token_invalid":              JwtTokenInvalid,
	"jwt_token_missing":              JwtTokenMissing,
	"oauth2":                         Oauth2,
	"saml2":                          Saml2,
	"beta_operation_denied":          BetaOperationDenied,
	"beta_scope":                     BetaScope,
	"beta_insufficient_scope":        BetaInsufficientScope,
}

// IsKnown tells the configuration callee if Couper
// has a defined error type with the given name.
func IsKnown(errorType string) bool {
	_, known := types[errorType]
	return known
}
