
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
    step "echo" "bad_http" {
        depends_on = [step.http.my_step_1]
        text = "foo"
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
        // astros.json = 201
        default = ["bad.json", "ugly.json", "astros.json"]
    }

    step "http" "http_step" {
        for_each = param.files
        url = "http://api.open-notify.org/${each.value}"
        error {
            ignore = true
        }
    }

    step "echo" "bad_http" {
        for_each = step.http.http_step.errors
        text = each.value.message
    }
}