pipeline "bad_http_retries" {
    description = "Bad HTTP step with retries. Retry is not working at the moment, but it's parsed correctly"
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.jsons"

        error {
          ignore = true
          retries = 2
        }
    }

    step "echo" "bad_http" {
        for_each = step.http.my_step_1.errors
        text = each.message
    }
}
