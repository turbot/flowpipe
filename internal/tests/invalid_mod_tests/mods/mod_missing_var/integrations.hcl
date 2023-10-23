integration "slack" "slack_app_from_var" {
  token           = var.slack_token
  signing_secret  = var.slack_signing_secret
}
