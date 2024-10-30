pipeline "invalid_error_attribute" {
  description = "my simple http pipeline"
  step "http" "my_step_1" {
    url = "http://api.open-notify.org/astros.jsons"

    error {
      ignored = true
    }
  }

  step "transform" "bad_http" {
    for_each = step.http.my_step_1.errors
    value    = each.message
  }
}

