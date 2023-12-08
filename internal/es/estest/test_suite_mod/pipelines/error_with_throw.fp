pipeline "error_with_throw_simple" {
    step "transform" "foo" {
        value = "bar"

        throw {
            if = result.value == "bar"
            message = "from throw block"
        }

        # retry does not catch throw, so there should only be 1 step execution here
        retry {
            max_attempts = 2
        }
    }
}

pipeline "error_with_throw_invalid_message" {
    step "transform" "foo" {
        value = "bar"
    }

    step "http" "bad_http" {
        url    = "http://api.google.com/bad.json"
        method = "get"

        throw {
            if      = step.transform.foo.value == "should not match"
            message = result.response_body.error
        }
    }
}

pipeline "error_with_throw_failing_to_calculate_throw" {
    step "transform" "foo" {
        value = "bar"
    }

    step "http" "bad_http" {
        // this is an error .. that normally would be ignored ... but
        url    = "http://api.google.com/bad.json"
        method = "get"

        // This will be ignored, throw block evaluation error does not respect ignore = true error directive
        error {
            ignore = true
        }

        throw {
            if      = step.transform.foo.value == "bar"
            message = result.response_body.error
        }
    }
}

pipeline "error_with_throw_does_not_ignore" {
    step "transform" "foo" {
        value = "bar"
    }

    step "transform" "good_step" {
        value = "baz"

        // This will be ignored, throw block evaluation error does not respect ignore = true error directive
        error {
            ignore = true
        }

        throw {
            if      = step.transform.foo.value == "bar"
            message = result.response_body.error
        }
    }
}

pipeline "error_with_throw_does_not_retry" {
    step "transform" "foo" {
        value = "bar"
    }

    step "transform" "good_step" {
        value = "baz"

        // This will be ignored, throw block evaluation error does not respect retry
        retry {
            max_attempts = 2
        }

        throw {
            if      = step.transform.foo.value == "bar"
            message = result.response_body.error
        }
    }
}


pipeline "error_with_throw_but_ignored" {
    step "transform" "foo" {
        value = "bar"

        throw {
            if = result.value == "bar"
            message = "from throw block"
        }

        error {
            ignore = true
        }
    }
}


pipeline "error_with_multiple_throws" {
    step "transform" "foo" {
        value = "bar"

        throw {
            if = result.value == "baz"
            message = "from throw block baz"
        }

        throw {
            if = result.value == "bar"
            message = "from throw block bar"
        }

        # retry does not catch throw, so there should only be 1 step execution here
        retry {
            max_attempts = 2
        }
    }
}


pipeline "error_with_throw_simple_nested_pipeline" {
    step "pipeline" "foo" {

        pipeline = pipeline.nested_for_throw

        throw {
            if = result.output.val == "bar"
            message = "from throw block"
        }

        retry {
            max_attempts = 2
        }
    }
}

pipeline "nested_for_throw" {

    output "val" {
        value = "bar"
    }
}

pipeline "error_with_retries_that_works" {
    step "transform" "foo" {
        value = "bar"

        throw {
            if = retry.count < 2
            message = "from throw block"
        }

        retry {
            max_attempts = 4
        }
    }
}