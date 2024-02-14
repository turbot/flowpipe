pipeline "input_opt_block_resolution" {
  step "transform" "get_approve_text" {
    value = "yes"
  }

  step "transform" "get_deny_text" {
    value = "no"
  }

  step "input" "input_test" {
    notifier = notifier["default"]
    type     = "button"
    prompt   = "do you approve?"

    option "Approve" {
      label = "Approve"
      value = step.transform.get_approve_text.value
    }

    option "Deny" {
      label = "Deny"
      value = step.transform.get_deny_text.value
    }
  }

  output "output" {
    value = step.input.input_test.value
  }
}

pipeline "input_opts_att_resolution" {
  step "transform" "get_approve_text" {
    value = "yes"
  }

  step "transform" "get_deny_text" {
    value = "no"
  }

  step "input" "input_test" {
    notifier = notifier["default"]
    type     = "button"
    prompt   = "do you approve?"
    options  = [
      {
        label = "Approve"
        value = step.transform.get_approve_text.value
      },
      {
        label = "Deny"
        value = step.transform.get_deny_text.value
      },
    ]
  }

  output "output" {
    value = step.input.input_test.value
  }
}