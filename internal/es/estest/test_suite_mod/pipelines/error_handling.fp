pipeline "bad_http_ignored" {
    description = "Ignored bad HTTP step."
    step "http" "my_step_1" {
        url = "http://google.com/astros.jsons"
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

    output "bad_http_if_error_false" {
        value = step.transform.bad_http_if_error_false
    }

    output "bad_http_if_error_true" {
        value = step.transform.bad_http_if_error_true
    }
}

pipeline "one_error" {
    step "http" "bad_http" {
        url = "http://api.google.com/astros.jsons"
    }
}

pipeline "error_retry_with_if" {
    step "http" "bad_http" {
        url = "http://api.google.com/astros.jsons"

        retry {
            if = result.status_code == 404
            max_attempts = 3
        }
    }
}

pipeline "error_retry_with_if_multi_step" {
    step "http" "bad_http" {
        url = "http://api.google.com/astros.jsons"

        retry {
            if = result.status_code == 404
            max_attempts = 3
        }
    }

    step "http" "bad_http_2" {
        url = "http://api.google.com/astros.jsons"
    }

    step "http" "good_http" {
        url = "https://google.com"
    }
}

pipeline "error_retry_with_if_not_match" {
    step "http" "bad_http" {
        url = "http://api.google.com/astros.jsons"

        retry {
            if = result.status_code == 405
            max_attempts = 3
        }
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

pipeline "error_retry_failed_calculating_output_block" {

    step "transform" "one" {
        value = "foo"

        // This output block will fail to calculate
        output "val" {
            value = step.transform.one.xyz
        }

        // This retry directive should be ignored, output block calculation error
        // ignore retry directive
        retry {
            max_attempts = 3
            min_interval = 1000
        }
    }
}

pipeline "error_retry_failed_calculating_output_block_ignored_error_should_not_be_followed" {

    step "transform" "one" {
        value = "foo"

        // Because of the failure is in the output block, the ignore error directive should not be followed
        error {
            ignore = true
        }

        // This output block will fail to calculate
        output "val" {
            value = step.transform.one.xyz
        }

        // This retry directive should be ignored, output block calculation error
        // ignore retry directive
        retry {
            max_attempts = 3
            min_interval = 1000
        }
    }

    // This step should not be executed
    step "transform" "two" {
        depends_on = [step.transform.one]
        value = "If you see this, it means there's a bug. This step shouldn't be executed."
    }

    // This output should NOT be calculated
    output "val" {
        value = step.transform.two.value
    }
}

pipeline "loop_block_evaluation_error" {

    // this step has ignore = true, loop block failed to render so ignore error is not followed
    step "transform" "one" {
        value = "bar ${loop.index}"

        error {
            ignore = true
        }

        retry {
            max_attempts = 5
        }

        loop {
            until = loop.index > 1
            value = result.xyz
        }
    }

    step "transform" "two" {
        depends_on = [step.transform.one]
        value = "should not exist"
    }

    output "val" {
        value = step.transform.one.value
    }

    output "val_two" {
        value = step.transform.two.value
    }
}