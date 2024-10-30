mod "mod_with_input_step_simple" {
  title = "mod_with_integration"
}

pipeline "simple_input_step" {
  step "input" "my_step" {
    notifier = notifier.default

    type     = "button"
    prompt   = "Do you want to approve?"

    option "Approve" {}
    option "Deny" {}
  }
}

pipeline "simple_input_step_with_option_list" {
  step "input" "my_step" {
    notifier = notifier.default

    type     = "button"
    prompt   = "Do you want to approve?"

    options = [
      {
        value     = "us-east-1"
        label     = "N. Virginia"
        selected  = true
      },
      {
        value     = "us-east-2"
        label     = "Ohio"
        selected  = true
      },
    ]
  }
}


