pipeline "integration_pipe_default_with_param" {
  param "notifier" {
    default = "default"
  }
  step "input" "my_step" {

    notifier = notifier[param.notifier]

    type   = "button"
    prompt = "Do you want to approve?"

    option "Approve" {}
    option "Deny" {}
  }
}