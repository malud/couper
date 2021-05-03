server "oauth2-options" {
  error_file = "./../integration/server_error.html"

  api {
    error_file = "./../integration/api_error.json"

    endpoint "/" {
      proxy {
        backend {
          url = "https://example.com/"

          oauth2 {
            token_endpoint = "${request.headers.x-token-endpoint}/options"
            retries        = 0
            client_id      = "user"
            client_secret  = "pass word"
            grant_type     = "client_credentials"
            scope          = "scope1 scope2"
            token_endpoint_auth_method = "client_secret_post"
          }
        }
      }
    }
  }
}
