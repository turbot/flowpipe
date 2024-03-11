
pipeline "parent_invalid_param" {

    step "pipeline" "call_nested" {
        pipeline = pipeline.nested_invalid_param
        args = {
            one = "foo bar"
            credentials = 11
        }
    }

    output "val" {
        value = step.pipeline.call_nested
    }
}


pipeline "nested_invalid_param" {

    param "one" {
        type = string
    }

    param "cred" {
        type = number
    }

    output "val" {
        value = "${param.one}  - ${param.two}"
    }
}
