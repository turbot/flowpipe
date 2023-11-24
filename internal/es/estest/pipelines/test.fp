pipeline "step_output_2" {

  step "transform" "start_step" {
      value = "foo"

      # output "test_output" {
      #    value = "foo1"
      # }

      # output "start_output" {
      #    value = "bar"
      # }
  }

  step "transform" "end_step" {
   #   text = step.echo.start_step.output.start_output
     value = step.transform.start_step.value
  }
}