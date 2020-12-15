server "api" {
  api {
    endpoint "/" {
      path = "/yyy"
      backend "anything" {
        path = "/zzz"
      }

      add_query_params = {
        ae = "ae"
      }
    }
  }
}

definitions {
  # backend origin within a definition block gets replaced with the integration test "anything" server.
  backend "anything" {
    path = "/xxx"
    origin = "http://anyserver/"

    add_query_params = {
      def = "def"
    }
  }
}