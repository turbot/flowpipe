pipeline "step_output_2" {

  step "echo" "start_step" {
      text = "foo"

      # output "test_output" {
      #    value = "foo1"
      # }

      # output "start_output" {
      #    value = "bar"
      # }
  }

  step "echo" "end_step" {
   #   text = step.echo.start_step.output.start_output
     text = step.echo.start_step.text
  }
}