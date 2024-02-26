pipeline "input_webform_button_simple" {


  step "input" "my_step" {
    type     = "button"
    prompt   = "Do you want to approve?"

    notifier = notifier.default

    option "Approve" {}
    option "Deny" {}
  }

  output "val" {
    value = step.input.my_step.value
  }

}