pipeline "input_webform_button_multiple" {

  step "input" "my_step" {
    notifier = notifier.default

    type     = "button"
    prompt   = "Do you want to approve?"

    option "Approve1" {}
    option "Approve2" {}
    option "Approve3" {}
    option "Approve4" {}
    option "Approve5" {}
    option "Approve6" {}
    option "Deny1" {}
    option "Deny2" {}
    option "Deny3" {}
    option "Deny4" {}
    option "Deny5" {}
    option "Deny6" {}
  }

  output "val" {
    value = step.input.my_step.value
  }

}