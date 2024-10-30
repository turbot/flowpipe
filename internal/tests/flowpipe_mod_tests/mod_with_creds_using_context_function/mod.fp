mod "mod_with_creds_using_context_function" {
  title = "mod_with_creds_using_context_function"
}

pipeline "with_slack_creds" {
  
  step "transform" "from_env" {
    value = credential.slack.slack_creds.token
  }
}
