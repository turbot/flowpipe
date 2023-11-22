pipeline "bad_http_ignored" {
    description = "Ignored bad HTTP step."
    step "http" "my_step_1" {
        url = "http://google.com/astros.jsons"
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

pipeline "one_error" {
    step "http" "bad_http" {
        url = "http://api.google.com/astros.jsons"
    }
}

pipeline "error_retry" {
    step "http" "bad_http" {
        url = "http://api.google.com/astros.jsons"

        retry {
            max_attempts = 3
        }
    }
}

pipeline "error_retry_with_backoff" {
    step "http" "bad_http" {
        url = "http://api.google.com/astros.jsons"

        retry {
            max_attempts = 3
            min_interval = 2000
        }
    }
}

pipeline "error_retry_with_linear_backoff" {
    step "http" "bad_http" {
        url = "http://api.google.com/astros.jsons"

        retry {
            max_attempts = 5
            strategy = "linear"
            min_interval = 100
            max_interval = 10000
        }
    }
}

pipeline "error_retry_with_exponential_backoff" {
    step "http" "bad_http" {
        url = "http://api.google.com/astros.jsons"

        retry {
            max_attempts = 5
            strategy = "exponential"
            min_interval = 100
            max_interval = 10000
        }
    }
}


pipeline "error_retry_with_backoff_linear" {
    step "http" "bad_http" {
        url = "http://api.google.com/astros.jsons"

        retry {
            max_attempts = 3
            strategy = "linear"
            min_interval = 1000
        }
    }
}



pipeline "error_in_for_each" {

    step "http" "bad_http" {
        for_each = ["bad_1.json", "bad_2.json", "bad_3.json"]
        url = "http://api.google.com/${each.value}"
    }

    output "val" {
        value = step.http.bad_http
    }
}

pipeline "error_in_for_each_nested_pipeline" {

    step "pipeline" "http" {
        for_each = ["bad_1.json", "bad_2.json", "bad_3.json"]
        pipeline = pipeline.nested_with_http
        args = {
            file = each.value
        }
    }

    output "val" {
        value = step.pipeline.http
    }
}


pipeline "nested_with_http" {

    param "file" {
        type = string
        default = "bad.json"
    }

    step "http" "http" {
        url = "http://api.open-notify.org/${param.file}"
    }

    output "val" {
        value = step.http.http
    }
}

pipeline "bad_http_pipeline" {
    step "http" "http" {
        url = "http://google.com/bad.json"
    }

    output "val" {
        value = step.http.http
    }
}

pipeline "error_in_for_each_nested_pipeline_one_works" {

    step "pipeline" "http" {
        for_each = ["bad_1.json", "astros.json", "bad_3.json"]
        pipeline = pipeline.nested_with_http
        args = {
            file = each.value
        }
    }

    output "val" {
        value = step.pipeline.http
    }
}

pipeline "error_in_for_each_nested_pipeline_one_works_error_ignored" {

    step "pipeline" "http" {
        for_each = ["bad_1.json", "astros.json", "bad_3.json"]
        pipeline = pipeline.nested_with_http
        args = {
            file = each.value
        }

        error {
            ignore = true
        }
    }

    output "val" {
        value = step.pipeline.http
    }
}

pipeline "error_retry_with_nested_pipeline" {

    step "pipeline" "http" {
        pipeline = pipeline.bad_http_pipeline

        retry {
            max_attempts = 4
            min_interval = 1000
        }
    }
}