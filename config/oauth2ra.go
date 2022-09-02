package config

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

const (
	ClientCredentials = "client_credentials"
	JwtBearer         = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	Password          = "password"
)

var OAuthBlockHeaderSchema = hcl.BlockHeaderSchema{
	Type: "oauth2",
}
var OAuthBlockSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		OAuthBlockHeaderSchema,
	},
}

var (
	_ BackendReference = &OAuth2ReqAuth{}
	_ Inline           = &OAuth2ReqAuth{}
	_ OAuth2Client     = &OAuth2ReqAuth{}
	_ OAuth2AS         = &OAuth2ReqAuth{}
)

// OAuth2ReqAuth represents the oauth2 block in a backend block.
type OAuth2ReqAuth struct {
	AssertionExpr           hcl.Expression `hcl:"assertion,optional" docs:"The assertion (JWT for jwt-bearer flow). Required if {grant_type} is {urn:ietf:params:oauth:grant-type:jwt-bearer}."`
	BackendName             string         `hcl:"backend,optional" docs:"[{backend} block](backend) reference."`
	ClientID                string         `hcl:"client_id,optional" docs:"The client identifier. Required unless the {grant_type} is {urn:ietf:params:oauth:grant-type:jwt-bearer}."`
	ClientSecret            string         `hcl:"client_secret,optional" docs:"The client password. Required unless the {grant_type} is {urn:ietf:params:oauth:grant-type:jwt-bearer}."`
	GrantType               string         `hcl:"grant_type" docs:"Required, valid values: {client_credentials}, {password}, {urn:ietf:params:oauth:grant-type:jwt-bearer}"`
	Password                string         `hcl:"password,optional" docs:"The (service account's) password (for password flow). Required if grant_type is {password}."`
	Remain                  hcl.Body       `hcl:",remain"`
	Retries                 *uint8         `hcl:"retries,optional" default:"1" docs:"The number of retries to get the token and resource, if the resource-request responds with {401 Unauthorized} HTTP status code."`
	Scope                   string         `hcl:"scope,optional" docs:"A space separated list of requested scope values for the access token."`
	TokenEndpoint           string         `hcl:"token_endpoint,optional" docs:"URL of the token endpoint at the authorization server."`
	TokenEndpointAuthMethod *string        `hcl:"token_endpoint_auth_method,optional" default:"client_secret_basic" docs:"Defines the method to authenticate the client at the token endpoint."`
	Username                string         `hcl:"username,optional" docs:"The (service account's) username (for password flow). Required if grant_type is {password}."`
}

// Reference implements the <BackendReference> interface.
func (oa *OAuth2ReqAuth) Reference() string {
	return oa.BackendName
}

// HCLBody implements the <Inline> interface.
func (oa *OAuth2ReqAuth) HCLBody() hcl.Body {
	return oa.Remain
}

// Inline implements the <Inline> interface.
func (oa *OAuth2ReqAuth) Inline() interface{} {
	type Inline struct {
		Backend *Backend `hcl:"backend,block"`
	}

	return &Inline{}
}

// Schema implements the <Inline> interface.
func (oa *OAuth2ReqAuth) Schema(inline bool) *hcl.BodySchema {
	if !inline {
		schema, _ := gohcl.ImpliedBodySchema(oa)
		return schema
	}

	schema, _ := gohcl.ImpliedBodySchema(oa.Inline())

	// A backend reference is defined, backend block is not allowed.
	if oa.BackendName != "" {
		schema.Blocks = nil
	}

	return schema
}

func (oa *OAuth2ReqAuth) ClientAuthenticationRequired() bool {
	return oa.GrantType != JwtBearer
}

func (oa *OAuth2ReqAuth) GetClientID() string {
	return oa.ClientID
}

func (oa *OAuth2ReqAuth) GetClientSecret() string {
	return oa.ClientSecret
}

func (oa *OAuth2ReqAuth) GetTokenEndpoint() (string, error) {
	return oa.TokenEndpoint, nil
}

func (oa *OAuth2ReqAuth) GetTokenEndpointAuthMethod() *string {
	return oa.TokenEndpointAuthMethod
}
