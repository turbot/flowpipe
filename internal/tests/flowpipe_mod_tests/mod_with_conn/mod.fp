mod "mod_with_conn" {
  title = "mod_with_conn"
}

pipeline "with_conn" {
  step "transform" "echo" {
    value = connection.aws.default.access_key
  }

  step "transform" "from_env" {
    value = env("ACCESS_KEY")
  }
}
