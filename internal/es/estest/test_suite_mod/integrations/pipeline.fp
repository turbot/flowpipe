pipeline "integration_pipe_default" {

  step "input" "my_step" {

    notifier = notifier.my_notifier

    type   = "button"
    prompt = "Do you want to approve?"

    option "Approve" {}
    option "Deny" {}

  }

  step "transform" "do_the_thing" {
    if       = step.input.my_step.value == "Approve"
    value = "got here"
  }

  output "val" {
    value = step.transform.do_the_thing.value
  }
}
