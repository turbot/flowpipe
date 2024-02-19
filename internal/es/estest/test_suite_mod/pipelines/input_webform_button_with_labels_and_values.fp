pipeline "input_webform_button_with_labels_and_values" {

  step "input" "my_step" {
    type     = "button"
    prompt   = "Do you want to approve?"

    option "approve_button" {
      label = "Approve"
      value = "approve_button_pressed"
    }

    option "deny_button" {
      label = "Deny"
      value = "deny_button_pressed"
    }
  }

  output "val" {
    value = step.input.my_step.value
  }

}
