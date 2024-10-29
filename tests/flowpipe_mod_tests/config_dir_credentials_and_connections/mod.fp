
mod "mod_with_connections" {
  title = "mod with connections"
}

pipeline "static_creds_test" {
  step "transform" "aws" {
    value = connection.aws["prod_conn"].profile
  }
}
