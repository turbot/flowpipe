pipeline "test_param_optional" {
    param "test_optional" {
        type = number
        optional = true
    }

    step "transform" "echo_optional" {
        if     = param.test_optional != null
        value  = "optional but passed: ${param.test_optional}"
    }

    step "transform" "echo_optional_1" {
        if     = param.test_optional == null
        value  = "optional and null"
    }

    step "transform" "echo_optional_2" {
        value = param.test_optional == null ? "IS_NULL" : "NOT_NULL"
    }

    output "test_output_1" {
        value = try(step.transform.echo_optional.value, "")
    }

    output "test_output_2" {
        value = step.transform.echo_optional_1.value
    }

    output "test_output_3" {
        value = step.transform.echo_optional_2.value
    }
}