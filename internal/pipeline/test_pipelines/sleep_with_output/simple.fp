
pipeline "sleep_with_output" {
    step "sleep" "sleep_1" {
        duration = "1s"
    }

    output "sleep_duration" {
      value = step.sleep.sleep_1.duration
    }
}
