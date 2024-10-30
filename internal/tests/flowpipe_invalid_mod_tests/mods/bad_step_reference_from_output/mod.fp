mod "test" {

}

pipeline "bad_step_ref" {

  param "prompt" {
    default = "Do you approve?"
  }

  param "subject" {
    default = "Do you approve?"
  }

  param "notifiers" {
    default = ["default"]
  }

  output "approved" {
    value = alltrue([for decision in step.input.approve : decision.value == "approve"])
  }
}