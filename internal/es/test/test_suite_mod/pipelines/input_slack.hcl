pipeline "input_slack_pipeline" {
  step "input" "my_step" {
    type   = "slack"

    // token = "Set Your Token Here and Uncomment the token parameter"
    channel = "DF8SL4GR5"

    slack_type = "button"
    prompt = "Do you want to approve?"

    // option "Approve" {}
    // option "Deny" {}

    // notify {
    //   integration = integration.slack.my_slack
    //   channel     = "DF8SL4GR5"
    // }
  }
}