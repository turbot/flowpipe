pipeline "bad_http_ignored_with_if" {
  description = "my simple http pipeline"
  step "http" "my_step_1" {
    url = "http://api.open-notify.org/astros.jsons"

    error {
      if = result.status_code == 404
      ignore = true
    }
  } 
}

pipeline "bad_http" {
  description = "my simple http pipeline"
  step "http" "my_step_1" {
    url = "http://api.open-notify.org/astros.jsons"

    error {
      ignore = true
    }
  }

  step "transform" "bad_http" {
    for_each = step.http.my_step_1.errors
    value    = each.message
  }
}

pipeline "bad_http_retries" {
  description = "Bad HTTP step with retries. Retry is not working at the moment, but it's parsed correctly"
  step "http" "my_step_1" {
    url = "http://api.open-notify.org/astros.jsons"

    error {
      ignore = true
    }
  }

  step "transform" "bad_http" {
    for_each = step.http.my_step_1.errors
    value    = each.message
  }
}
