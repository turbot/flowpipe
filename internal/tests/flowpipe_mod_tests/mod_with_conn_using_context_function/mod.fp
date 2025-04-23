mod "mod_with_conn_using_context_function" {
  title = "mod_with_conn_using_context_function"
}

pipeline "with_slack_conn" {
  
  step "transform" "from_env" {
    value = connection.slack.slack_conn.token
  }
}
