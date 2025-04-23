
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

pipeline "approval_with_override_in_step" {

  step "input" "my_step" {
    notifier = notifier["admins"]

    type     = "button"
    prompt   = "Do you want to approve?"

    subject = "this subject is in step"

    channel = "this channel is in step override"

    to = ["foo", "bar", "baz override"]
    cc = ["foo", "bar", "baz cc"]
    bcc = ["foo bb", "bar", "baz override"]


    option "Approve" {}
    option "Deny" {}
  }
}


pipeline "approval_with_notifies_dynamic" {

  param "notifier" {
    type = string
    default = "wrong"
  }

  step "input" "my_step" {
    notifier = notifier[param.notifier]

    type     = "button"
    prompt   = "Do you want to approve?"

    option "Approve" {}
    option "Deny" {}
  }
}
