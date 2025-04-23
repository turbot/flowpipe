pipeline "step_output" {

  step "transform" "start_step" {
    value = "foo"

    output "start_output" {
      value = "bar"
    }
  }

  step "transform" "end_step" {
    value = step.transform.start_step.output.start_output
  }
}
