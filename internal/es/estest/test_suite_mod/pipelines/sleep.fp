pipeline "pipeline_with_sleep_step_int_duration" {

  step "sleep" "sleep_test" {
    duration = 100
  }

  output "sleep_duration" {
    value = step.sleep.sleep_test.duration
  }
}