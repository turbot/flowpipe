pipeline "input_webform_button_styles" {


  step "input" "my_step" {
    type     = "button"
    prompt   = "Do you want to approve?"

    notifier = notifier.default

    option "Approve" {
      style = "ok"
    }
    option "Deny" {
      style = "alert"
    }
  }

  output "val" {
    value = step.input.my_step.value
  }

}