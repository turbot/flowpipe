mod "mod_with_creds_resolved" {
  title = "mod_with_creds_resolved"
}

pipeline "static_creds_test" {

  step "transform" "slack" {
    value = credential.slack["slack_static"].token
  }
}
