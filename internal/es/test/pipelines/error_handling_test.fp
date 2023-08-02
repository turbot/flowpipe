
pipeline "bad_http_not_ignored" {
    description = "my simple http pipeline"
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.jsons"
    }
    step "echo" "bad_http" {
        depends_on = [step.http.my_step_1]
        text = "foo"
    }
}


pipeline "bad_http_ignored" {
    description = "my simple http pipeline"
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
}

pipeline "bad_http_ignored_get_error_code" {
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.jsons"
        error {
            ignore = true
        }
    }
    step "echo" "bad_http" {
        text = step.http.my_step_1.status_code
    }

    output "one" {
        value = step.echo.bad_http.text
    }
}

pipeline "bad_http_ignored_for_each" {
    step "http" "my_step_1" {
        url = "http://api.open-notify.org/astros.jsons"
        error {
            ignore = true
        }
    }
    step "echo" "bad_http" {
        for_each = step.http.my_step_1
        text = each.value.message
    }
}


pipeline "bad_http_with_for" {
    param "files" {
        type = list(string)
        // bad.json & ugly.json = 404
        // astros.json = 200
        default = ["bad.json", "ugly.json", "astros.json"]
    }

    step "http" "http_step" {
        for_each = param.files
        url = "http://api.open-notify.org/${each.value}"
        error {
            ignore = true
        }
    }

    step "echo" "http_step" {
        for_each = step.http.http_step
        text = each.value.status_code
        if = each.value.status_code == 200
    }
}