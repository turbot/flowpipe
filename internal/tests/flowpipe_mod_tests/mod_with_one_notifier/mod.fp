
mod "mod_with_integration" {
  title = "mod_with_integration"
}

pipeline "approval_with_notifies" {

  step "input" "my_step" {
    notifier = notifier["admins"]

    type     = "button"
    prompt   = "Do you want to approve?"

    option "Approve" {}
    option "Deny" {}
  }

  step "input" "my_step_2" {
    notifier = notifier.admins

    type     = "button"
    prompt   = "Do you want to approve (2)?"

    option "Approve" {}
    option "Deny" {}
  }

}


