package config

import (
	"flag"
	"fmt"
	"strings"
)

type AcceptForwarded struct {
	port, protocol, host bool
}

func (a *AcceptForwarded) Set(forwarded []string) error {
	a.protocol = false
	a.host = false
	a.port = false
	for _, part := range forwarded {
		switch strings.TrimSpace(part) {
		case "":
			continue
		case "port":
			a.port = true
		case "proto":
			a.protocol = true
		case "host":
			a.host = true
		default:
			return fmt.Errorf("invalid X-Forwarded-* name (%s)", part)
		}
	}

	return nil
}

func (a AcceptForwarded) String() string {
	var parts []string

	if a.protocol {
		parts = append(parts, "proto")
	}
	if a.host {
		parts = append(parts, "host")
	}
	if a.port {
		parts = append(parts, "port")
	}

	return strings.Join(parts, ",")
}

const otelCollectorEndpoint = "localhost:4317"

// DefaultSettings defines the <DefaultSettings> object.
var DefaultSettings = Settings{
	DefaultPort:              8080,
	Environment:              "",
	HealthPath:               "/healthz",
	LogFormat:                "common",
	LogLevel:                 "info",
	LogPretty:                false,
	NoProxyFromEnv:           false,
	RequestIDBackendHeader:   "Couper-Request-ID",
	RequestIDClientHeader:    "Couper-Request-ID",
	RequestIDFormat:          "common",
	TelemetryMetricsEndpoint: otelCollectorEndpoint,
	TelemetryMetricsExporter: "prometheus",
	TelemetryMetricsPort:     9090, // default prometheus port
	TelemetryServiceName:     "couper",
	TelemetryTracesEndpoint:  otelCollectorEndpoint,
	XForwardedHost:           false,

	// TODO: refactor
	AcceptForwardedURL: []string{},
	AcceptForwarded:    &AcceptForwarded{},
}

// Settings represents the <Settings> object.
type Settings struct {
	AcceptForwarded *AcceptForwarded
	Certificate     []byte

	CAFile                    string   `hcl:"ca_file,optional" docs:"adds the given PEM encoded CA certificate to the existing system certificate pool for all outgoing connections"`
	AcceptForwardedURL        []string `hcl:"accept_forwarded_url,optional" docs:"Which {X-Forwarded-*} request headers should be accepted to change the [request variables](../variables#request) {url}, {origin}, {protocol}, {host}, {port}. Valid values: {\"proto\"}, {\"host\"} and {\"port\"}. The port in a {X-Forwarded-Port} header takes precedence over a port in {X-Forwarded-Host}. Affects relative URL values for [{sp_acs_url}](saml) attribute and {redirect_uri} attribute within [{beta_oauth2}](oauth2) and [{oidc}](oidc)."`
	DefaultPort               int      `hcl:"default_port,optional" docs:"Port which will be used if not explicitly specified per host within the [{hosts}](server) attribute." default:"8080"`
	Environment               string   `hcl:"environment,optional" docs:"[environment](../command-line#global-options) Couper is to run in"`
	HealthPath                string   `hcl:"health_path,optional" docs:"Health path for all configured servers and ports" default:"/healthz"`
	LogFormat                 string   `hcl:"log_format,optional" docs:"tab/field based colored logs or JSON logs: {\"common\"} or {\"json\"}" default:"common"`
	LogLevel                  string   `hcl:"log_level,optional" docs:"sets the log level: {\"panic\"}, {\"fatal\"}, {\"error\"}, {\"warn\"}, {\"info\"}, {\"debug\"}, {\"trace\"}" default:"info"`
	LogPretty                 bool     `hcl:"log_pretty,optional" docs:"global option for {json} log format which pretty prints with basic key coloring"`
	NoProxyFromEnv            bool     `hcl:"no_proxy_from_env,optional" docs:"disables the connect hop to configured [proxy via environment](https://godoc.org/golang.org/x/net/http/httpproxy)"`
	RequestIDAcceptFromHeader string   `hcl:"request_id_accept_from_header,optional" docs:"client request HTTP header field that transports the {request.id} which Couper takes for logging and transport to the backend (if configured)"`
	RequestIDBackendHeader    string   `hcl:"request_id_backend_header,optional" docs:"HTTP header field which Couper uses to transport the {request.id} to the backend" default:"Couper-Request-ID"`
	RequestIDClientHeader     string   `hcl:"request_id_client_header,optional" docs:"HTTP header field which Couper uses to transport the {request.id} to the client" default:"Couper-Request-ID"`
	RequestIDFormat           string   `hcl:"request_id_format,optional" docs:"{\"common\"} or {\"uuid4\"}. If set to {\"uuid4\"} an RFC 4122 UUID is used for {request.id} and related log fields. " default:"common"`
	SecureCookies             string   `hcl:"secure_cookies,optional" docs:"{\"\"} or {\"strip\"}. If set to {\"strip\"}, the {Secure} flag is removed from all {Set-Cookie} HTTP header fields." default:"\u200C"`
	TLSDevProxy               List     `hcl:"https_dev_proxy,optional" docs:"TLS port mappings to define the TLS listen port and the target one. Self-signed certificates will be generated on the fly based on the given hostname. Certificates will be held in memory."`
	TelemetryMetrics          bool     `hcl:"beta_metrics,optional" docs:"enables the Prometheus [metrics](/observation/metrics) exporter"`
	TelemetryMetricsEndpoint  string   `hcl:"beta_metrics_endpoint,optional" docs:"" default:""`
	TelemetryMetricsExporter  string   `hcl:"beta_metrics_exporter,optional" docs:"" default:""`
	TelemetryMetricsPort      int      `hcl:"beta_metrics_port,optional" docs:"Prometheus exporter listen port" default:"9090"`
	TelemetryServiceName      string   `hcl:"beta_service_name,optional" docs:"service name which applies to the {service_name} metric labels" default:"couper"`
	TelemetryTraces           bool     `hcl:"beta_traces,optional" docs:"" default:""`
	TelemetryTracesEndpoint   string   `hcl:"beta_traces_endpoint,optional" docs:"" default:""`
	XForwardedHost            bool     `hcl:"xfh,optional" docs:"whether to use the {X-Forwarded-Host} header as the request host"`
}

var _ flag.Value = &List{}

type List []string

func (s *List) String() string {
	return strings.Join(*s, ",")
}

func (s *List) Set(val string) error {
	if len(*s) > 0 { // argument priority over settings
		*s = nil
	}
	*s = append(*s, strings.Split(val, ",")...)
	return nil
}

func (s *Settings) SetAcceptForwarded() error {
	return s.AcceptForwarded.Set(s.AcceptForwardedURL)
}

func (s *Settings) AcceptsForwardedPort() bool {
	return s.AcceptForwarded.port
}

func (s *Settings) AcceptsForwardedProtocol() bool {
	return s.AcceptForwarded.protocol
}

func (s *Settings) AcceptsForwardedHost() bool {
	return s.AcceptForwarded.host
}
