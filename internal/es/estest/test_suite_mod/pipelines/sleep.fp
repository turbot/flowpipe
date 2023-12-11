pipeline "pipeline_with_sleep_step" {

  step "sleep" "sleep_test" {
    duration = "4s"
  }

  output "sleep_duration" {
    value = step.sleep.sleep_test.duration
  }
}

pipeline "pipeline_with_sleep_step_int_duration" {

  step "sleep" "sleep_test" {
    duration = 4000
  }

  output "sleep_duration" {
    value = step.sleep.sleep_test.duration
  }
}