package request

type ContextKey uint8

const (
	ContextType ContextKey = iota
	ContextVariablesSynced
	APIName
	AccessControls
	BackendName
	BufferOptions
	Endpoint
	EndpointExpectedStatus
	EndpointKind
	EndpointSequenceDependsOn
	Error
	Handler
	LogCustomAccess
	LogCustomUpstream
	LogDebugLevel
	LogEntry
	OpenAPI
	PathParams
	ResponseBlock
	ResponseWriter
	RoundTripName
	RoundTripProxy
	Scopes
	ServerName
	StartTime
	TokenRequest
	TokenRequestRetries
	UID
	URLAttribute
	WebsocketsAllowed
	WebsocketsTimeout
	Wildcard
	XFF
)
