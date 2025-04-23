mod "mod_with_creds" {
  title = "mod_with_creds"
}

pipeline "with_creds" {
  step "transform" "echo" {
    value = credential.aws.default.access_key
  }

  step "transform" "from_env" {
    value = env("ACCESS_KEY")
  }
}
