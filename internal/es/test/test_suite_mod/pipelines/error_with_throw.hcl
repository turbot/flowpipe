pipeline "error_with_throw_simple" {
    step "transform" "foo" {
        value = "bar"

        throw {
            if = result.value == "bar"
            message = "from throw block"
        }

        retry {
            retries = 2
        }
    }
}

pipeline "error_with_throw_and_recover" {
    step "transform" "foo" {
        value = "loop: ${loop.index}"

        loop {
            until = loop.index < 2
            value = "loop: ${loop.index}"
        }
    }

    output "val" {
        value = step.transform.foo
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