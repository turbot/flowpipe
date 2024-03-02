pipeline "input_with_no_options_button_type" {
  step "input" "test" {
    notifier  = notifier.default
    prompt    = "Prompt"
    type      = "button"
  }
}

pipeline "input_with_no_options_text_type" {
  step "input" "test" {
    notifier  = notifier.default
    prompt    = "Prompt"
    type      = "text"
  }
}

pipeline "input_with_slack_notifier_no_channel_set" {
  step "input" "test" {
    notifier  = notifier.bare_minimum_slack
    prompt    = "Prompt"
    type      = "button"
    option "a" {}
    option "b" {}
  }
}

pipeline "input_with_slack_notifier_no_channel_set_wh" {
  step "input" "test" {
    notifier  = notifier.bare_minimum_slack_wh
    prompt    = "Prompt"
    type      = "button"
    option "a" {}
    option "b" {}
  }
}

pipeline "input_with_email_notifier_no_recipients" {
  step "input" "test" {
    notifier  = notifier.bare_minimum_email
    prompt    = "Prompt"
    type      = "select"
    option "a" {}
    option "b" {}
  }
}

