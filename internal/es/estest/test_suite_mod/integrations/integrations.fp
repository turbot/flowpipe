integration "slack" "my_slack_app" {
  token           = "xoxp-111111"

  # optional - if you want to verify the source
  signing_secret  = "Q#$$#@#$$#W"
}

integration "slack" "slack_app_from_var" {
  token           = var.slack_token
  # TODO: this doesn't work
  signing_secret  = var.slack_signing_secret
}
