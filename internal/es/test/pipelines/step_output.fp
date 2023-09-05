pipeline "step_output" {

  step "echo" "begin" {
    text = "baz"
  }

  step "echo" "start_step" {
      text = "foo"

      output "start_output" {
         value = "bar"
      }

      output "start_output_two" {
         value = step.echo.begin.text
      }
  }

  step "echo" "end_step" {
     text = step.echo.start_step.output.start_output_two
  }

  output "end_output" {
     value = step.echo.start_step.output.start_output_two
  }
}