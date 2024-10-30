mod "throw_config" {
  title = "my_mod"
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

        // message has reference to "result", ensure it does not error out
        throw {
            if      = step.transform.foo.value == "bar"
            message = result.response_body.error
        }
    }
}