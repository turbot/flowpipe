pipeline "step_output" {

  step "echo" "start_step" {
      text = "foo"

      output "start_output" {
         value = "bar"
      }
  }

  step "echo" "end_step" {
     text = step.echo.start_step.output.start_output
  }
}