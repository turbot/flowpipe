

integration "slack" "my_slack_app" {
  token           = "xoxp-111111"

  # optional - if you want to verify the source
  signing_secret  = "Q#$$#@#$$#W"
}

integration "email" "email_integration" {
  smtp_host = "foo bar baz"
  default_subject = "bar foo baz"
  smtp_username = "baz bar foo"
}

pipeline "approval" {

  step "input" "input" {
    notify {
      integration = integration.slack.my_slack_app
      channel = "foo"
    }
  }

}