server "scoped" {

  access_control = ["secured"]

  api {
    base_path = "/api"

    beta_scope = "read"
    endpoint "/" {
      response {
        status = 204
      }
    }

    endpoint "/pow/" {
      beta_scope = {
        post = "power"
      }

      response {
        status = 204
      }

      error_handler "beta_insufficient_scope" {
        response {
          status = 400
          body = "Not enough power"
        }
      }

      error_handler "beta_operation_denied" {
        response {
          status = 405
          body = "Not enough power"
        }
      }
    }

    error_handler "beta_insufficient_scope" {
      response {
        status = 418
      }
    }

  }

  api {
    base_path = "/wildcard1"

    error_handler "beta_insufficient_scope" {
      response {
        status = 418
        body = "Not enough power"
      }
    }

    endpoint "/" {
      beta_scope = "power"

      response {
        status = 204
      }

      error_handler "*" {
        response {
          status = 400
          body = "Not enough power"
        }
      }
    }
  }

  api {
    base_path = "/wildcard2"

    error_handler {
      response {
        status = 418
        body = "Not enough power"
      }
    }

    endpoint "/" {
      beta_scope = "power"

      response {
        status = 204
      }

      error_handler "beta_insufficient_scope" {
        response {
          status = 400
          body = "Not enough power"
        }
      }
    }
  }

  endpoint "/" {
    beta_scope = "write"

    response {
      body = "OK"
    }

    error_handler "beta_scope" {
      response {
        status = 418
      }
    }

  }
}

definitions {
  jwt "secured" {
    header = "Authorization"
    signature_algorithm = "HS256"
    key = "s3cr3t"
    beta_scope_claim = "scopes"
  }
}
