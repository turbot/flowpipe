mod "local" {

}

integration "slack" "my_slack_app" {
  token           = var.slack_token

  # optional - if you want to verify the source
  signing_secret  = "Q#$$#@#$$#W"
}

integration "slack" "my_slack_app_two" {
  token           = "token for my slack app two"

  # optional - if you want to verify the source
  signing_secret  = "Q#$$#@#$$#W"
}

pipeline "get_integration" {
  output "integration" {
    value = integration.my_slack_app_two
  }
}

pipeline "approval_with_depends_on" {

  step "pipeline" "get_integration" {
    pipeline = pipeline.get_integration
  }

  step "input" "input" {
    token = "remove this after integrated"
    notify {
      integration = step.pipeline.get_integration.integration
      channel = var.channel_name
    }
  }
}

variable "channel_name" {
  type = string
  default = "bar"
}

variable "slack_token" {
  type = string
  default = "just the default here"
}
