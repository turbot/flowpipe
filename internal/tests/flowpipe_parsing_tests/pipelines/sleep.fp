pipeline "pipeline_with_sleep" {

  step "sleep" "sleep_duration_string_input" {
    duration = "5s"
  }

  step "sleep" "sleep_duration_integer_input" {
    duration = 2000
  }
}

pipeline "pipeline_with_sleep_unresolved" {

  param "duration_string" {
    type    = string
    default = "3s"
  }

  param "duration_integer" {
    type    = number
    default = 3000
  }

  step "sleep" "sleep_duration_string_input" {
    duration = param.duration_string
  }

  step "sleep" "sleep_duration_integer_input" {
    duration = param.duration_integer
  }
}
