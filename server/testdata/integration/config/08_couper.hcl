server "jwt" {
  api {
    base_path = "/separate"
    endpoint "/{p}/create-jwt" {
      response {
        headers = {
          x-jwt = jwt_sign("my_jwt", {})
        }
      }
    }
    endpoint "/{p}/jwt" {
      access_control = [ "my_jwt" ]
      response {
        json_body = request.context.my_jwt
      }
    }
  }
  api {
    base_path = "/self-signed"
    endpoint "/{p}/create-jwt" {
      response {
        headers = {
          x-jwt = jwt_sign("my_jwt", {groups = []})
        }
      }
    }
    endpoint "/{p}/jwt" {
      access_control = [ "self_signed_jwt" ]
      response {
        json_body = request.context.self_signed_jwt
      }
    }
  }
}
definitions {
  jwt_signing_profile "my_jwt" {
    signature_algorithm = "HS256"
    key = "asdf"
    ttl = "1h"
    claims = {
      iss = "the_issuer"
      pid = request.path_params.p
      groups = ["g1", "g2"]
      tuple = [0, "s", true]
      object = { i = 42, f = 1.23, s = "s", b = true }
    }
  }
  jwt "my_jwt" {
    signature_algorithm = "HS256"
    key = "asdf"
    claims = {
      iss = "the_issuer"
      pid = request.path_params.p
      tuple = [0, "s", true]
      object = { b = true, f = 1.23, i = 42, s = "s"}
    }
  }
  jwt "self_signed_jwt" {
    signature_algorithm = "HS256"
    key = "asdf"
    signing_ttl = "1h"
    claims = {
      iss = "the_issuer"
      pid = request.path_params.p
    }
  }
}
