pipeline "step_output" {

  step "transform" "begin" {
    value = "baz"
  }

  step "transform" "start_step" {
      value = "foo"

      output "start_output" {
         value = "bar"
      }

      output "start_output_two" {
         value = step.transform.begin.value
      }
  }

  step "transform" "end_step" {
     value = step.transform.start_step.output.start_output_two
  }

  output "end_output" {
     value = step.transform.start_step.output.start_output_two
  }
}