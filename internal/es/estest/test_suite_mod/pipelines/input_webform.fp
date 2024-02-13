pipeline "approval_webform" {

  step "input" "my_step" {
    type     = "button"
    prompt   = "Do you want to approve?"

    option "Approve" {}
    option "Deny" {}
  }

}