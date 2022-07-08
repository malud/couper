# Settings

The `settings` block lets you configure the more basic and global behavior of your
gateway instance.

## Access Control

The configuration of access control is twofold in Couper: You define the particular
type (such as `jwt` or `basic_auth`) in `definitions`, each with a distinct label (must not be one of the reserved names: `beta_granted_permissions`, `beta_required_permission`).
Anywhere in the `server` block those labels can be used in the `access_control`
list to protect that block. &#9888; access rights are inherited by nested blocks.
You can also disable `access_control` for blocks. By typing `disable_access_control = ["bar"]`,
the `access_control` type `bar` will be disabled for the corresponding block context.

All access controls have an option to handle related errors. Please refer to [Errors](ERRORS.md).

## Health-Check

The health check will answer a status `200 OK` on every port with the configured
`health_path`. As soon as the gateway instance will receive a `SIGINT` or `SIGTERM`
the check will return a status `500 StatusInternalServerError`. A shutdown delay
of `5s` for example allows the server to finish all running requests and gives a load-balancer
time to pick another gateway instance. After this delay the server goes into
shutdown mode with a deadline of `5s` and no new requests will be accepted.
The shutdown timings defaults to `0` which means no delaying with development setups.
Both durations can be configured via environment variable. Please refer to the [docker document](../DOCKER.md).

## Error Handler Block

The `error_handler` block lets you configure the handling of errors thrown in components configured by the parent blocks.

