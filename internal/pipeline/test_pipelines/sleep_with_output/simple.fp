
pipeline "sleep_with_output" {
    step "sleep" "sleep_1" {
        duration = 1
    }

    output "sleep_duration" {
      value = step.sleep.sleep_1.duration
    }
}
