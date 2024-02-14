pipeline "input_webform_button_simple" {

  step "input" "my_step" {
    type     = "button"
    prompt   = "Do you want to approve?"

    option "Approve" {}
    option "Deny" {}
  }

}