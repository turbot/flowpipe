pipeline "bad_http_ignored" {
    description = "Ignored bad HTTP step."
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.jsons"
        error {
            ignore = true
        }
    }

    step "echo" "bad_http_if_error_true" {
        text = "bar"
        if = is_error(step.http.my_step_1)
    }

    step "echo" "bad_http_if_error_false" {
        text = "baz"
        if = !is_error(step.http.my_step_1)
    }

    step "echo" "error_message" {
        text = error_message(step.http.my_step_1)
    }

    step "echo" "bad_http" {
        depends_on = [step.http.my_step_1]
        text = "foo"
    }

    output "one" {
        value = step.echo.bad_http.text
    }

    output "bad_http_if_error_false" {
        value = step.echo.bad_http_if_error_false
    }

    output "bad_http_if_error_true" {
        value = step.echo.bad_http_if_error_true
    }
}


pipeline "error_retry_throw" {

    step "http" "bad_http" {
        url = "http://api.open-notify.org/astros.jsons"

        retry {
            retries = 2
        }
    }
}
