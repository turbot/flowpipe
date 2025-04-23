

integration "slack" "my_slack_app" {
  token           = "xoxp-111111"

  # optional - if you want to verify the source
  signing_secret  = "Q#$$#@#$$#W"
}

integration "slack" "my_slack_app_two" {
  token           = "xoxp-111111"

  # optional - if you want to verify the source
  signing_secret  = "Q#$$#@#$$#W"
}

integration "email" "email_integration" {
  smtp_host       = "foo bar baz"
  default_subject = "bar foo baz"
  smtp_username   = "baz bar foo"
  from            = "test@test.com"
}

pipeline "approval" {
  step "input" "input" {
    type = "button"
    option "test" {}

    notify {
      integration = integration.slack.my_slack_app
      channel = "foo"
    }
  }
}

pipeline "approval_email" {
  step "input" "input_email" {
    type = "button"
    option "test" {}

    notify {
      integration = integration.email.email_integration
      to = "victor@turbot.com"
    }
  }
}

// TODO: param doesn't work yet
pipeline "approval_dynamic_integration" {

  param "integration_param" {
  }

  step "input" "input" {
    type = "button"
    option "test" {}

    notify {
      integration = integration.slack.my_slack_app
      channel = "foo"
    }
  }
}


