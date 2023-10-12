pipeline "test_param_optional" {
    param "test_optional" {
        type = number
        optional = true
    }

    step "echo" "echo_optional" {
        if     = param.test_optional != null
        text = "optional but passed: ${param.test_optional}"
    }

    step "echo" "echo_optional_1" {
        if     = param.test_optional == null
        text = "optional and null"
    }

    step "echo" "echo_optional_2" {
        text = param.test_optional == null ? "IS_NULL" : "NOT_NULL"
    }

    output "test_output_1" {
        value = step.echo.echo_optional.text
    }

    output "test_output_2" {
        value = step.echo.echo_optional_1.text
    }

    output "test_output_3" {
        value = step.echo.echo_optional_2.text
    }
}