pipeline "error_with_throw_simple" {
    step "transform" "foo" {
        value = "bar"

        throw {
            if = result.value == "bar"
            message = "from throw block"
        }

        # retry does not catch throw, so there should only be 1 step execution here
        retry {
            retries = 2
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
            ignored = true
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
            retries = 2
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
            retries = 2
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
            if = retry.index < 2
            message = "from throw block"
        }

        retry {
            retries = 4
        }
    }
}