The error handler label specifies which [error type](ERRORS.md#error-types) should be handled. Multiple labels are allowed. The label can be omitted to catch all relevant errors. This has the same behavior as the error type `*`, that catches all errors explicitly.

Concerning child blocks and attributes, the `error_handler` block is similar to an [Endpoint Block](#endpoint-block).

| Block name  |Context|Label|Nested block(s)|
| :-----------| :-----------| :-----------| :-----------|
| `error_handler` | [API Block](#api-block), [Endpoint Block](#endpoint-block), [Basic Auth Block](#basic-auth-block), [JWT Block](#jwt-block), [OAuth2 AC Block (Beta)](#oauth2-ac-block-beta), [OIDC Block](#oidc-block), [SAML Block](#saml-block) | optional | [Proxy Block(s)](#proxy-block),  [Request Block(s)](#request-block), [Response Block](#response-block), [Error Handler Block(s)](#error-handler-block) |

| Attribute(s)            | Type             | Default | Description                                                                                                       | Characteristic(s)                                                                                                                                                                                                                                                                                                                                                                                                                               | Example                                                              |
|:------------------------|:-----------------|:--------|:------------------------------------------------------------------------------------------------------------------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:---------------------------------------------------------------------|
| `custom_log_fields`     | object           | -       | Defines log fields for [Custom Logging](LOGS.md#custom-logging).                                                  | &#9888; Inherited by nested blocks.                                                                                                                                                                                                                                                                                                                                                                                                             | -                                                                    |
| [Modifiers](#modifiers) | -                | -       | -                                                                                                                 | -                                                                                                                                                                                                                                                                                                                                                                                                                                               | -                                                                    |

Examples:

- [Error Handling for Access Controls](https://github.com/avenga/couper-examples/blob/master/error-handling-ba/README.md).


| Attribute(s)                    | Type           | Default             | Description                                                                                                                                                                                                                                                                             | Characteristic(s)                                                                                                                                                  | Example                     |
|:--------------------------------|:---------------|:--------------------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------|:----------------------------|
| `accept_forwarded_url`          | tuple (string) | `[]`                | Which `X-Forwarded-*` request headers should be accepted to change the [request variables](#request) `url`, `origin`, `protocol`, `host`, `port`. Valid values are `"proto"`, `"host"` and `"port"`. The port in `X-Forwarded-Port` takes precedence over a port in `X-Forwarded-Host`. | Affects relative url values for [`sp_acs_url`](#saml-block) attribute and `redirect_uri` attribute within [beta_oauth2](#beta-oauth2-block) & [oidc](#oidc-block). | `["proto","host","port"]`   |
| `default_port`                  | number         | `8080`              | Port which will be used if not explicitly specified per host within the [`hosts`](#server-block) list.                                                                                                                                                                                  | -                                                                                                                                                                  | -                           |
| `health_path`                   | string         | `"/healthz"`        | Health path which is available for all configured server and ports.                                                                                                                                                                                                                     | -                                                                                                                                                                  | -                           |
| `https_dev_proxy`               | tuple (string) | `[]`                | List of tls port mappings to define the tls listen port and the target one. A self-signed certificate will be generated on the fly based on given hostname.                                                                                                                             | Certificates will be hold in memory and are generated once.                                                                                                        | `["443:8080", "8443:8080"]` |
| `log_format`                    | string         | `"common"`          | Switch for tab/field based colored view or JSON log lines. Valid values are `"common"` and `"json"`.                                                                                                                                                                                    | -                                                                                                                                                                  | -                           |
| `log_level`                     | string         | `"info"`            | Set the log-level to one of: `"info"`, `"panic"`, `"fatal"`, `"error"`, `"warn"`, `"debug"`, `"trace"`.                                                                                                                                                                                 | -                                                                                                                                                                  | -                           |
| `log_pretty`                    | bool           | `false`             | Global option for `json` log format which pretty prints with basic key coloring.                                                                                                                                                                                                        | -                                                                                                                                                                  | -                           |
| `no_proxy_from_env`             | bool           | `false`             | Disables the connect hop to configured [proxy via environment](https://godoc.org/golang.org/x/net/http/httpproxy).                                                                                                                                                                      | -                                                                                                                                                                  | -                           |
| `request_id_accept_from_header` | string         | `""`                | Name of a client request HTTP header field that transports the `request.id` which Couper takes for logging and transport to the backend (if configured).                                                                                                                                | -                                                                                                                                                                  | `X-UID`                     |
| `request_id_backend_header`     | string         | `Couper-Request-ID` | Name of a HTTP header field which Couper uses to transport the `request.id` to the backend.                                                                                                                                                                                             | -                                                                                                                                                                  | -                           |
| `request_id_client_header`      | string         | `Couper-Request-ID` | Name of a HTTP header field which Couper uses to transport the `request.id` to the client.                                                                                                                                                                                              | -                                                                                                                                                                  | -                           |
| `request_id_format`             | string         | `"common"`          | Valid values are `"common"` and `"uuid4"`. If set to `"uuid4"` a rfc4122 uuid is used for `request.id` and related log fields.                                                                                                                                                          | -                                                                                                                                                                  | -                           |
| `secure_cookies`                | string         | `""`                | Valid values are `""` and `"strip"`. If set to `"strip"`, the `Secure` flag is removed from all `Set-Cookie` HTTP header fields.                                                                                                                                                        | -                                                                                                                                                                  | -                           |
| `xfh`                           | bool           | `false`             | Option to use the `X-Forwarded-Host` header as the request host.                                                                                                                                                                                                                        | -                                                                                                                                                                  | -                           |
| `beta_metrics`                  | bool           | `false`             | Option to enable the Prometheus [metrics](METRICS.md) exporter.                                                                                                                                                                                                                         | -                                                                                                                                                                  | -                           |
| `beta_metrics_port`             | number         | `9090`              | Prometheus exporter listen port.                                                                                                                                                                                                                                                        | -                                                                                                                                                                  | -                           |
| `beta_service_name`             | string         | `"couper"`          | The service name which applies to the `service_name` metric labels.                                                                                                                                                                                                                     | -                                                                                                                                                                  | -                           |
| `ca_file`                       | string         | `""`                | Option for adding the given PEM encoded ca-certificate to the existing system certificate pool for all outgoing connections.                                                                                                                                                            | -                                                                                                                                                                  | -                           |