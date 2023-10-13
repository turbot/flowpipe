pipeline "param_on_echo" {
    param "test_optional" {
        type = string
        # optional = true
        # default = null
    }

    step "echo" "echo_one" {
        text = "Hello World"
    }

    step "echo" "test" {
        text = param.test_optional != null ? "${param.test_optional}" : "default"
    }

    output "echo_one_output" {
        value = step.echo.test.text
    }
}