
pipeline "bad_http_not_ignored" {
    description = "Pipeline with a HTTP step that will fail. Error is not ignored."
    step "http" "my_step_1" {
        url = "http://localhost:7104/astros.jsons"
    }

    step "transform" "bad_http" {
        depends_on = [step.http.my_step_1]
        value      = "foo"
    }
}

pipeline "bad_http_ignored_one_step" {
    description = "A simple pipeline with a single bad HTTP step that is ignored."
    step "http" "my_step_1" {
        url = "http://localhost:7104/astros.jsons"

        error {
            ignore = true
        }
    }
}

pipeline "bad_http_ignored_two_steps" {
    description = "Bad HTTP step with an echo step. Bad HTTP step error is ignored."
    step "http" "my_step_1" {
        url = "http://localhost:7104/astros.jsons"

        error {
            ignore = true
        }
    }

    step "transform" "text_1" {
        depends_on = [step.http.my_step_1]
        value      = "foo"
    }
}


pipeline "bad_http_one_step" {
    description = "Bad HTTP step, just one step in the pipeline."

    step "http" "my_step_1" {
        # should return 404
        url = "http://localhost:7104/astros.jsons"
    }
}


pipeline "bad_http_ignored" {
    description = "Ignored bad HTTP step."
    step "http" "my_step_1" {
        url = "http://localhost:7104/astros.jsons"
        error {
            ignore = true
        }
    }

    step "transform" "bad_http_if_error_true" {
        value = "bar"
        if    = is_error(step.http.my_step_1)
    }

    step "transform" "bad_http_if_error_false" {
        value = "baz"
        if    = !is_error(step.http.my_step_1)
    }

    step "transform" "error_message" {
        value = error_message(step.http.my_step_1)
    }

    step "transform" "bad_http" {
        depends_on = [step.http.my_step_1]
        value      = "foo"
    }

    output "one" {
        value = step.transform.bad_http.value
    }
}

pipeline "bad_http_ignored_get_error_code" {
    step "http" "my_step_1" {
        url = "http://localhost:7104/astros.jsons"
        error {
            ignore = true
        }
    }
    step "transform" "bad_http" {
        value = step.http.my_step_1.status_code
    }

    output "one" {
        value = step.transform.bad_http.value
    }
}

pipeline "bad_http_ignored_for_each" {
    step "http" "my_step_1" {
        url = "http://localhost:7104/astros.jsons"
        error {
            ignore = true
        }
    }
    step "transform" "bad_http" {
        for_each = step.http.my_step_1
        value    = each.value.message
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
        url = "http://localhost:7104/${each.value}"
        error {
            ignore = true
        }
    }

    step "transform" "http_step" {
        for_each = step.http.http_step
        value    = each.value.status_code
        if       = each.value.status_code == 200
    }
}